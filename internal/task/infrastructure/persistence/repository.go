package persistence

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umran/new.crm/backend/internal/task/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").
		AutoMigrate(&TaskModel{}, &TaskCustomerModel{})
}

func (r *Repository) InvalidCustomerIDsForBranch(ctx context.Context, customerIDs []uint64, branchID uint64) ([]uint64, error) {
	if len(customerIDs) == 0 || branchID == 0 {
		return customerIDs, nil
	}

	var validCustomerIDs []uint64
	if err := r.db.WithContext(ctx).
		Model(&CustomerModel{}).
		Where("id IN ?", customerIDs).
		Where("branch_id = ?", branchID).
		Pluck("id", &validCustomerIDs).Error; err != nil {
		return nil, err
	}

	validCustomers := make(map[uint64]struct{}, len(validCustomerIDs))
	for _, customerID := range validCustomerIDs {
		validCustomers[customerID] = struct{}{}
	}

	invalidCustomerIDs := make([]uint64, 0)
	for _, customerID := range customerIDs {
		if _, ok := validCustomers[customerID]; !ok {
			invalidCustomerIDs = append(invalidCustomerIDs, customerID)
		}
	}

	return invalidCustomerIDs, nil
}

func (r *Repository) CreateTask(ctx context.Context, input domain.CreateTaskInput) (domain.Task, error) {
	if r == nil || r.db == nil {
		return domain.Task{}, gorm.ErrInvalidDB
	}

	task := TaskModel{
		UUID:                  uuid.NewString(),
		Title:                 input.Title,
		Description:           stringPointer(input.Description),
		AssignedUserID:        input.AssignedUserID,
		AssignedUserFullName:  input.AssignedUserFullName,
		CreatedByUserID:       input.CreatedByUserID,
		CreatedByUserFullName: input.CreatedByUserFullName,
		BranchID:              input.BranchID,
		BranchName:            input.BranchName,
		VisitDate:             datePointer(input.VisitDate),
		DueDate:               datePointer(input.DueDate),
		Status:                "pending",
		Priority:              input.Priority,
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}

		taskCustomers := make([]TaskCustomerModel, 0, len(input.CustomerIDs))
		for _, customerID := range input.CustomerIDs {
			taskCustomers = append(taskCustomers, TaskCustomerModel{
				TaskID:     task.ID,
				CustomerID: customerID,
			})
		}

		if len(taskCustomers) > 0 {
			if err := tx.Create(&taskCustomers).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return domain.Task{}, err
	}

	return toTask(task, input.CustomerIDs), nil
}

func (r *Repository) ListTasks(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if r == nil || r.db == nil {
		return domain.ListResult{}, gorm.ErrInvalidDB
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	var total int64
	if err := r.taskListBaseQuery(ctx, query).Distinct("tasks.id").Count(&total).Error; err != nil {
		return domain.ListResult{}, err
	}

	var taskIDs []uint64
	if err := r.taskListBaseQuery(ctx, query).
		Select("tasks.id").
		Group("tasks.id").
		Order(taskListOrder(query)).
		Offset((page-1)*perPage).
		Limit(perPage).
		Pluck("tasks.id", &taskIDs).Error; err != nil {
		return domain.ListResult{}, err
	}

	tasks, err := r.tasksByIDs(ctx, taskIDs)
	if err != nil {
		return domain.ListResult{}, err
	}

	customersByTaskID, err := r.customersByTaskIDs(ctx, taskIDs)
	if err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.TaskListItem, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, toTaskListItem(task, customersByTaskID[task.ID]))
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage <= 0 {
		lastPage = 1
	}

	var from *int
	var to *int
	if total > 0 {
		fromValue := ((page - 1) * perPage) + 1
		toValue := fromValue + len(items) - 1
		from = &fromValue
		to = &toValue
	}

	return domain.ListResult{
		Items: items,
		Pagination: domain.Pagination{
			CurrentPage: page,
			LastPage:    lastPage,
			PerPage:     perPage,
			Total:       int(total),
			From:        from,
			To:          to,
		},
	}, nil
}

