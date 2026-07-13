package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umran/new.crm/backend/internal/followup/domain"
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
		AutoMigrate(&FollowUpModel{}, &FollowUpImageModel{}, &MeetPersonModel{})
}

func (r *Repository) ListFollowUps(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
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
	if err := r.followUpListBaseQuery(ctx, query).
		Distinct("tasks_follow_ups.id").
		Count(&total).Error; err != nil {
		return domain.ListResult{}, err
	}

	var rows []followUpListRow
	if err := r.followUpListBaseQuery(ctx, query).
		Select(`
			tasks_follow_ups.uuid,
			tasks_customers.uuid AS tasks_customer_uuid,
			tasks.uuid AS task_uuid,
			tasks.title,
			customers.id AS customer_id,
			customers.unvan AS customer_unvan,
			tasks.assigned_user_full_name,
			tasks.branch_name,
			tasks_follow_ups.visit_date,
			tasks_follow_ups.next_visit_date,
			tasks_follow_ups.agreement_reached
		`).
		Order(followUpListOrder(query)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		Scan(&rows).Error; err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.FollowUpListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.FollowUpListItem{
			UUID:                 row.UUID,
			TasksCustomerUUID:    row.TasksCustomerUUID,
			TaskUUID:             row.TaskUUID,
			Title:                row.Title,
			CustomerID:           row.CustomerID,
			CustomerUnvan:        stringValue(row.CustomerUnvan),
			AssignedUserFullName: row.AssignedUserFullName,
			BranchName:           row.BranchName,
			VisitDate:            formatDate(row.VisitDate),
			NextVisitDate:        formatDate(row.NextVisitDate),
			AgreementReached:     row.AgreementReached,
		})
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

func (r *Repository) FindTaskCustomerByUUID(ctx context.Context, taskCustomerUUID string) (domain.TaskCustomer, error) {
	if r == nil || r.db == nil || strings.TrimSpace(taskCustomerUUID) == "" {
		return domain.TaskCustomer{}, gorm.ErrInvalidDB
	}

	var taskCustomer taskCustomerRow
	if err := r.db.WithContext(ctx).
		Model(&TaskCustomerModel{}).
		Select("tasks_customers.id, tasks_customers.uuid, tasks_customers.status, tasks.assigned_user_id").
		Joins("JOIN tasks ON tasks.id = tasks_customers.task_id AND tasks.deleted_at IS NULL").
		Where("tasks_customers.uuid = ?", strings.TrimSpace(taskCustomerUUID)).
		Scan(&taskCustomer).Error; err != nil {
		return domain.TaskCustomer{}, err
	}
	if taskCustomer.ID == 0 {
		return domain.TaskCustomer{}, gorm.ErrRecordNotFound
	}

	return domain.TaskCustomer{
		ID:             taskCustomer.ID,
		UUID:           taskCustomer.UUID,
		Status:         taskCustomer.Status,
		AssignedUserID: taskCustomer.AssignedUserID,
	}, nil
}

func (r *Repository) FindFollowUpUpdateTargetByUUID(ctx context.Context, followUpUUID string) (domain.FollowUpUpdateTarget, error) {
	if r == nil || r.db == nil || strings.TrimSpace(followUpUUID) == "" {
		return domain.FollowUpUpdateTarget{}, gorm.ErrInvalidDB
	}

	var row followUpUpdateTargetRow
	if err := r.db.WithContext(ctx).
		Model(&FollowUpModel{}).
		Select("id, uuid, tasks_customer_id, visit_date").
		Where("uuid = ?", strings.TrimSpace(followUpUUID)).
		Where("deleted_at IS NULL").
		Scan(&row).Error; err != nil {
		return domain.FollowUpUpdateTarget{}, err
	}
	if row.ID == 0 {
		return domain.FollowUpUpdateTarget{}, gorm.ErrRecordNotFound
	}

	return domain.FollowUpUpdateTarget{
		ID:              row.ID,
		UUID:            row.UUID,
		TasksCustomerID: row.TasksCustomerID,
		VisitDate:       formatDate(row.VisitDate),
	}, nil
}

func (r *Repository) GetFollowUp(ctx context.Context, followUpUUID string) (domain.FollowUp, error) {
	if r == nil || r.db == nil || strings.TrimSpace(followUpUUID) == "" {
		return domain.FollowUp{}, gorm.ErrInvalidDB
	}

	var row followUpDetailRow
	if err := r.db.WithContext(ctx).
		Model(&FollowUpModel{}).
		Select(`
			tasks_follow_ups.id,
			tasks_follow_ups.uuid,
			tasks_customers.uuid AS tasks_customer_uuid,
			tasks.uuid AS task_uuid,
			tasks.title,
			customers.id AS customer_id,
			customers.unvan AS customer_unvan,
			tasks.assigned_user_full_name,
			tasks.branch_name,
			tasks_follow_ups.visit_type,
			tasks_follow_ups.visit_date,
			tasks_follow_ups.next_visit_date,
			tasks_follow_ups.agreement_reached,
			tasks_follow_ups.agreement_failure_reason,
			tasks_follow_ups.note
		`).
		Joins("JOIN tasks_customers ON tasks_customers.id = tasks_follow_ups.tasks_customer_id").
		Joins("JOIN tasks ON tasks.id = tasks_customers.task_id AND tasks.deleted_at IS NULL").
		Joins("JOIN customers ON customers.id = tasks_customers.customer_id AND customers.deleted_at IS NULL").
		Where("tasks_follow_ups.uuid = ?", strings.TrimSpace(followUpUUID)).
		Where("tasks_follow_ups.deleted_at IS NULL").
		Scan(&row).Error; err != nil {
		return domain.FollowUp{}, err
	}
	if row.ID == 0 {
		return domain.FollowUp{}, gorm.ErrRecordNotFound
	}

	images, err := r.followUpImages(ctx, row.ID)
	if err != nil {
		return domain.FollowUp{}, err
	}

	meetPeople, err := r.followUpMeetPeople(ctx, row.ID)
	if err != nil {
		return domain.FollowUp{}, err
	}

	return domain.FollowUp{
		UUID:                   row.UUID,
		TasksCustomerUUID:      row.TasksCustomerUUID,
		TaskUUID:               row.TaskUUID,
		Title:                  row.Title,
		CustomerID:             row.CustomerID,
		CustomerUnvan:          stringValue(row.CustomerUnvan),
		AssignedUserFullName:   row.AssignedUserFullName,
		BranchName:             row.BranchName,
		VisitType:              row.VisitType,
		VisitDate:              formatDate(row.VisitDate),
		NextVisitDate:          formatDate(row.NextVisitDate),
		AgreementReached:       row.AgreementReached,
		AgreementFailureReason: stringValue(row.AgreementFailureReason),
		Note:                   stringValue(row.Note),
		Images:                 images,
		MeetPeople:             meetPeople,
	}, nil
}

type taskCustomerRow struct {
	ID             uint64
	UUID           string
	Status         string
	AssignedUserID uint64
}

type followUpUpdateTargetRow struct {
	ID              uint64
	UUID            string
	TasksCustomerID uint64
	VisitDate       *time.Time
}

type followUpListRow struct {
	UUID                 string
	TasksCustomerUUID    string
	TaskUUID             string
	Title                string
	CustomerID           uint64
	CustomerUnvan        *string
	AssignedUserFullName string
	BranchName           string
	VisitDate            *time.Time
	NextVisitDate        *time.Time
	AgreementReached     bool
}

type followUpDetailRow struct {
	ID                     uint64
	UUID                   string
	TasksCustomerUUID      string
	TaskUUID               string
	Title                  string
	CustomerID             uint64
	CustomerUnvan          *string
	AssignedUserFullName   string
	BranchName             string
	VisitType              string
	VisitDate              *time.Time
	NextVisitDate          *time.Time
	AgreementReached       bool
	AgreementFailureReason *string
	Note                   *string
}

func (r *Repository) CreateFollowUp(ctx context.Context, input domain.PersistFollowUpInput) (domain.FollowUp, error) {
	if r == nil || r.db == nil {
		return domain.FollowUp{}, gorm.ErrInvalidDB
	}

	visitDate, err := parseDateTime(input.VisitDate)
	if err != nil {
		return domain.FollowUp{}, err
	}

	var nextVisitDate *time.Time
	if strings.TrimSpace(input.NextVisitDate) != "" {
		parsedNextVisitDate, err := parseDateTime(input.NextVisitDate)
		if err != nil {
			return domain.FollowUp{}, err
		}
		nextVisitDate = &parsedNextVisitDate
	}

	followUp := FollowUpModel{
		UUID:                   input.UUID,
		TasksCustomerID:        input.TasksCustomerID,
		VisitType:              input.VisitType,
		VisitDate:              visitDate,
		NextVisitDate:          nextVisitDate,
		AgreementReached:       input.AgreementReached,
		AgreementFailureReason: stringPointer(input.AgreementFailureReason),
		Note:                   stringPointer(input.Note),
	}

	imageModels := make([]FollowUpImageModel, 0, len(input.Images))
	images := make([]domain.Image, 0, len(input.Images))
	for _, image := range input.Images {
		imageUUID := image.UUID
		if imageUUID == "" {
			imageUUID = uuid.NewString()
		}

		imageModels = append(imageModels, FollowUpImageModel{
			UUID: imageUUID,
			Path: image.Path,
			URL:  image.URL,
		})
		images = append(images, domain.Image{
			UUID: imageUUID,
			URL:  image.URL,
		})
	}

	meetPersonModels := make([]MeetPersonModel, 0, len(input.MeetPeople))
	meetPeople := make([]domain.MeetPerson, 0, len(input.MeetPeople))
	for _, person := range input.MeetPeople {
		personUUID := uuid.NewString()
		meetPersonModels = append(meetPersonModels, MeetPersonModel{
			UUID:    personUUID,
			Title:   person.Title,
			Name:    person.Name,
			Surname: person.Surname,
			Phone:   person.Phone,
			Email:   stringPointer(person.Email),
		})
		meetPeople = append(meetPeople, domain.MeetPerson{
			UUID:    personUUID,
			Title:   person.Title,
			Name:    person.Name,
			Surname: person.Surname,
			Phone:   person.Phone,
			Email:   person.Email,
		})
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&followUp).Error; err != nil {
			return err
		}

		for index := range imageModels {
			imageModels[index].TasksFollowUpID = followUp.ID
		}
		if len(imageModels) > 0 {
			if err := tx.Create(&imageModels).Error; err != nil {
				return err
			}
		}

		for index := range meetPersonModels {
			meetPersonModels[index].TasksFollowUpID = followUp.ID
		}
		if len(meetPersonModels) > 0 {
			if err := tx.Create(&meetPersonModels).Error; err != nil {
				return err
			}
		}

		nextStatus := "in_progress"
		if strings.TrimSpace(input.NextVisitDate) == "" {
			nextStatus = "completed"
		}
		if err := tx.Model(&TaskCustomerModel{}).
			Where("id = ?", input.TasksCustomerID).
			Update("status", nextStatus).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return domain.FollowUp{}, err
	}

	return domain.FollowUp{
		UUID:                   followUp.UUID,
		TasksCustomerUUID:      input.TasksCustomerUUID,
		VisitType:              input.VisitType,
		VisitDate:              input.VisitDate,
		NextVisitDate:          input.NextVisitDate,
		AgreementReached:       input.AgreementReached,
		AgreementFailureReason: input.AgreementFailureReason,
		Note:                   input.Note,
		Images:                 images,
		MeetPeople:             meetPeople,
	}, nil
}

