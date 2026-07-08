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

type taskCustomerRow struct {
	ID             uint64
	UUID           string
	Status         string
	AssignedUserID uint64
}

func (r *Repository) CreateFollowUp(ctx context.Context, input domain.PersistFollowUpInput) (domain.FollowUp, error) {
	if r == nil || r.db == nil {
		return domain.FollowUp{}, gorm.ErrInvalidDB
	}

	visitDate, err := parseDateTime(input.VisitDate)
	if err != nil {
		return domain.FollowUp{}, err
	}
	nextVisitDate, err := parseDateTime(input.NextVisitDate)
	if err != nil {
		return domain.FollowUp{}, err
	}

	followUp := FollowUpModel{
		UUID:                   input.UUID,
		TasksCustomerID:        input.TasksCustomerID,
		VisitDate:              visitDate,
		NextVisitDate:          &nextVisitDate,
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

		return nil
	}); err != nil {
		return domain.FollowUp{}, err
	}

	return domain.FollowUp{
		UUID:                   followUp.UUID,
		TasksCustomerUUID:      input.TasksCustomerUUID,
		VisitDate:              input.VisitDate,
		NextVisitDate:          input.NextVisitDate,
		AgreementReached:       input.AgreementReached,
		AgreementFailureReason: input.AgreementFailureReason,
		Note:                   input.Note,
		Images:                 images,
		MeetPeople:             meetPeople,
	}, nil
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