func (r *Repository) GetTask(ctx context.Context, taskUUID string) (domain.TaskListItem, error) {
	if r == nil || r.db == nil || strings.TrimSpace(taskUUID) == "" {
		return domain.TaskListItem{}, gorm.ErrInvalidDB
	}

	var task TaskModel
	if err := r.db.WithContext(ctx).
		Where("uuid = ?", strings.TrimSpace(taskUUID)).
		Where("deleted_at IS NULL").
		First(&task).Error; err != nil {
		return domain.TaskListItem{}, err
	}

	customersByTaskID, err := r.customersByTaskIDs(ctx, []uint64{task.ID})
	if err != nil {
		return domain.TaskListItem{}, err
	}

	return toTaskListItem(task, customersByTaskID[task.ID]), nil
}

func (r *Repository) taskListBaseQuery(ctx context.Context, filters domain.ListQuery) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&TaskModel{}).Where("tasks.deleted_at IS NULL")

	if strings.TrimSpace(filters.Customer) != "" {
		query = query.
			Joins("JOIN tasks_customers ON tasks_customers.task_id = tasks.id").
			Joins("JOIN customers ON customers.id = tasks_customers.customer_id AND customers.deleted_at IS NULL")
	}

	return applyTaskFilters(query, filters)
}

func applyTaskFilters(query *gorm.DB, filters domain.ListQuery) *gorm.DB {
	if strings.TrimSpace(filters.Title) != "" {
		query = query.Where("tasks.title LIKE ?", "%"+strings.TrimSpace(filters.Title)+"%")
	}

	if strings.TrimSpace(filters.Customer) != "" {
		pattern := "%" + strings.TrimSpace(filters.Customer) + "%"
		query = query.Where("customers.unvan LIKE ? OR customers.ad LIKE ? OR customers.soyad LIKE ?", pattern, pattern, pattern)
	}

	if strings.TrimSpace(filters.AssignedUserFullName) != "" {
		query = query.Where("tasks.assigned_user_full_name LIKE ?", "%"+strings.TrimSpace(filters.AssignedUserFullName)+"%")
	}

	if strings.TrimSpace(filters.BranchName) != "" {
		query = query.Where("tasks.branch_name LIKE ?", "%"+strings.TrimSpace(filters.BranchName)+"%")
	}

	if strings.TrimSpace(filters.VisitDate) != "" {
		query = query.Where("tasks.visit_date LIKE ?", "%"+strings.TrimSpace(filters.VisitDate)+"%")
	}

	if strings.TrimSpace(filters.DueDate) != "" {
		query = query.Where("tasks.due_date LIKE ?", "%"+strings.TrimSpace(filters.DueDate)+"%")
	}

	if strings.TrimSpace(filters.Priority) != "" {
		query = query.Where("tasks.priority = ?", strings.ToLower(strings.TrimSpace(filters.Priority)))
	}

	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("tasks.status = ?", strings.ToLower(strings.TrimSpace(filters.Status)))
	}

	if strings.TrimSpace(filters.CreatedByUserFullName) != "" {
		query = query.Where("tasks.created_by_user_full_name LIKE ?", "%"+strings.TrimSpace(filters.CreatedByUserFullName)+"%")
	}

	return query
}

func taskListOrder(query domain.ListQuery) string {
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	if sortBy != "visit_date" && sortBy != "due_date" {
		return "tasks.id DESC"
	}

	sortOrder := "DESC"
	if strings.ToLower(strings.TrimSpace(query.SortOrder)) == "asc" {
		sortOrder = "ASC"
	}

	return "tasks." + sortBy + " " + sortOrder + ", tasks.id DESC"
}

