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

func (p *Provider) CountVehicleEntries(ctx context.Context, filter domain.Filter) (int64, error) {
	return p.client.DashboardVehicleEntryCount(ctx, toStatsQuery(filter))
}

func (p *Provider) SumTotalAmount(ctx context.Context, filter domain.Filter) (float64, error) {
	return p.client.DashboardTotalAmount(ctx, toStatsQuery(filter))
}

func (p *Provider) SumLoadedCredit(ctx context.Context, filter domain.Filter) (float64, error) {
	return p.client.DashboardLoadedCredit(ctx, toStatsQuery(filter))
}

var _ dashboardapp.UmramonlineStatsProvider = (*Provider)(nil)

func toStatsQuery(filter domain.Filter) integration.DashboardStatsQuery {
	return integration.DashboardStatsQuery{
		StartDate:        filter.StartDate,
		EndDate:          filter.EndDate,
		BranchIDs:        filter.BranchIDs,
		AllowAllBranches: filter.AllowAllBranches,
	}
}
