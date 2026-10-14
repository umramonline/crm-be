package umramonline

import (
	"context"

	dashboardapp "github.com/umran/new.crm/backend/internal/dashboard/application"
	"github.com/umran/new.crm/backend/internal/dashboard/domain"
	integration "github.com/umran/new.crm/backend/internal/umramonline"
)

type Provider struct {
	client *integration.Client
}

func NewProvider(client *integration.Client) *Provider {
	return &Provider{client: client}
}

func (p *Provider) GetStats(ctx context.Context, filter domain.Filter) (dashboardapp.ExternalStats, error) {
	stats, err := p.client.GetDashboardStats(ctx, integration.DashboardStatsQuery{
		StartDate:        filter.StartDate,
		EndDate:          filter.EndDate,
		BranchIDs:        filter.BranchIDs,
		AllowAllBranches: filter.AllowAllBranches,
	})
	if err != nil {
		return dashboardapp.ExternalStats{}, err
	}

	return dashboardapp.ExternalStats{
		VehicleEntryCount: stats.VehicleEntryCount,
		TotalAmount:       stats.TotalAmount,
		LoadedCredit:      stats.LoadedCredit,
	}, nil
}

var _ dashboardapp.UmramonlineStatsProvider = (*Provider)(nil)
