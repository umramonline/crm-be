package umramonline

import (
	"context"

	"github.com/umran/new.crm/backend/internal/authorization/domain"
	integration "github.com/umran/new.crm/backend/internal/umramonline"
)

type Provider struct {
	client *integration.Client
}

func NewProvider(client *integration.Client) *Provider {
	return &Provider{client: client}
}

func (p *Provider) ListRoles(ctx context.Context) ([]domain.Role, error) {
	roles, err := p.client.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Role, 0, len(roles))
	for _, role := range roles {
		result = append(result, domain.Role{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	return result, nil
}
