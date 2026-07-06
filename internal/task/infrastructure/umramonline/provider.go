package umramonline

import (
	"context"

	taskapp "github.com/umran/new.crm/backend/internal/task/application"
	"github.com/umran/new.crm/backend/internal/task/domain"
	integration "github.com/umran/new.crm/backend/internal/umramonline"
)

type Provider struct {
	client *integration.Client
}

func NewProvider(client *integration.Client) *Provider {
	return &Provider{client: client}
}

func (p *Provider) GetBranch(ctx context.Context, branchID uint64) (domain.Branch, error) {
	branch, err := p.client.GetBranch(ctx, branchID)
	if err != nil {
		return domain.Branch{}, err
	}

	return domain.Branch{
		ID:    branch.ID,
		Name:  branch.Name,
		Title: branch.Title,
	}, nil
}

func (p *Provider) GetBranchUser(ctx context.Context, branchID uint64, userID uint64) (domain.BranchUser, error) {
	user, err := p.client.GetBranchUser(ctx, branchID, userID)
	if err != nil {
		return domain.BranchUser{}, err
	}

	return domain.BranchUser{
		ID:    user.ID,
		Name:  user.Name,
		Phone: user.Phone,
	}, nil
}

func (p *Provider) SendTaskCreatedSMS(ctx context.Context, input domain.TaskCreatedSMSInput) error {
	return p.client.SendTaskCreatedSMS(
		ctx,
		input.Phone,
		input.TaskUUID,
		input.Title,
		input.AssignedUserFullName,
		input.BranchName,
		input.VisitDate,
		input.DueDate,
		input.Priority,
	)
}

var _ taskapp.ReferenceProvider = (*Provider)(nil)