func (r *Repository) UpdateFollowUp(ctx context.Context, input domain.PersistUpdateFollowUpInput) (domain.FollowUp, []domain.StoredImage, error) {
	if r == nil || r.db == nil {
		return domain.FollowUp{}, nil, gorm.ErrInvalidDB
	}

	var nextVisitDate *time.Time
	if strings.TrimSpace(input.NextVisitDate) != "" {
		parsedNextVisitDate, err := parseDateTime(input.NextVisitDate)
		if err != nil {
			return domain.FollowUp{}, nil, err
		}
		nextVisitDate = &parsedNextVisitDate
	}

	imageModels := make([]FollowUpImageModel, 0, len(input.Images))
	for _, image := range input.Images {
		imageUUID := image.UUID
		if imageUUID == "" {
			imageUUID = uuid.NewString()
		}
		imageModels = append(imageModels, FollowUpImageModel{
			UUID:            imageUUID,
			TasksFollowUpID: input.ID,
			Path:            image.Path,
			URL:             image.URL,
		})
	}

	meetPersonModels := make([]MeetPersonModel, 0, len(input.MeetPeople))
	for _, person := range input.MeetPeople {
		meetPersonModels = append(meetPersonModels, MeetPersonModel{
			UUID:            uuid.NewString(),
			TasksFollowUpID: input.ID,
			Title:           person.Title,
			Name:            person.Name,
			Surname:         person.Surname,
			Phone:           person.Phone,
			Email:           stringPointer(person.Email),
		})
	}

	var deletedImages []domain.StoredImage
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&FollowUpModel{}).
			Where("id = ?", input.ID).
			Updates(map[string]any{
				"visit_type":               input.VisitType,
				"next_visit_date":          nextVisitDate,
				"agreement_reached":        input.AgreementReached,
				"agreement_failure_reason": stringPointer(input.AgreementFailureReason),
				"note":                     stringPointer(input.Note),
			}).Error; err != nil {
			return err
		}

		var existingImages []FollowUpImageModel
		if err := tx.Where("tasks_follow_up_id = ?", input.ID).
			Find(&existingImages).Error; err != nil {
			return err
		}

		keepImageUUIDs := map[string]struct{}{}
		for _, imageUUID := range input.ExistingImageUUIDs {
			keepImageUUIDs[imageUUID] = struct{}{}
		}

		keptImageCount := 0
		deleteImageIDs := make([]uint64, 0)
		for _, image := range existingImages {
			if _, keep := keepImageUUIDs[image.UUID]; keep {
				keptImageCount++
				continue
			}

			deleteImageIDs = append(deleteImageIDs, image.ID)
			deletedImages = append(deletedImages, domain.StoredImage{
				UUID: image.UUID,
				Path: image.Path,
				URL:  image.URL,
			})
		}

		if keptImageCount+len(imageModels) > 3 {
			return gorm.ErrInvalidData
		}

		if len(deleteImageIDs) > 0 {
			if err := tx.Delete(&FollowUpImageModel{}, deleteImageIDs).Error; err != nil {
				return err
			}
		}

		if len(imageModels) > 0 {
			if err := tx.Create(&imageModels).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("tasks_follow_up_id = ?", input.ID).
			Delete(&MeetPersonModel{}).Error; err != nil {
			return err
		}

		if len(meetPersonModels) > 0 {
			if err := tx.Create(&meetPersonModels).Error; err != nil {
				return err
			}
		}

		nextStatus := "in_progress"
		if strings.TrimSpace(input.NextVisitDate) == "" {
			nextStatus = "completed"
		}
		if err := tx.Model(&TaskCustomerModel{}).
			Where("id = ?", input.TasksCustomerID).
			Update("status", nextStatus).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return domain.FollowUp{}, nil, err
	}

	followUp, err := r.GetFollowUp(ctx, input.UUID)
	if err != nil {
		return domain.FollowUp{}, nil, err
	}

	return followUp, deletedImages, nil
}

