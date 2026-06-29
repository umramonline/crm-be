package persistence

import (
	"context"
	"strconv"
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

func (r *Repository) ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if r == nil || r.db == nil {
		return domain.ListResult{}, gorm.ErrInvalidDB
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	dbQuery := r.db.WithContext(ctx).Model(&CustomerModel{})
	dbQuery = applyCustomerFilters(dbQuery, query)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return domain.ListResult{}, err
	}

	if query.SortBy == "created_at" {
		sortOrder := "desc"
		if strings.ToLower(query.SortOrder) == "asc" {
			sortOrder = "asc"
		}
		dbQuery = dbQuery.Order("created_at " + sortOrder)
	} else {
		dbQuery = dbQuery.Order("id DESC")
	}

	var customers []CustomerModel
	if err := dbQuery.Offset((page - 1) * perPage).Limit(perPage).Find(&customers).Error; err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.Customer, 0, len(customers))
	for _, customer := range customers {
		items = append(items, toCustomer(customer))
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage <= 0 {
		lastPage = 1
	}

	var from *int
	var to *int
	if total > 0 {
		fromValue := ((page - 1) * perPage) + 1
		toValue := fromValue + len(items) - 1
		from = &fromValue
		to = &toValue
	}

	return domain.ListResult{
		Items: items,
		Pagination: domain.Pagination{
			CurrentPage: page,
			LastPage:    lastPage,
			PerPage:     perPage,
			Total:       int(total),
			From:        from,
			To:          to,
		},
	}, nil
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

func (r *Repository) GetCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error) {
	if r == nil || r.db == nil || id == 0 {
		return domain.CustomerDetail{}, gorm.ErrInvalidDB
	}

	var customer CustomerModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&customer).Error; err != nil {
		return domain.CustomerDetail{}, err
	}

	return toCustomerDetail(customer), nil
}

func (r *Repository) PhoneExists(ctx context.Context, phone string) (bool, error) {
	normalizedPhone := strings.TrimSpace(phone)
	if r == nil || r.db == nil || normalizedPhone == "" {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&CustomerModel{}).
		Where("cep = ? OR telefon = ?", normalizedPhone, normalizedPhone).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) CreateCustomer(ctx context.Context, input domain.CreateCustomerInput) (domain.CustomerDetail, error) {
	if r == nil || r.db == nil {
		return domain.CustomerDetail{}, gorm.ErrInvalidDB
	}

	customer := CustomerModel{
		BranchID:   &input.BranchID,
		Unvan:      stringPointer(input.Unvan),
		Ad:         stringPointer(input.Ad),
		Soyad:      stringPointer(input.Soyad),
		YetkiliAdi: stringPointer(input.YetkiliAdi),
		Cep:        stringPointer(input.Cep),
		Telefon:    stringPointer(input.Telefon),
		Mahalle:    stringPointer(input.Mahalle),
		IlKodu:     stringPointer(input.IlKodu),
		IlceKodu:   stringPointer(input.IlceKodu),
		Type:       stringPointer(input.Type),
	}

	if input.Type == "bireysel" {
		customer.Unvan = nil
		customer.YetkiliAdi = nil
		customer.Telefon = nil
	} else {
		customer.Ad = nil
		customer.Soyad = nil
		customer.Cep = nil
	}

	if err := r.db.WithContext(ctx).Create(&customer).Error; err != nil {
		return domain.CustomerDetail{}, err
	}

	return toCustomerDetail(customer), nil
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

func toCustomer(customer CustomerModel) domain.Customer {
	createdAt := customer.CreatedAt.Format("2006-01-02 15:04:05")
	branchID := ""
	if customer.BranchID != nil {
		branchID = strconv.FormatInt(int64(*customer.BranchID), 10)
	}

	return domain.Customer{
		ID:         customer.ID,
		Situation:  "Potansiyel Müşteri",
		Unvan:      stringValue(customer.Unvan),
		Cep:        stringValue(customer.Cep),
		Ad:         stringValue(customer.Ad),
		Soyad:      stringValue(customer.Soyad),
		BranchName: branchID,
		ZoneName:   "",
		PlusCardNo: "",
		Credit:     "0",
		Source:     "Manuel",
		City:       stringValue(customer.IlKodu),
		Town:       stringValue(customer.IlceKodu),
		CreatedAt:  &createdAt,
		Type:       stringValue(customer.Type),
	}
}

func applyCustomerFilters(query *gorm.DB, filters domain.ListQuery) *gorm.DB {
	if strings.TrimSpace(filters.Unvan) != "" {
		query = query.Where("unvan LIKE ?", "%"+strings.TrimSpace(filters.Unvan)+"%")
	}

	if strings.TrimSpace(filters.Cep) != "" {
		query = query.Where("cep LIKE ? OR telefon LIKE ?", "%"+strings.TrimSpace(filters.Cep)+"%", "%"+strings.TrimSpace(filters.Cep)+"%")
	}

	if strings.TrimSpace(filters.Ad) != "" {
		query = query.Where("ad LIKE ?", "%"+strings.TrimSpace(filters.Ad)+"%")
	}

	if strings.TrimSpace(filters.Soyad) != "" {
		query = query.Where("soyad LIKE ?", "%"+strings.TrimSpace(filters.Soyad)+"%")
	}

	if strings.TrimSpace(filters.CreatedAt) != "" {
		query = query.Where("created_at LIKE ?", "%"+strings.TrimSpace(filters.CreatedAt)+"%")
	}

	if strings.TrimSpace(filters.Type) != "" {
		query = query.Where("LOWER(type) = ?", strings.ToLower(strings.TrimSpace(filters.Type)))
	}

	if len(filters.BranchIDs) > 0 {
		query = query.Where("branch_id IN ?", filters.BranchIDs)
	}

	if len(filters.CityIDs) > 0 {
		query = query.Where("il_kodu IN ?", filters.CityIDs)
	}

	if len(filters.TownIDs) > 0 {
		query = query.Where("ilce_kodu IN ?", filters.TownIDs)
	}

	return query
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func stringPointer(value string) *string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}

var _ application.CustomerRepository = (*Repository)(nil)
