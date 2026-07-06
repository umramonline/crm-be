package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/umran/new.crm/backend/internal/task/domain"
)

var ErrTaskCreateUnavailable = errors.New("task create unavailable")

var ErrTaskListUnavailable = errors.New("task list unavailable")

var ErrTaskDetailUnavailable = errors.New("task detail unavailable")

var ErrInvalidTaskCreateInput = errors.New("invalid task create input")

type ValidationErrors map[string]string

type ReferenceProvider interface {
	GetBranch(ctx context.Context, branchID uint64) (domain.Branch, error)
	GetBranchUser(ctx context.Context, branchID uint64, userID uint64) (domain.BranchUser, error)
}

type Repository interface {
	InvalidCustomerIDsForBranch(ctx context.Context, customerIDs []uint64, branchID uint64) ([]uint64, error)
	CreateTask(ctx context.Context, input domain.CreateTaskInput) (domain.Task, error)
	ListTasks(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	GetTask(ctx context.Context, uuid string) (domain.TaskListItem, error)
}

type Service struct {
	provider   ReferenceProvider
	repository Repository
}

const (
	taskTextMaxLength       = 255
	taskBranchNameMaxLength = 50
)

var taskPriorityOptions = map[string]struct{}{
	"high":   {},
	"medium": {},
	"low":    {},
}

func NewService(provider ReferenceProvider, repository Repository) *Service {
	return &Service{provider: provider, repository: repository}
}

func (s *Service) CreateTask(ctx context.Context, input domain.CreateTaskInput) (domain.Task, ValidationErrors, error) {
	normalizedInput := normalizeCreateTaskInput(input)
	validationErrors := validateCreateTaskInput(normalizedInput)
	if len(validationErrors) > 0 {
		return domain.Task{}, validationErrors, ErrInvalidTaskCreateInput
	}

	if _, err := s.provider.GetBranch(ctx, normalizedInput.BranchID); err != nil {
		return domain.Task{}, ValidationErrors{"branch_id": "Seçili bayi geçersiz."}, ErrInvalidTaskCreateInput
	}

	if _, err := s.provider.GetBranchUser(ctx, normalizedInput.BranchID, normalizedInput.AssignedUserID); err != nil {
		return domain.Task{}, ValidationErrors{"assigned_user_id": "Atanacak kullanıcı seçili bayiye ait değil."}, ErrInvalidTaskCreateInput
	}

	invalidCustomerIDs, err := s.repository.InvalidCustomerIDsForBranch(ctx, normalizedInput.CustomerIDs, normalizedInput.BranchID)
	if err != nil {
		return domain.Task{}, nil, ErrTaskCreateUnavailable
	}
	if len(invalidCustomerIDs) > 0 {
		return domain.Task{}, ValidationErrors{
			"customer_ids": "Seçilen müşterilerden bazıları seçili bayiye ait değil.",
		}, ErrInvalidTaskCreateInput
	}

	task, err := s.repository.CreateTask(ctx, normalizedInput)
	if err != nil {
		return domain.Task{}, nil, ErrTaskCreateUnavailable
	}

	return task, nil, nil
}

func (s *Service) ListTasks(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if s == nil || s.repository == nil {
		return domain.ListResult{}, ErrTaskListUnavailable
	}

	result, err := s.repository.ListTasks(ctx, normalizeListQuery(query))
	if err != nil {
		return domain.ListResult{}, ErrTaskListUnavailable
	}

	return result, nil
}

func (s *Service) GetTask(ctx context.Context, uuid string) (domain.TaskListItem, error) {
	normalizedUUID := strings.TrimSpace(uuid)
	if s == nil || s.repository == nil || normalizedUUID == "" {
		return domain.TaskListItem{}, ErrTaskDetailUnavailable
	}

	task, err := s.repository.GetTask(ctx, normalizedUUID)
	if err != nil {
		return domain.TaskListItem{}, ErrTaskDetailUnavailable
	}

	return task, nil
}

func normalizeListQuery(query domain.ListQuery) domain.ListQuery {
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	if sortBy != "visit_date" && sortBy != "due_date" {
		sortBy = ""
	}

	sortOrder := strings.ToLower(strings.TrimSpace(query.SortOrder))
	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	return domain.ListQuery{
		Page:                  query.Page,
		PerPage:               query.PerPage,
		Title:                 strings.TrimSpace(query.Title),
		Customer:              strings.TrimSpace(query.Customer),
		AssignedUserFullName:  strings.TrimSpace(query.AssignedUserFullName),
		BranchName:            strings.TrimSpace(query.BranchName),
		VisitDate:             strings.TrimSpace(query.VisitDate),
		DueDate:               strings.TrimSpace(query.DueDate),
		Priority:              strings.ToLower(strings.TrimSpace(query.Priority)),
		Status:                strings.ToLower(strings.TrimSpace(query.Status)),
		CreatedByUserFullName: strings.TrimSpace(query.CreatedByUserFullName),
		SortBy:                sortBy,
		SortOrder:             sortOrder,
	}
}

func normalizeCreateTaskInput(input domain.CreateTaskInput) domain.CreateTaskInput {
	customerIDs := make([]uint64, 0, len(input.CustomerIDs))
	seenCustomerIDs := map[uint64]struct{}{}
	for _, customerID := range input.CustomerIDs {
		if customerID == 0 {
			continue
		}

		if _, ok := seenCustomerIDs[customerID]; ok {
			continue
		}

		seenCustomerIDs[customerID] = struct{}{}
		customerIDs = append(customerIDs, customerID)
	}

	return domain.CreateTaskInput{
		Title:                 strings.TrimSpace(input.Title),
		Description:           strings.TrimSpace(input.Description),
		AssignedUserID:        input.AssignedUserID,
		AssignedUserFullName:  input.AssignedUserFullName,
		CreatedByUserID:       input.CreatedByUserID,
		CreatedByUserFullName: input.CreatedByUserFullName,
		BranchID:              input.BranchID,
		BranchName:            strings.TrimSpace(input.BranchName),
		VisitDate:             strings.TrimSpace(input.VisitDate),
		DueDate:               strings.TrimSpace(input.DueDate),
		Priority:              strings.ToLower(strings.TrimSpace(input.Priority)),
		CustomerIDs:           customerIDs,
	}
}

func validateCreateTaskInput(input domain.CreateTaskInput) ValidationErrors {
	errors := ValidationErrors{}

	requireField(errors, "title", input.Title, "Başlık zorunludur.")
	requireField(errors, "assigned_user_full_name", input.AssignedUserFullName, "Atanacak kullanıcı adı zorunludur.")
	requireField(errors, "branch_name", input.BranchName, "Bayi adı zorunludur.")
	validateMaxLength(errors, "title", input.Title, "Başlık")
	validateMaxLength(errors, "description", input.Description, "Açıklama")
	validateMaxLength(errors, "assigned_user_full_name", input.AssignedUserFullName, "Atanacak kullanıcı adı")
	validateMaxLengthLimit(errors, "branch_name", input.BranchName, "Bayi adı", taskBranchNameMaxLength)

	if input.AssignedUserID == 0 {
		errors["assigned_user_id"] = "Atanacak kullanıcı zorunludur."
	}

	if input.BranchID == 0 {
		errors["branch_id"] = "Bayi zorunludur."
	}

	validateDate(errors, "visit_date", input.VisitDate)
	validateDate(errors, "due_date", input.DueDate)

	visitDate, visitDateErr := parseOptionalDate(input.VisitDate)
	dueDate, dueDateErr := parseOptionalDate(input.DueDate)
	if visitDateErr == nil && dueDateErr == nil && visitDate != nil && dueDate != nil && dueDate.Before(*visitDate) {
		errors["due_date"] = "Bitiş tarihi ziyaret tarihinden küçük olamaz."
	}

	if _, ok := taskPriorityOptions[input.Priority]; !ok {
		errors["priority"] = "Öncelik high, medium veya low olmalıdır."
	}

	if len(input.CustomerIDs) == 0 {
		errors["customer_ids"] = "En az 1 müşteri seçilmelidir."
	}

	return errors
}

func requireField(errors ValidationErrors, field string, value string, message string) {
	if strings.TrimSpace(value) == "" {
		errors[field] = message
	}
}

func validateMaxLength(errors ValidationErrors, field string, value string, label string) {
	validateMaxLengthLimit(errors, field, value, label, taskTextMaxLength)
}

func validateMaxLengthLimit(errors ValidationErrors, field string, value string, label string, limit int) {
	if len([]rune(strings.TrimSpace(value))) > limit {
		errors[field] = label + " en fazla " + strconv.Itoa(limit) + " karakter olabilir."
	}
}

func validateDate(errors ValidationErrors, field string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	if _, err := time.Parse("2006-01-02", value); err != nil {
		errors[field] = "Tarih YYYY-AA-GG formatında olmalıdır."
	}
}

func parseOptionalDate(value string) (*time.Time, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	parsedDate, err := time.Parse("2006-01-02", trimmedValue)
	if err != nil {
		return nil, err
	}

	return &parsedDate, nil
}