func (r *Repository) followUpListBaseQuery(ctx context.Context, filters domain.ListQuery) *gorm.DB {
	query := r.db.WithContext(ctx).
		Model(&FollowUpModel{}).
		Joins("JOIN tasks_customers ON tasks_customers.id = tasks_follow_ups.tasks_customer_id").
		Joins("JOIN tasks ON tasks.id = tasks_customers.task_id AND tasks.deleted_at IS NULL").
		Joins("JOIN customers ON customers.id = tasks_customers.customer_id AND customers.deleted_at IS NULL").
		Where("tasks_follow_ups.deleted_at IS NULL")

	return applyFollowUpFilters(query, filters)
}

func applyFollowUpFilters(query *gorm.DB, filters domain.ListQuery) *gorm.DB {
	if strings.TrimSpace(filters.Title) != "" {
		query = query.Where("tasks.title LIKE ?", "%"+strings.TrimSpace(filters.Title)+"%")
	}

	if strings.TrimSpace(filters.Customer) != "" {
		query = query.Where("customers.unvan LIKE ?", "%"+strings.TrimSpace(filters.Customer)+"%")
	}

	if filters.AssignedUserID > 0 {
		query = query.Where("tasks.assigned_user_id = ?", filters.AssignedUserID)
	}

	if strings.TrimSpace(filters.AssignedUserFullName) != "" {
		query = query.Where("tasks.assigned_user_full_name LIKE ?", "%"+strings.TrimSpace(filters.AssignedUserFullName)+"%")
	}

	if strings.TrimSpace(filters.BranchName) != "" {
		query = query.Where("tasks.branch_name LIKE ?", "%"+strings.TrimSpace(filters.BranchName)+"%")
	}

	if strings.TrimSpace(filters.VisitDate) != "" {
		query = query.Where("tasks_follow_ups.visit_date LIKE ?", "%"+strings.TrimSpace(filters.VisitDate)+"%")
	}

	if strings.TrimSpace(filters.NextVisitDate) != "" {
		query = query.Where("tasks_follow_ups.next_visit_date LIKE ?", "%"+strings.TrimSpace(filters.NextVisitDate)+"%")
	}

	return query
}

