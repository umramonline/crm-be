package persistence

import (
	"context"
	"strings"
	"time"

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

	dbQuery = applyCustomerSort(dbQuery, query)

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

func (r *Repository) ListCustomerUOIds(ctx context.Context, query domain.ListQuery) ([]uint64, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	dbQuery := r.db.WithContext(ctx).Model(&CustomerModel{}).
		Select("uo_id").
		Where("uo_id > 0")
	dbQuery = applyCustomerFilters(dbQuery, query)

	var uoIDs []uint64
	if err := dbQuery.Order("id DESC").Pluck("uo_id", &uoIDs).Error; err != nil {
		return nil, err
	}

	return uoIDs, nil
}

func (r *Repository) ListCustomersByUOIds(ctx context.Context, uoIDs []uint64) ([]domain.Customer, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if len(uoIDs) == 0 {
		return []domain.Customer{}, nil
	}

	var customers []CustomerModel
	if err := r.db.WithContext(ctx).
		Where("uo_id IN ?", uoIDs).
		Find(&customers).Error; err != nil {
		return nil, err
	}

	items := make([]domain.Customer, 0, len(customers))
	for _, customer := range customers {
		items = append(items, toCustomer(customer))
	}

	return items, nil
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

func (r *Repository) GetFullRegistrationCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error) {
	customer, err := r.GetCustomer(ctx, id)
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	telephones, err := r.customerTelephones(ctx, id)
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	customer.Telephones = telephones

	return customer, nil
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

func (r *Repository) PhoneExistsExcept(ctx context.Context, phone string, customerID uint64) (bool, error) {
	normalizedPhone := strings.TrimSpace(phone)
	if r == nil || r.db == nil || normalizedPhone == "" {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&CustomerModel{}).
		Where("id <> ?", customerID).
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

func (r *Repository) CreateCustomerFromIetts(
	ctx context.Context,
	unvan string,
	ad string,
	soyad string,
	addressDetail string,
) (uint64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	customer := CustomerModel{
		Unvan:         stringPointer(unvan),
		Ad:            stringPointer(ad),
		Soyad:         stringPointer(soyad),
		AddressDetail: stringPointer(addressDetail),
		Type:          stringPointer("kurumsal"),
	}

	if err := r.db.WithContext(ctx).Create(&customer).Error; err != nil {
		return 0, err
	}

	return customer.ID, nil
}

func (r *Repository) CompleteFullRegistration(ctx context.Context, id uint64, input domain.FullRegistrationInput) (domain.CustomerDetail, error) {
	var customer CustomerModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&customer).Error; err != nil {
			return err
		}

		customer.Type = stringPointer(input.Type)
		customer.Cep = stringPointer(input.Cep)
		customer.Ad = stringPointer(input.Ad)
		customer.Soyad = stringPointer(input.Soyad)
		customer.Unvan = stringPointer(input.Unvan)
		customer.CorporateSector = stringPointer(input.CorporateSector)
		customer.TCNo = stringPointer(input.TCNo)
		customer.DogumTarihi = datePointer(input.DogumTarihi)
		customer.Eposta = stringPointer(input.Eposta)
		customer.Website = stringPointer(input.Website)
		customer.Web = stringPointer(input.Website)
		customer.GoogleMapLink = stringPointer(input.GoogleMapLink)
		customer.ClassifiedsWebsiteLink = stringPointer(input.ClassifiedsWebsiteLink)
		customer.VehicleStockCount = &input.VehicleStockCount
		customer.BranchID = &input.BranchID
		customer.VergiNo = stringPointer(input.VergiNo)
		customer.VergiDairesi = stringPointer(input.VergiDairesi)
		customer.IlKodu = stringPointer(input.IlKodu)
		customer.IlceKodu = stringPointer(input.IlceKodu)
		customer.Mahalle = stringPointer(input.Mahalle)
		customer.AddressDetail = stringPointer(input.AddressDetail)

		if err := tx.Save(&customer).Error; err != nil {
			return err
		}

		if err := tx.Where("customer_id = ?", id).Delete(&CustomerTelephoneModel{}).Error; err != nil {
			return err
		}

		for _, telephone := range input.Telephones {
			if strings.TrimSpace(telephone.PhoneNumber) == "" && strings.TrimSpace(telephone.Title) == "" {
				continue
			}

			model := CustomerTelephoneModel{
				CustomerID:  id,
				PhoneNumber: strings.TrimSpace(telephone.PhoneNumber),
				Title:       stringPointer(telephone.Title),
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	return r.GetFullRegistrationCustomer(ctx, id)
}

func (r *Repository) UpdateSourceEditableFullRegistration(ctx context.Context, id uint64, input domain.FullRegistrationInput) (domain.CustomerDetail, error) {
	err := r.db.WithContext(ctx).
		Model(&CustomerModel{}).
		Where("id = ? AND uo_id <> ?", id, 0).
		Updates(map[string]interface{}{
			"corporate_sector":         stringPointer(input.CorporateSector),
			"website":                  stringPointer(input.Website),
			"web":                      stringPointer(input.Website),
			"google_map_link":          stringPointer(input.GoogleMapLink),
			"classifieds_website_link": stringPointer(input.ClassifiedsWebsiteLink),
			"vehicle_stock_count":      input.VehicleStockCount,
		}).Error
	if err != nil {
		return domain.CustomerDetail{}, err
	}

	return r.GetFullRegistrationCustomer(ctx, id)
}

func toCustomerDetail(customer CustomerModel) domain.CustomerDetail {
	createdAt := customer.CreatedAt.Format("2006-01-02 15:04:05")
	dogumTarihi := ""
	if customer.DogumTarihi != nil {
		dogumTarihi = customer.DogumTarihi.Format("2006-01-02")
	}

	return domain.CustomerDetail{
		ID:                     customer.ID,
		UOId:                   customer.UOId,
		BranchID:               customer.BranchID,
		Unvan:                  stringValue(customer.Unvan),
		Ad:                     stringValue(customer.Ad),
		Soyad:                  stringValue(customer.Soyad),
		YetkiliAdi:             stringValue(customer.YetkiliAdi),
		Cep:                    stringValue(customer.Cep),
		Telefon:                stringValue(customer.Telefon),
		Eposta:                 stringValue(customer.Eposta),
		Website:                firstNonEmptyString(customer.Website, customer.Web),
		GoogleMapLink:          stringValue(customer.GoogleMapLink),
		ClassifiedsWebsiteLink: stringValue(customer.ClassifiedsWebsiteLink),
		Mahalle:                stringValue(customer.Mahalle),
		AddressDetail:          stringValue(customer.AddressDetail),
		IlKodu:                 stringValue(customer.IlKodu),
		IlceKodu:               stringValue(customer.IlceKodu),
		VergiNo:                stringValue(customer.VergiNo),
		VergiDairesi:           stringValue(customer.VergiDairesi),
		TCNo:                   stringValue(customer.TCNo),
		DogumTarihi:            dogumTarihi,
		VehicleStockCount:      customer.VehicleStockCount,
		CorporateSector:        stringValue(customer.CorporateSector),
		Type:                   stringValue(customer.Type),
		CreatedAt:              &createdAt,
	}
}

func toCustomer(customer CustomerModel) domain.Customer {
	createdAt := customer.CreatedAt.Format("2006-01-02 15:04:05")

	return domain.Customer{
		ID:                customer.ID,
		UOId:              customer.UOId,
		Unvan:             stringValue(customer.Unvan),
		Cep:               stringValue(customer.Cep),
		Ad:                stringValue(customer.Ad),
		Soyad:             stringValue(customer.Soyad),
		CreatedAt:         &createdAt,
		VehicleStockCount: customer.VehicleStockCount,
		Type:              stringValue(customer.Type),
	}
}

func applyCustomerSort(query *gorm.DB, filters domain.ListQuery) *gorm.DB {
	if filters.SortBy == "created_at" || filters.SortBy == "vehicle_stock_count" {
		sortOrder := "desc"
		if strings.ToLower(filters.SortOrder) == "asc" {
			sortOrder = "asc"
		}
		return query.Order(filters.SortBy + " " + sortOrder)
	}

	return query.Order("id DESC")
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

func datePointer(value string) *time.Time {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	date, err := time.Parse("2006-01-02", trimmedValue)
	if err != nil {
		return nil
	}

	return &date
}

func firstNonEmptyString(values ...*string) string {
	for _, value := range values {
		if strings.TrimSpace(stringValue(value)) != "" {
			return stringValue(value)
		}
	}

	return ""
}

func (r *Repository) customerTelephones(ctx context.Context, customerID uint64) ([]domain.CustomerTelephone, error) {
	var models []CustomerTelephoneModel
	if err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("id ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	telephones := make([]domain.CustomerTelephone, 0, len(models))
	for _, model := range models {
		telephones = append(telephones, domain.CustomerTelephone{
			ID:          model.ID,
			PhoneNumber: model.PhoneNumber,
			Title:       stringValue(model.Title),
		})
	}

	return telephones, nil
}

var _ application.CustomerRepository = (*Repository)(nil)
