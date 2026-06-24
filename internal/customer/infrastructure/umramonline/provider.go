package umramonline

import (
	"context"

	customerapp "github.com/umran/new.crm/backend/internal/customer/application"
	"github.com/umran/new.crm/backend/internal/customer/domain"
	integration "github.com/umran/new.crm/backend/internal/umramonline"
)

type Provider struct {
	client *integration.Client
}

func NewProvider(client *integration.Client) *Provider {
	return &Provider{client: client}
}

func (p *Provider) ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	result, err := p.client.ListCustomers(ctx, integration.CustomerListQuery{
		Page:       query.Page,
		PerPage:    query.PerPage,
		Situation:  query.Situation,
		Unvan:      query.Unvan,
		Cep:        query.Cep,
		Ad:         query.Ad,
		Soyad:      query.Soyad,
		BranchName: query.BranchName,
		PlusCardNo: query.PlusCardNo,
		Source:     query.Source,
		City:       query.City,
		Town:       query.Town,
		CreatedAt:  query.CreatedAt,
		Type:       query.Type,
		SortBy:     query.SortBy,
		SortOrder:  query.SortOrder,
	})
	if err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.Customer, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, domain.Customer{
			Situation:    item.Situation,
			Unvan:        item.Unvan,
			Cep:          item.Cep,
			Ad:           item.Ad,
			Soyad:        item.Soyad,
			BranchName:   item.BranchName,
			PlusCardNo:   item.PlusCardNo,
			Credit:       item.Credit,
			Source:       item.Source,
			City:         item.City,
			Town:         item.Town,
			CreatedAt:    item.CreatedAt,
			Type:         item.Type,
			DaysSpending: item.DaysSpending,
			DaysLoading:  item.DaysLoading,
		})
	}

	return domain.ListResult{
		Items: items,
		Pagination: domain.Pagination{
			CurrentPage: result.Pagination.CurrentPage,
			LastPage:    result.Pagination.LastPage,
			PerPage:     result.Pagination.PerPage,
			Total:       result.Pagination.Total,
			From:        result.Pagination.From,
			To:          result.Pagination.To,
		},
	}, nil
}

var _ customerapp.CustomerProvider = (*Provider)(nil)
