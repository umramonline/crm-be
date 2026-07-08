package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umran/new.crm/backend/internal/followup/domain"
)

type fakeRepository struct {
	taskCustomer domain.TaskCustomer
	findErr      error
	createErr    error
	createInput  domain.PersistFollowUpInput
}

func (f *fakeRepository) FindTaskCustomerByUUID(_ context.Context, _ string) (domain.TaskCustomer, error) {
	if f.findErr != nil {
		return domain.TaskCustomer{}, f.findErr
	}

	return f.taskCustomer, nil
}

func (f *fakeRepository) CreateFollowUp(_ context.Context, input domain.PersistFollowUpInput) (domain.FollowUp, error) {
	f.createInput = input
	if f.createErr != nil {
		return domain.FollowUp{}, f.createErr
	}

	return domain.FollowUp{UUID: input.UUID, TasksCustomerUUID: input.TasksCustomerUUID}, nil
}

type fakeStorage struct {
	saveErr       error
	savedImages   []domain.StoredImage
	deleteInvoked bool
}

func (f *fakeStorage) SaveFollowUpImages(_ context.Context, _ string, _ []domain.ImageUpload) ([]domain.StoredImage, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}

	return f.savedImages, nil
}

func (f *fakeStorage) DeleteImages(_ context.Context, _ []domain.StoredImage) error {
	f.deleteInvoked = true

	return nil
}

func TestCreateFollowUpRejectsAgreementFailureReasonWhenAgreementReached(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeStorage{})
	agreementReached := true

	_, validationErrors, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput(func(input *domain.CreateFollowUpInput) {
		input.AgreementReached = &agreementReached
		input.AgreementFailureReason = "Fiyat yüksek"
	}))

	if !errors.Is(err, ErrInvalidFollowUpCreateInput) {
		t.Fatalf("expected ErrInvalidFollowUpCreateInput, got %v", err)
	}
	if validationErrors["agreement_failure_reason"] == "" {
		t.Fatalf("expected agreement_failure_reason validation error, got %#v", validationErrors)
	}
}

func TestCreateFollowUpRejectsInvalidTaskCustomerStatus(t *testing.T) {
	repository := &fakeRepository{
		taskCustomer: domain.TaskCustomer{ID: 10, UUID: "task-customer-uuid", Status: "completed", AssignedUserID: 20},
	}
	service := NewService(repository, &fakeStorage{})

	_, validationErrors, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput())

	if !errors.Is(err, ErrInvalidFollowUpCreateInput) {
		t.Fatalf("expected ErrInvalidFollowUpCreateInput, got %v", err)
	}
	if validationErrors["tasks_customer_uuid"] == "" {
		t.Fatalf("expected tasks_customer_uuid validation error, got %#v", validationErrors)
	}
}

func TestCreateFollowUpRejectsUnassignedUser(t *testing.T) {
	repository := &fakeRepository{
		taskCustomer: domain.TaskCustomer{ID: 10, UUID: "task-customer-uuid", Status: "pending", AssignedUserID: 99},
	}
	service := NewService(repository, &fakeStorage{})

	_, validationErrors, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput())

	if !errors.Is(err, ErrInvalidFollowUpCreateInput) {
		t.Fatalf("expected ErrInvalidFollowUpCreateInput, got %v", err)
	}
	if validationErrors["tasks_customer_uuid"] != "Bu görev müşterisi için takip kaydı oluşturma yetkiniz yok." {
		t.Fatalf("expected assigned user validation error, got %#v", validationErrors)
	}
}

func TestCreateFollowUpPassesResolvedTaskCustomerIDToRepository(t *testing.T) {
	repository := &fakeRepository{
		taskCustomer: domain.TaskCustomer{ID: 10, UUID: "task-customer-uuid", Status: "pending", AssignedUserID: 20},
	}
	service := NewService(repository, &fakeStorage{})

	_, validationErrors, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput())

	if err != nil {
		t.Fatalf("expected nil error, got %v with validation errors %#v", err, validationErrors)
	}
	if repository.createInput.TasksCustomerID != 10 {
		t.Fatalf("expected tasks customer id 10, got %d", repository.createInput.TasksCustomerID)
	}
	if repository.createInput.TasksCustomerUUID != "task-customer-uuid" {
		t.Fatalf("expected tasks customer uuid, got %q", repository.createInput.TasksCustomerUUID)
	}
}

func TestCreateFollowUpDeletesStoredImagesWhenRepositoryFails(t *testing.T) {
	repository := &fakeRepository{
		taskCustomer: domain.TaskCustomer{ID: 10, UUID: "task-customer-uuid", Status: "pending", AssignedUserID: 20},
		createErr:    errors.New("create failed"),
	}
	storage := &fakeStorage{
		savedImages: []domain.StoredImage{{UUID: "image-uuid", Path: "storage/follow-ups/image.png", URL: "/storage/follow-ups/image.png"}},
	}
	service := NewService(repository, storage)

	_, _, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput())

	if !errors.Is(err, ErrFollowUpCreateUnavailable) {
		t.Fatalf("expected ErrFollowUpCreateUnavailable, got %v", err)
	}
	if !storage.deleteInvoked {
		t.Fatal("expected stored images to be deleted")
	}
}

func validCreateFollowUpInput(mutators ...func(*domain.CreateFollowUpInput)) domain.CreateFollowUpInput {
	agreementReached := false
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	input := domain.CreateFollowUpInput{
		AuthenticatedUserID:    20,
		TasksCustomerUUID:      "task-customer-uuid",
		VisitDate:              today,
		NextVisitDate:          tomorrow,
		AgreementReached:       &agreementReached,
		AgreementFailureReason: "Fiyat yüksek",
		MeetPeople: []domain.MeetPersonInput{
			{
				Title:   "Genel Müdür",
				Name:    "Ali",
				Surname: "Veli",
				Phone:   "05555555555",
			},
		},
	}

	for _, mutate := range mutators {
		mutate(&input)
	}

	return input
}
