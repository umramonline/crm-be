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
		BranchName: query.BranchName,
		ZoneName:   query.ZoneName,
		PlusCardNo: query.PlusCardNo,
		City:       query.City,
		Town:       query.Town,
		SortBy:     query.SortBy,
		SortOrder:  query.SortOrder,
		ZoneID:     query.ZoneID,
		BranchIDs:  query.BranchIDs,
		IDs:        query.IDs,
	})
	if err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.Customer, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, domain.Customer{
			UOId:       item.ID,
			Situation:  item.Situation,
			BranchName: item.BranchName,
			ZoneName:   item.ZoneName,
			PlusCardNo: item.PlusCardNo,
			Credit:     item.Credit,
			Point:      item.Point,
			City:       item.City,
			Town:       item.Town,
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

func (p *Provider) ListZones(ctx context.Context, branchIDs []uint64) ([]domain.Zone, error) {
	zones, err := p.client.ListZones(ctx, branchIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Zone, 0, len(zones))
	for _, zone := range zones {
		result = append(result, domain.Zone{
			ID:   zone.ID,
			Name: zone.Name,
		})
	}

	return result, nil
}

func (p *Provider) SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error) {
	customer, found, err := p.client.SearchCustomer(ctx, query)
	if err != nil || !found {
		return domain.CustomerDetail{}, false, err
	}

	return domain.CustomerDetail{
		ID:         customer.ID,
		UOId:       customer.UOId,
		BranchID:   customer.BranchID,
		Unvan:      customer.Unvan,
		Ad:         customer.Ad,
		Soyad:      customer.Soyad,
		YetkiliAdi: customer.YetkiliAdi,
		Cep:        customer.Cep,
		Telefon:    customer.Telefon,
		Mahalle:    customer.Mahalle,
		IlKodu:     customer.IlKodu,
		IlceKodu:   customer.IlceKodu,
		VergiNo:    customer.VergiNo,
		TCNo:       customer.TCNo,
		Type:       customer.Type,
		CreatedAt:  customer.CreatedAt,
		PlusCardNo: customer.PlusCardNo,
		Credit:     customer.Credit,
		Point:      customer.Point,
	}, true, nil
}

func (p *Provider) GetCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error) {
	customer, err := p.client.GetCustomer(ctx, id)
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	return domain.CustomerDetail{
		ID:         customer.ID,
		UOId:       customer.UOId,
		BranchID:   customer.BranchID,
		Unvan:      customer.Unvan,
		Ad:         customer.Ad,
		Soyad:      customer.Soyad,
		YetkiliAdi: customer.YetkiliAdi,
		Cep:        customer.Cep,
		Telefon:    customer.Telefon,
		Mahalle:    customer.Mahalle,
		IlKodu:     customer.IlKodu,
		IlceKodu:   customer.IlceKodu,
		VergiNo:    customer.VergiNo,
		TCNo:       customer.TCNo,
		Type:       customer.Type,
		CreatedAt:  customer.CreatedAt,
		PlusCardNo: customer.PlusCardNo,
		Credit:     customer.Credit,
		Point:      customer.Point,
	}, nil
}

func (p *Provider) PhoneExists(ctx context.Context, phone string) (bool, error) {
	return p.client.CustomerPhoneExists(ctx, phone)
}

func (p *Provider) ListCities(ctx context.Context) ([]domain.City, error) {
	cities, err := p.client.ListCities(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.City, 0, len(cities))
	for _, city := range cities {
		result = append(result, domain.City{
			ID:    city.ID,
			Title: city.Title,
		})
	}

	return result, nil
}

func (p *Provider) ListTowns(ctx context.Context, cityID uint64) ([]domain.Town, error) {
	towns, err := p.client.ListTowns(ctx, cityID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Town, 0, len(towns))
	for _, town := range towns {
		result = append(result, domain.Town{
			ID:        town.ID,
			Title:     town.Title,
			CityID:    town.CityID,
			CityTitle: town.CityTitle,
		})
	}

	return result, nil
}

func (p *Provider) ListBranches(ctx context.Context, branchIDs []uint64) ([]domain.Branch, error) {
	branches, err := p.client.ListBranches(ctx, branchIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Branch, 0, len(branches))
	for _, branch := range branches {
		result = append(result, domain.Branch{
			ID:    branch.ID,
			Name:  branch.Name,
			Title: branch.Title,
		})
	}

	return result, nil
}

func (p *Provider) ListBranchUsers(ctx context.Context, branchID uint64) ([]domain.BranchUser, error) {
	users, err := p.client.ListBranchUsers(ctx, branchID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.BranchUser, 0, len(users))
	for _, user := range users {
		result = append(result, domain.BranchUser{
			ID:    user.ID,
			Name:  user.Name,
			Phone: user.Phone,
		})
	}

	return result, nil
}

var _ customerapp.CustomerProvider = (*Provider)(nil)
