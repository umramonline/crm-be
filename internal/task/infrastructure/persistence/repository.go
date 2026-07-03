package persistence

import (
	"context"
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
		UUID:           task.UUID,
		Title:          task.Title,
		Description:    description,
		AssignedUserID: task.AssignedUserID,
		BranchID:       task.BranchID,
		VisitDate:      visitDate,
		DueDate:        dueDate,
		Status:         task.Status,
		Priority:       task.Priority,
		CustomerIDs:    customerIDs,
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