func followUpListOrder(query domain.ListQuery) string {
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	switch sortBy {
	case "visit_date", "next_visit_date", "agreement_reached":
	default:
		return "tasks_follow_ups.id DESC"
	}

	sortOrder := "DESC"
	if strings.ToLower(strings.TrimSpace(query.SortOrder)) == "asc" {
		sortOrder = "ASC"
	}

	return "tasks_follow_ups." + sortBy + " " + sortOrder + ", tasks_follow_ups.id DESC"
}

func (r *Repository) followUpImages(ctx context.Context, followUpID uint64) ([]domain.Image, error) {
	var rows []FollowUpImageModel
	if err := r.db.WithContext(ctx).
		Where("tasks_follow_up_id = ?", followUpID).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	images := make([]domain.Image, 0, len(rows))
	for _, row := range rows {
		images = append(images, domain.Image{
			UUID: row.UUID,
			URL:  row.URL,
		})
	}

	return images, nil
}

func (r *Repository) followUpMeetPeople(ctx context.Context, followUpID uint64) ([]domain.MeetPerson, error) {
	var rows []MeetPersonModel
	if err := r.db.WithContext(ctx).
		Where("tasks_follow_up_id = ?", followUpID).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	meetPeople := make([]domain.MeetPerson, 0, len(rows))
	for _, row := range rows {
		meetPeople = append(meetPeople, domain.MeetPerson{
			UUID:    row.UUID,
			Title:   row.Title,
			Name:    row.Name,
			Surname: row.Surname,
			Phone:   row.Phone,
			Email:   stringValue(row.Email),
		})
	}

	return meetPeople, nil
}

func parseDateTime(value string) (time.Time, error) {
	trimmedValue := strings.TrimSpace(value)
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	var parseErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, trimmedValue, time.Local)
		if err == nil {
			return parsed, nil
		}
		parseErr = err
	}

	return time.Time{}, parseErr
}

func stringPointer(value string) *string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02")
}