func (r *Repository) tasksByIDs(ctx context.Context, taskIDs []uint64) ([]TaskModel, error) {
	if len(taskIDs) == 0 {
		return []TaskModel{}, nil
	}

	var tasks []TaskModel
	if err := r.db.WithContext(ctx).Where("id IN ?", taskIDs).Find(&tasks).Error; err != nil {
		return nil, err
	}

	taskOrder := make(map[uint64]int, len(taskIDs))
	for index, taskID := range taskIDs {
		taskOrder[taskID] = index
	}

	sort.SliceStable(tasks, func(left int, right int) bool {
		return taskOrder[tasks[left].ID] < taskOrder[tasks[right].ID]
	})

	return tasks, nil
}

type taskCustomerRow struct {
	TaskID     uint64
	CustomerID uint64
	Unvan      *string
	Ad         *string
	Soyad      *string
}

func (r *Repository) customersByTaskIDs(ctx context.Context, taskIDs []uint64) (map[uint64][]domain.TaskCustomer, error) {
	customersByTaskID := make(map[uint64][]domain.TaskCustomer, len(taskIDs))
	if len(taskIDs) == 0 {
		return customersByTaskID, nil
	}

	var rows []taskCustomerRow
	if err := r.db.WithContext(ctx).
		Model(&TaskCustomerModel{}).
		Select("tasks_customers.task_id, customers.id AS customer_id, customers.unvan, customers.ad, customers.soyad").
		Joins("JOIN customers ON customers.id = tasks_customers.customer_id AND customers.deleted_at IS NULL").
		Where("tasks_customers.task_id IN ?", taskIDs).
		Order("tasks_customers.task_id ASC, customers.unvan ASC, customers.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		customersByTaskID[row.TaskID] = append(customersByTaskID[row.TaskID], domain.TaskCustomer{
			ID:    row.CustomerID,
			Unvan: stringValue(row.Unvan),
			Ad:    stringValue(row.Ad),
			Soyad: stringValue(row.Soyad),
		})
	}

	return customersByTaskID, nil
}

func toTask(task TaskModel, customerIDs []uint64) domain.Task {
	description := ""
	if task.Description != nil {
		description = *task.Description
	}

	visitDate := ""
	if task.VisitDate != nil {
		visitDate = task.VisitDate.Format("2006-01-02")
	}

	dueDate := ""
	if task.DueDate != nil {
		dueDate = task.DueDate.Format("2006-01-02")
	}

	return domain.Task{
		ID:                    task.ID,
		UUID:                  task.UUID,
		Title:                 task.Title,
		Description:           description,
		CreatedByUserID:       task.CreatedByUserID,
		CreatedByUserFullName: task.CreatedByUserFullName,
		AssignedUserID:        task.AssignedUserID,
		AssignedUserFullName:  task.AssignedUserFullName,
		BranchID:              task.BranchID,
		BranchName:            task.BranchName,
		VisitDate:             visitDate,
		DueDate:               dueDate,
		Status:                task.Status,
		Priority:              task.Priority,
		CustomerIDs:           customerIDs,
	}
}

func toTaskListItem(task TaskModel, customers []domain.TaskCustomer) domain.TaskListItem {
	description := ""
	if task.Description != nil {
		description = *task.Description
	}

	visitDate := ""
	if task.VisitDate != nil {
		visitDate = task.VisitDate.Format("2006-01-02")
	}

	dueDate := ""
	if task.DueDate != nil {
		dueDate = task.DueDate.Format("2006-01-02")
	}

	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = "Potansiyel Müşteri"
	}

	return domain.TaskListItem{
		UUID:                  task.UUID,
		Title:                 title,
		Description:           description,
		CreatedByUserFullName: task.CreatedByUserFullName,
		AssignedUserFullName:  task.AssignedUserFullName,
		BranchName:            task.BranchName,
		VisitDate:             visitDate,
		DueDate:               dueDate,
		Status:                task.Status,
		Priority:              task.Priority,
		Customers:             customers,
	}
}

func stringPointer(value string) *string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}

func datePointer(value string) *time.Time {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	date, err := time.Parse("2006-01-02", trimmedValue)
	if err != nil {
		return nil
	}

	return &date
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
