package persistence

import (
	"context"
	"strings"

	"github.com/umran/new.crm/backend/internal/customer/application"
	"github.com/umran/new.crm/backend/internal/customer/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error) {
	normalizedQuery := strings.TrimSpace(query)
	if r == nil || r.db == nil || normalizedQuery == "" {
		return domain.CustomerDetail{}, false, nil
	}

	pattern := "%" + normalizedQuery + "%"
	var customer CustomerModel
	err := r.db.WithContext(ctx).
		Where("cep LIKE ?", pattern).
		Or("telefon LIKE ?", pattern).
		Or("tc_no LIKE ?", pattern).
		Or("vergi_no LIKE ?", pattern).
		Order("id DESC").
		First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.CustomerDetail{}, false, nil
		}

		return domain.CustomerDetail{}, false, err
	}

	return toCustomerDetail(customer), true, nil
}

func toCustomerDetail(customer CustomerModel) domain.CustomerDetail {
	createdAt := customer.CreatedAt.Format("2006-01-02 15:04:05")

	return domain.CustomerDetail{
		ID:         customer.ID,
		UOId:       customer.UOId,
		BranchID:   customer.BranchID,
		Unvan:      stringValue(customer.Unvan),
		Ad:         stringValue(customer.Ad),
		Soyad:      stringValue(customer.Soyad),
		YetkiliAdi: stringValue(customer.YetkiliAdi),
		Cep:        stringValue(customer.Cep),
		Telefon:    stringValue(customer.Telefon),
		Mahalle:    stringValue(customer.Mahalle),
		IlKodu:     stringValue(customer.IlKodu),
		IlceKodu:   stringValue(customer.IlceKodu),
		VergiNo:    stringValue(customer.VergiNo),
		TCNo:       stringValue(customer.TCNo),
		Type:       stringValue(customer.Type),
		CreatedAt:  &createdAt,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

var _ application.CustomerRepository = (*Repository)(nil)
