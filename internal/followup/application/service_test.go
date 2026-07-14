package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umran/new.crm/backend/internal/followup/domain"
)

type fakeRepository struct {
	taskCustomer          domain.TaskCustomer
	findErr               error
	createErr             error
	createInput           domain.PersistFollowUpInput
	standaloneInput       domain.PersistStandaloneFollowUpInput
	customerAccessible    bool
	customerAccessibleErr error
}

func (f *fakeRepository) FindTaskCustomerByUUID(_ context.Context, _ string) (domain.TaskCustomer, error) {
	if f.findErr != nil {
		return domain.TaskCustomer{}, f.findErr
	}

	return f.taskCustomer, nil
}

func (f *fakeRepository) CustomerExistsForBranches(_ context.Context, _ uint64, _ []uint64, _ bool) (bool, error) {
	if f.customerAccessibleErr != nil {
		return false, f.customerAccessibleErr
	}

	return f.customerAccessible, nil
}

func (f *fakeRepository) CreateFollowUp(_ context.Context, input domain.PersistFollowUpInput) (domain.FollowUp, error) {
	f.createInput = input
	if f.createErr != nil {
		return domain.FollowUp{}, f.createErr
	}

	return domain.FollowUp{UUID: input.UUID, TasksCustomerUUID: input.TasksCustomerUUID}, nil
}

func (f *fakeRepository) CreateStandaloneFollowUp(_ context.Context, input domain.PersistStandaloneFollowUpInput) (domain.FollowUp, error) {
	f.standaloneInput = input
	f.createInput = input.FollowUp
	if f.createErr != nil {
		return domain.FollowUp{}, f.createErr
	}

	return domain.FollowUp{UUID: input.FollowUp.UUID}, nil
}

func (f *fakeRepository) FindFollowUpUpdateTargetByUUID(_ context.Context, uuid string) (domain.FollowUpUpdateTarget, error) {
	return domain.FollowUpUpdateTarget{
		ID:              1,
		UUID:            uuid,
		TasksCustomerID: 10,
		AssignedUserID:  20,
		VisitDate:       time.Now().Format("2006-01-02"),
	}, nil
}

func (f *fakeRepository) UpdateFollowUp(_ context.Context, input domain.PersistUpdateFollowUpInput) (domain.FollowUp, []domain.StoredImage, error) {
	return domain.FollowUp{
		UUID:                   input.UUID,
		VisitDate:              input.VisitDate,
		NextVisitDate:          input.NextVisitDate,
		AgreementReached:       input.AgreementReached,
		AgreementFailureReason: input.AgreementFailureReason,
		Note:                   input.Note,
	}, nil, nil
}

func (f *fakeRepository) ListFollowUps(_ context.Context, _ domain.ListQuery) (domain.ListResult, error) {
	return domain.ListResult{}, nil
}

func (f *fakeRepository) GetFollowUp(_ context.Context, _ string) (domain.FollowUp, error) {
	return domain.FollowUp{}, nil
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

func TestCreateFollowUpRejectsMissingVisitType(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeStorage{})

	_, validationErrors, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput(func(input *domain.CreateFollowUpInput) {
		input.VisitType = " "
	}))

	if !errors.Is(err, ErrInvalidFollowUpCreateInput) {
		t.Fatalf("expected ErrInvalidFollowUpCreateInput, got %v", err)
	}
	if validationErrors["visit_type"] != "Ziyaret tipi zorunludur." {
		t.Fatalf("expected visit_type required validation error, got %#v", validationErrors)
	}
}

func TestCreateFollowUpRejectsInvalidVisitType(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeStorage{})

	_, validationErrors, err := service.CreateFollowUp(context.Background(), validCreateFollowUpInput(func(input *domain.CreateFollowUpInput) {
		input.VisitType = "Telefon Görüşmesi"
	}))

	if !errors.Is(err, ErrInvalidFollowUpCreateInput) {
		t.Fatalf("expected ErrInvalidFollowUpCreateInput, got %v", err)
	}
	if validationErrors["visit_type"] != "Ziyaret tipi geçersiz." {
		t.Fatalf("expected visit_type enum validation error, got %#v", validationErrors)
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
	if repository.createInput.VisitType != "Yerinde Ziyaret" {
		t.Fatalf("expected visit type, got %q", repository.createInput.VisitType)
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

func TestCreateStandaloneFollowUpPassesCustomerAndClaimsToRepository(t *testing.T) {
	repository := &fakeRepository{customerAccessible: true}
	service := NewService(repository, &fakeStorage{})

	_, validationErrors, err := service.CreateStandaloneFollowUp(context.Background(), validStandaloneFollowUpInput())
	if err != nil {
		t.Fatalf("expected nil error, got %v with validation errors %#v", err, validationErrors)
	}
	if repository.standaloneInput.CustomerID != 42 {
		t.Fatalf("expected customer id 42, got %d", repository.standaloneInput.CustomerID)
	}
	if repository.createInput.AssignedUserID != 20 {
		t.Fatalf("expected assigned user id 20, got %d", repository.createInput.AssignedUserID)
	}
	if repository.createInput.AssignedUserFullName != "Test User" {
		t.Fatalf("expected assigned user full name, got %q", repository.createInput.AssignedUserFullName)
	}
}

func TestCreateStandaloneFollowUpRejectsInaccessibleCustomer(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeStorage{})

	_, validationErrors, err := service.CreateStandaloneFollowUp(context.Background(), validStandaloneFollowUpInput())
	if !errors.Is(err, ErrInvalidFollowUpCreateInput) {
		t.Fatalf("expected ErrInvalidFollowUpCreateInput, got %v", err)
	}
	if validationErrors["customer_id"] == "" {
		t.Fatalf("expected customer_id validation error, got %#v", validationErrors)
	}
}

func validCreateFollowUpInput(mutators ...func(*domain.CreateFollowUpInput)) domain.CreateFollowUpInput {
	agreementReached := false
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	input := domain.CreateFollowUpInput{
		AuthenticatedUserID:       20,
		AuthenticatedUserFullName: "Test User",
		TasksCustomerUUID:         "task-customer-uuid",
		VisitType:                 "Yerinde Ziyaret",
		VisitDate:                 today,
		NextVisitDate:             tomorrow,
		AgreementReached:          &agreementReached,
		AgreementFailureReason:    "Fiyat yüksek",
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

func validStandaloneFollowUpInput() domain.CreateStandaloneFollowUpInput {
	input := validCreateFollowUpInput()

	return domain.CreateStandaloneFollowUpInput{
		AuthenticatedUserID:       input.AuthenticatedUserID,
		AuthenticatedUserFullName: input.AuthenticatedUserFullName,
		CustomerID:                42,
		AllowedBranchIDs:          []uint64{3},
		VisitType:                 input.VisitType,
		VisitDate:                 input.VisitDate,
		NextVisitDate:             input.NextVisitDate,
		AgreementReached:          input.AgreementReached,
		AgreementFailureReason:    input.AgreementFailureReason,
		Note:                      input.Note,
		Images:                    input.Images,
		MeetPeople:                input.MeetPeople,
	}
}
