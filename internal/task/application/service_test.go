package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/umran/new.crm/backend/internal/task/domain"
)

type fakeReferenceProvider struct {
	branchErr     error
	branchUserErr error
}

func (f fakeReferenceProvider) GetBranch(_ context.Context, branchID uint64) (domain.Branch, error) {
	if f.branchErr != nil {
		return domain.Branch{}, f.branchErr
	}

	return domain.Branch{ID: branchID, Name: "Merkez"}, nil
}

func (f fakeReferenceProvider) GetBranchUser(_ context.Context, _ uint64, userID uint64) (domain.BranchUser, error) {
	if f.branchUserErr != nil {
		return domain.BranchUser{}, f.branchUserErr
	}

	return domain.BranchUser{ID: userID, Name: "Test User"}, nil
}

type fakeTaskRepository struct {
	input domain.CreateTaskInput
}

func (f *fakeTaskRepository) InvalidCustomerIDsForBranch(_ context.Context, _ []uint64, _ uint64) ([]uint64, error) {
	return nil, nil
}

func (f *fakeTaskRepository) CreateTask(_ context.Context, input domain.CreateTaskInput) (domain.Task, error) {
	f.input = input

	return domain.Task{
		ID:             1,
		Title:          input.Title,
		AssignedUserID: input.AssignedUserID,
		BranchID:       input.BranchID,
		BranchName:     input.BranchName,
		Status:         "pending",
		Priority:       input.Priority,
		CustomerIDs:    input.CustomerIDs,
	}, nil
}

func TestCreateTaskRejectsMissingBranchName(t *testing.T) {
	service := NewService(fakeReferenceProvider{}, &fakeTaskRepository{})

	_, validationErrors, err := service.CreateTask(context.Background(), validCreateTaskInput(func(input *domain.CreateTaskInput) {
		input.BranchName = " "
	}))

	if !errors.Is(err, ErrInvalidTaskCreateInput) {
		t.Fatalf("expected ErrInvalidTaskCreateInput, got %v", err)
	}

	if validationErrors["branch_name"] != "Bayi adı zorunludur." {
		t.Fatalf("expected branch_name validation error, got %#v", validationErrors)
	}
}

func TestCreateTaskRejectsTooLongBranchName(t *testing.T) {
	service := NewService(fakeReferenceProvider{}, &fakeTaskRepository{})

	_, validationErrors, err := service.CreateTask(context.Background(), validCreateTaskInput(func(input *domain.CreateTaskInput) {
		input.BranchName = strings.Repeat("a", 51)
	}))

	if !errors.Is(err, ErrInvalidTaskCreateInput) {
		t.Fatalf("expected ErrInvalidTaskCreateInput, got %v", err)
	}

	if validationErrors["branch_name"] != "Bayi adı en fazla 50 karakter olabilir." {
		t.Fatalf("expected branch_name max length validation error, got %#v", validationErrors)
	}
}

func TestCreateTaskPassesNormalizedBranchNameToRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewService(fakeReferenceProvider{}, repository)

	task, validationErrors, err := service.CreateTask(context.Background(), validCreateTaskInput(func(input *domain.CreateTaskInput) {
		input.BranchName = "  Merkez Bayi  "
	}))

	if err != nil {
		t.Fatalf("expected nil error, got %v with validation errors %#v", err, validationErrors)
	}

	if repository.input.BranchName != "Merkez Bayi" {
		t.Fatalf("expected normalized branch name in repository input, got %q", repository.input.BranchName)
	}

	if task.BranchName != "Merkez Bayi" {
		t.Fatalf("expected task branch name, got %q", task.BranchName)
	}
}

func validCreateTaskInput(mutators ...func(*domain.CreateTaskInput)) domain.CreateTaskInput {
	input := domain.CreateTaskInput{
		Title:                "Ziyaret",
		AssignedUserID:       10,
		AssignedUserFullName: "Assigned User",
		CreatedByUserID:      20,
		BranchID:             30,
		BranchName:           "Merkez",
		Priority:             "medium",
		CustomerIDs:          []uint64{40},
	}

	for _, mutate := range mutators {
		mutate(&input)
	}

	return input
}
