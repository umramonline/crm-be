package application

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

var ErrCustomerListUnavailable = errors.New("customer list unavailable")

var ErrZoneListUnavailable = errors.New("zone list unavailable")

var ErrCustomerSearchUnavailable = errors.New("customer search unavailable")

var ErrReferenceDataUnavailable = errors.New("reference data unavailable")

var ErrInvalidCustomerSearchQuery = errors.New("customer search query is required")

var ErrCustomerCreateUnavailable = errors.New("customer create unavailable")

var ErrInvalidCustomerCreateInput = errors.New("invalid customer create input")

type ValidationErrors map[string]string

var turkeyMobilePhonePattern = regexp.MustCompile(`^05[0-9]{9}$`)

const customerTextMaxLength = 255

var corporateSectorOptions = map[string]struct{}{
	"Teknoloji": {},
	"İnşaat":    {},
	"Otomotiv":  {},
	"Gıda":      {},
	"Tekstil":   {},
	"Sağlık":    {},
	"Eğitim":    {},
	"Finans":    {},
	"Turizm":    {},
	"Diğer":     {},
}

type CustomerProvider interface {
	ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	ListZones(ctx context.Context, branchIDs []uint64) ([]domain.Zone, error)
	SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error)
	GetCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	ListCities(ctx context.Context) ([]domain.City, error)
	ListTowns(ctx context.Context, cityID uint64) ([]domain.Town, error)
	ListBranches(ctx context.Context, branchIDs []uint64) ([]domain.Branch, error)
	ListBranchUsers(ctx context.Context, branchID uint64) ([]domain.BranchUser, error)
}

type CustomerRepository interface {
	ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	ListCustomerUOIds(ctx context.Context, query domain.ListQuery) ([]uint64, error)
	ListCustomersByUOIds(ctx context.Context, uoIDs []uint64) ([]domain.Customer, error)
	SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error)
	GetCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	PhoneExistsExcept(ctx context.Context, phone string, customerID uint64) (bool, error)
	CreateCustomer(ctx context.Context, input domain.CreateCustomerInput) (domain.CustomerDetail, error)
	GetFullRegistrationCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error)
	CompleteFullRegistration(ctx context.Context, id uint64, input domain.FullRegistrationInput) (domain.CustomerDetail, error)
	UpdateSourceEditableFullRegistration(ctx context.Context, id uint64, input domain.FullRegistrationInput) (domain.CustomerDetail, error)
}

type Service struct {
	provider   CustomerProvider
	repository CustomerRepository
}

func NewService(provider CustomerProvider, repositories ...CustomerRepository) *Service {
	var repository CustomerRepository
	if len(repositories) > 0 {
		repository = repositories[0]
	}

	return &Service{provider: provider, repository: repository}
}

func (s *Service) ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	switch customerDataSource(query.DataSource) {
	case "backend":
		return s.listBackendCustomers(ctx, query)
	case "umramonline":
		return s.provider.ListCustomers(ctx, query)
	default:
		return s.listMergedCustomers(ctx, query)
	}
}

func (s *Service) ListZones(ctx context.Context, branchIDs []uint64, includeAll bool) ([]domain.Zone, error) {
	if s == nil || s.provider == nil {
		return nil, ErrZoneListUnavailable
	}

	if !includeAll && len(branchIDs) == 0 {
		return []domain.Zone{}, nil
	}

	if includeAll {
		return s.provider.ListZones(ctx, nil)
	}

	return s.provider.ListZones(ctx, branchIDs)
}

func (s *Service) GetCustomer(ctx context.Context, id uint64, dataSource string) (domain.CustomerDetail, error) {
	if s == nil || s.provider == nil || id == 0 {
		return domain.CustomerDetail{}, ErrCustomerSearchUnavailable
	}

	if customerDataSource(dataSource) == "backend" {
		if s.repository == nil {
			return domain.CustomerDetail{}, ErrCustomerSearchUnavailable
		}

		return s.repository.GetCustomer(ctx, id)
	}

	return s.provider.GetCustomer(ctx, id)
}

func (s *Service) SearchCustomer(ctx context.Context, query string) (domain.CustomerSearchResult, error) {
	normalizedQuery := strings.TrimSpace(query)
	if normalizedQuery == "" {
		return domain.CustomerSearchResult{}, ErrInvalidCustomerSearchQuery
	}

	if s == nil || s.repository == nil || s.provider == nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	customer, found, err := s.provider.SearchCustomer(ctx, normalizedQuery)
	if err != nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	if found {
		return domain.CustomerSearchResult{
			Found:    true,
			Source:   "umramonline",
			Customer: &customer,
		}, nil
	}

	customer, found, err = s.repository.SearchCustomer(ctx, normalizedQuery)
	if err != nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	if !found {
		return domain.CustomerSearchResult{Found: false}, nil
	}

	return domain.CustomerSearchResult{
		Found:    true,
		Source:   "backend",
		Customer: &customer,
	}, nil
}

func (s *Service) GetFullRegistrationCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error) {
	return s.repository.GetFullRegistrationCustomer(ctx, id)
}

func (s *Service) CompleteFullRegistration(ctx context.Context, id uint64, input domain.FullRegistrationInput) (domain.CustomerDetail, ValidationErrors, error) {
	normalizedInput := normalizeFullRegistrationInput(input)

	if s == nil || s.repository == nil || s.provider == nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}

	existingCustomer, err := s.repository.GetFullRegistrationCustomer(ctx, id)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}

	if existingCustomer.UOId > 0 {
		validationErrors := validateSourceEditableFullRegistrationInput(normalizedInput, existingCustomer.Type)
		if len(validationErrors) > 0 {
			return domain.CustomerDetail{}, validationErrors, ErrInvalidCustomerCreateInput
		}

		customer, err := s.repository.UpdateSourceEditableFullRegistration(ctx, id, normalizedInput)
		if err != nil {
			return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
		}

		return customer, nil, nil
	}

	validationErrors := validateFullRegistrationInput(normalizedInput)
	if len(validationErrors) > 0 {
		return domain.CustomerDetail{}, validationErrors, ErrInvalidCustomerCreateInput
	}

	exists, err := s.repository.PhoneExistsExcept(ctx, normalizedInput.Cep, id)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}
	if exists {
		return domain.CustomerDetail{}, ValidationErrors{"cep": "Bu cep numarası backend müşteri kayıtlarında zaten var."}, ErrInvalidCustomerCreateInput
	}

	exists, err = s.provider.PhoneExists(ctx, normalizedInput.Cep)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}
	if exists {
		return domain.CustomerDetail{}, ValidationErrors{"cep": "Bu cep numarası umramonline müşteri kayıtlarında zaten var."}, ErrInvalidCustomerCreateInput
	}

	customer, err := s.repository.CompleteFullRegistration(ctx, id, normalizedInput)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}

	return customer, nil, nil
}

func (s *Service) FullRegistrationPhoneExists(ctx context.Context, id uint64, cep string) (bool, error) {
	normalizedCep := strings.TrimSpace(cep)
	if id == 0 || !turkeyMobilePhonePattern.MatchString(normalizedCep) {
		return false, ErrInvalidCustomerCreateInput
	}

	if s == nil || s.repository == nil || s.provider == nil {
		return false, ErrCustomerCreateUnavailable
	}

	exists, err := s.repository.PhoneExistsExcept(ctx, normalizedCep, id)
	if err != nil || exists {
		return exists, err
	}

	return s.provider.PhoneExists(ctx, normalizedCep)
}

func (s *Service) CreateCustomer(ctx context.Context, input domain.CreateCustomerInput) (domain.CustomerDetail, ValidationErrors, error) {
	normalizedInput := normalizeCreateCustomerInput(input)
	validationErrors := validateCreateCustomerInput(normalizedInput)
	if len(validationErrors) > 0 {
		return domain.CustomerDetail{}, validationErrors, ErrInvalidCustomerCreateInput
	}

	if s == nil || s.repository == nil || s.provider == nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}

	phoneField := phoneFieldForCustomerType(normalizedInput.Type)
	phone := phoneValueForCustomerType(normalizedInput)

	exists, err := s.repository.PhoneExists(ctx, phone)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}
	if exists {
		return domain.CustomerDetail{}, ValidationErrors{
			phoneField: "Bu telefon numarası backend müşteri kayıtlarında zaten var.",
		}, ErrInvalidCustomerCreateInput
	}

	exists, err = s.provider.PhoneExists(ctx, phone)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}
	if exists {
		return domain.CustomerDetail{}, ValidationErrors{
			phoneField: "Bu telefon numarası umramonline müşteri kayıtlarında zaten var.",
		}, ErrInvalidCustomerCreateInput
	}

	customer, err := s.repository.CreateCustomer(ctx, normalizedInput)
	if err != nil {
		return domain.CustomerDetail{}, nil, ErrCustomerCreateUnavailable
	}

	return customer, nil, nil
}

func (s *Service) ListCities(ctx context.Context) ([]domain.City, error) {
	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListCities(ctx)
}

func (s *Service) ListTowns(ctx context.Context, cityID uint64) ([]domain.Town, error) {
	if cityID == 0 {
		return []domain.Town{}, nil
	}

	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListTowns(ctx, cityID)
}

func (s *Service) ListBranches(ctx context.Context, branchIDs []uint64, includeAll bool) ([]domain.Branch, error) {
	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	if !includeAll && len(branchIDs) == 0 {
		return []domain.Branch{}, nil
	}

	if includeAll {
		return s.provider.ListBranches(ctx, nil)
	}

	return s.provider.ListBranches(ctx, branchIDs)
}

func (s *Service) ListBranchUsers(ctx context.Context, branchID uint64) ([]domain.BranchUser, error) {
	if s == nil || s.provider == nil || branchID == 0 {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListBranchUsers(ctx, branchID)
}

func (s *Service) listMergedCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if s == nil || s.repository == nil || s.provider == nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 10
	}

	if usesUmramonlineListControl(query) {
		uoIDs, err := s.repository.ListCustomerUOIds(ctx, query)
		if err != nil {
			return domain.ListResult{}, ErrCustomerListUnavailable
		}
		if len(uoIDs) == 0 {
			return emptyListResult(query), nil
		}

		uoQuery := domain.ListQuery{
			Page:       page,
			PerPage:    perPage,
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
			IDs:        uoIDs,
		}

		uoResult, err := s.provider.ListCustomers(ctx, uoQuery)
		if err != nil {
			return domain.ListResult{}, ErrCustomerListUnavailable
		}

		backendByUOID, err := s.backendCustomersByUOID(ctx, extractUOIds(uoResult.Items))
		if err != nil {
			return domain.ListResult{}, ErrCustomerListUnavailable
		}

		items := make([]domain.Customer, 0, len(uoResult.Items))
		for _, uoItem := range uoResult.Items {
			items = append(items, mergeCustomer(backendByUOID[uoItem.UOId], uoItem))
		}

		return domain.ListResult{
			Items:      items,
			Pagination: uoResult.Pagination,
		}, nil
	}

	backendResult, err := s.repository.ListCustomers(ctx, query)
	if err != nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}
	if len(backendResult.Items) == 0 {
		return backendResult, nil
	}

	uoIDs := make([]uint64, 0, len(backendResult.Items))
	for _, item := range backendResult.Items {
		if item.UOId > 0 {
			uoIDs = append(uoIDs, item.UOId)
		}
	}

	uoByID := map[uint64]domain.Customer{}
	if len(uoIDs) > 0 {
		uoResult, err := s.provider.ListCustomers(ctx, domain.ListQuery{
			Page:    1,
			PerPage: len(uoIDs),
			IDs:     uoIDs,
		})
		if err != nil {
			return domain.ListResult{}, ErrCustomerListUnavailable
		}
		for _, item := range uoResult.Items {
			uoByID[item.UOId] = item
		}
	}

	items := make([]domain.Customer, 0, len(backendResult.Items))
	for _, backendItem := range backendResult.Items {
		items = append(items, mergeCustomer(backendItem, uoByID[backendItem.UOId]))
	}

	backendResult.Items = items
	return backendResult, nil
}

func (s *Service) listBackendCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if s == nil || s.repository == nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	result, err := s.repository.ListCustomers(ctx, query)
	if err != nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	return result, nil
}

func (s *Service) backendCustomersByUOID(ctx context.Context, uoIDs []uint64) (map[uint64]domain.Customer, error) {
	result := make(map[uint64]domain.Customer, len(uoIDs))
	if len(uoIDs) == 0 {
		return result, nil
	}

	customers, err := s.repository.ListCustomersByUOIds(ctx, uoIDs)
	if err != nil {
		return nil, err
	}

	for _, customer := range customers {
		if customer.UOId > 0 {
			result[customer.UOId] = customer
		}
	}

	return result, nil
}

func usesUmramonlineListControl(query domain.ListQuery) bool {
	if strings.TrimSpace(query.BranchName) != "" ||
		strings.TrimSpace(query.ZoneName) != "" ||
		strings.TrimSpace(query.PlusCardNo) != "" ||
		strings.TrimSpace(query.City) != "" ||
		strings.TrimSpace(query.Town) != "" ||
		strings.TrimSpace(query.Situation) != "" {
		return true
	}

	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	return sortBy == "credit" || sortBy == "point"
}

func mergeCustomer(backend domain.Customer, umramonline domain.Customer) domain.Customer {
	merged := domain.Customer{
		ID:                backend.ID,
		UOId:              backend.UOId,
		Unvan:             backend.Unvan,
		Cep:               backend.Cep,
		Ad:                backend.Ad,
		Soyad:             backend.Soyad,
		VehicleStockCount: backend.VehicleStockCount,
		CreatedAt:         backend.CreatedAt,
		Type:              backend.Type,
		BranchName:        umramonline.BranchName,
		ZoneName:          umramonline.ZoneName,
		PlusCardNo:        umramonline.PlusCardNo,
		Credit:            umramonline.Credit,
		Point:             umramonline.Point,
		City:              umramonline.City,
		Town:              umramonline.Town,
		Situation:         umramonline.Situation,
	}

	if merged.UOId == 0 && umramonline.UOId > 0 {
		merged.UOId = umramonline.UOId
	}

	return merged
}

func extractUOIds(items []domain.Customer) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		if item.UOId > 0 {
			ids = append(ids, item.UOId)
		}
	}
	return ids
}

func customerDataSource(dataSource string) string {
	normalizedDataSource := strings.ToLower(strings.TrimSpace(dataSource))
	switch normalizedDataSource {
	case "backend":
		return "backend"
	case "umramonline":
		return "umramonline"
	default:
		return "merged"
	}
}

func emptyListResult(query domain.ListQuery) domain.ListResult {
	page := query.Page
	if page <= 0 {
		page = 1
	}

	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 10
	}

	return domain.ListResult{
		Items: []domain.Customer{},
		Pagination: domain.Pagination{
			CurrentPage: page,
			LastPage:    1,
			PerPage:     perPage,
			Total:       0,
		},
	}
}

func normalizeCreateCustomerInput(input domain.CreateCustomerInput) domain.CreateCustomerInput {
	return domain.CreateCustomerInput{
		Type:       strings.ToLower(strings.TrimSpace(input.Type)),
		Ad:         strings.TrimSpace(input.Ad),
		Soyad:      strings.TrimSpace(input.Soyad),
		Cep:        strings.TrimSpace(input.Cep),
		Unvan:      strings.TrimSpace(input.Unvan),
		YetkiliAdi: strings.TrimSpace(input.YetkiliAdi),
		Telefon:    strings.TrimSpace(input.Telefon),
		IlKodu:     strings.TrimSpace(input.IlKodu),
		IlceKodu:   strings.TrimSpace(input.IlceKodu),
		Mahalle:    strings.TrimSpace(input.Mahalle),
		BranchID:   input.BranchID,
	}
}

func normalizeFullRegistrationInput(input domain.FullRegistrationInput) domain.FullRegistrationInput {
	telephones := make([]domain.CustomerTelephone, 0, len(input.Telephones))
	for _, telephone := range input.Telephones {
		telephones = append(telephones, domain.CustomerTelephone{
			ID:          telephone.ID,
			PhoneNumber: strings.TrimSpace(telephone.PhoneNumber),
			Title:       strings.TrimSpace(telephone.Title),
		})
	}

	return domain.FullRegistrationInput{
		Type:                   strings.ToLower(strings.TrimSpace(input.Type)),
		Cep:                    strings.TrimSpace(input.Cep),
		Ad:                     strings.TrimSpace(input.Ad),
		Soyad:                  strings.TrimSpace(input.Soyad),
		Unvan:                  strings.TrimSpace(input.Unvan),
		CorporateSector:        strings.TrimSpace(input.CorporateSector),
		TCNo:                   strings.TrimSpace(input.TCNo),
		DogumTarihi:            strings.TrimSpace(input.DogumTarihi),
		Eposta:                 strings.TrimSpace(input.Eposta),
		Website:                strings.TrimSpace(input.Website),
		GoogleMapLink:          strings.TrimSpace(input.GoogleMapLink),
		ClassifiedsWebsiteLink: strings.TrimSpace(input.ClassifiedsWebsiteLink),
		VehicleStockCount:      input.VehicleStockCount,
		BranchID:               input.BranchID,
		VergiNo:                strings.TrimSpace(input.VergiNo),
		VergiDairesi:           strings.TrimSpace(input.VergiDairesi),
		Telephones:             telephones,
		IlKodu:                 strings.TrimSpace(input.IlKodu),
		IlceKodu:               strings.TrimSpace(input.IlceKodu),
		Mahalle:                strings.TrimSpace(input.Mahalle),
		AddressDetail:          strings.TrimSpace(input.AddressDetail),
	}
}

func validateCreateCustomerInput(input domain.CreateCustomerInput) ValidationErrors {
	errors := ValidationErrors{}

	if input.Type != "bireysel" && input.Type != "kurumsal" {
		errors["type"] = "Müşteri türü bireysel veya kurumsal olmalıdır."
	}

	if input.Type == "bireysel" {
		requireField(errors, "ad", input.Ad, "Ad zorunludur.")
		validateMaxLength(errors, "ad", input.Ad, "Ad")
		requireField(errors, "soyad", input.Soyad, "Soyad zorunludur.")
		validateMaxLength(errors, "soyad", input.Soyad, "Soyad")
		validatePhone(errors, "cep", input.Cep)
	}

	if input.Type == "kurumsal" {
		requireField(errors, "unvan", input.Unvan, "Ünvan zorunludur.")
		validateMaxLength(errors, "unvan", input.Unvan, "Ünvan")
		requireField(errors, "yetkili_adi", input.YetkiliAdi, "Yetkili adı zorunludur.")
		validateMaxLength(errors, "yetkili_adi", input.YetkiliAdi, "Yetkili adı")
		validatePhone(errors, "telefon", input.Telefon)
	}

	requireField(errors, "il_kodu", input.IlKodu, "İl zorunludur.")
	validateMaxLength(errors, "il_kodu", input.IlKodu, "İl")
	requireField(errors, "ilce_kodu", input.IlceKodu, "İlçe zorunludur.")
	validateMaxLength(errors, "ilce_kodu", input.IlceKodu, "İlçe")
	requireField(errors, "mahalle", input.Mahalle, "Mahalle zorunludur.")
	validateMaxLength(errors, "mahalle", input.Mahalle, "Mahalle")
	if input.BranchID <= 0 {
		errors["branch_id"] = "Bayi zorunludur."
	}

	return errors
}

func validateFullRegistrationInput(input domain.FullRegistrationInput) ValidationErrors {
	errors := ValidationErrors{}

	if input.Type != "bireysel" && input.Type != "kurumsal" {
		errors["type"] = "Müşteri türü bireysel veya kurumsal olmalıdır."
	}
	validatePhone(errors, "cep", input.Cep)
	requireField(errors, "ad", input.Ad, "Ad zorunludur.")
	validateMaxLength(errors, "ad", input.Ad, "Ad")
	requireField(errors, "soyad", input.Soyad, "Soyad zorunludur.")
	validateMaxLength(errors, "soyad", input.Soyad, "Soyad")

	if input.Type == "bireysel" {
		validateMaxLength(errors, "tc_no", input.TCNo, "T.C. no")
		validateDate(errors, "dogum_tarihi", input.DogumTarihi)
	}

	if input.Type == "kurumsal" {
		requireField(errors, "unvan", input.Unvan, "Ünvan zorunludur.")
		validateMaxLength(errors, "unvan", input.Unvan, "Ünvan")
		requireField(errors, "corporate_sector", input.CorporateSector, "Sektör zorunludur.")
		validateMaxLength(errors, "corporate_sector", input.CorporateSector, "Sektör")
		validateCorporateSector(errors, input.CorporateSector)
	}

	validateMaxLength(errors, "eposta", input.Eposta, "E-posta")
	validateEmail(errors, "eposta", input.Eposta)
	validateMaxLength(errors, "website", input.Website, "Website")
	validateMaxLength(errors, "google_map_link", input.GoogleMapLink, "Google map link")
	validateMaxLength(errors, "classifieds_website_link", input.ClassifiedsWebsiteLink, "İlan sitesi linki")
	if input.VehicleStockCount < 0 {
		errors["vehicle_stock_count"] = "Araç stok adedi 0 veya daha büyük olmalıdır."
	}
	if input.BranchID <= 0 {
		errors["branch_id"] = "Bayi zorunludur."
	}

	if input.Type == "kurumsal" {
		requireField(errors, "vergi_no", input.VergiNo, "Vergi no zorunludur.")
		validateMaxLength(errors, "vergi_no", input.VergiNo, "Vergi no")
		requireField(errors, "vergi_dairesi", input.VergiDairesi, "Vergi dairesi zorunludur.")
		validateMaxLength(errors, "vergi_dairesi", input.VergiDairesi, "Vergi dairesi")
	}

	for _, telephone := range input.Telephones {
		if strings.TrimSpace(telephone.PhoneNumber) == "" && strings.TrimSpace(telephone.Title) == "" {
			continue
		}
		validatePhone(errors, "telephones", telephone.PhoneNumber)
		validateMaxLength(errors, "telephones", telephone.Title, "Telefon başlığı")
	}
	requireField(errors, "il_kodu", input.IlKodu, "İl zorunludur.")
	validateMaxLength(errors, "il_kodu", input.IlKodu, "İl")
	validateMaxLength(errors, "ilce_kodu", input.IlceKodu, "İlçe")
	validateMaxLength(errors, "mahalle", input.Mahalle, "Mahalle")
	requireField(errors, "address_detail", input.AddressDetail, "Adres detayı zorunludur.")
	validateMaxLength(errors, "address_detail", input.AddressDetail, "Adres detayı")

	return errors
}

func validateSourceEditableFullRegistrationInput(input domain.FullRegistrationInput, customerType string) ValidationErrors {
	errors := ValidationErrors{}

	if strings.TrimSpace(customerType) == "kurumsal" {
		requireField(errors, "corporate_sector", input.CorporateSector, "Sektör zorunludur.")
	}
	validateMaxLength(errors, "corporate_sector", input.CorporateSector, "Sektör")
	validateCorporateSector(errors, input.CorporateSector)
	validateMaxLength(errors, "website", input.Website, "Website")
	validateMaxLength(errors, "google_map_link", input.GoogleMapLink, "Google map link")
	validateMaxLength(errors, "classifieds_website_link", input.ClassifiedsWebsiteLink, "İlan sitesi linki")
	if input.VehicleStockCount < 0 {
		errors["vehicle_stock_count"] = "Araç stok adedi 0 veya daha büyük olmalıdır."
	}

	return errors
}

func requireField(errors ValidationErrors, field string, value string, message string) {
	if strings.TrimSpace(value) == "" {
		errors[field] = message
	}
}

func validatePhone(errors ValidationErrors, field string, value string) {
	if !turkeyMobilePhonePattern.MatchString(value) {
		errors[field] = "Telefon 05XXXXXXXXX formatında, toplam 11 hane olmalıdır."
	}
}

func validateMaxLength(errors ValidationErrors, field string, value string, label string) {
	if len([]rune(strings.TrimSpace(value))) > customerTextMaxLength {
		errors[field] = label + " en fazla 255 karakter olabilir."
	}
}

func validateEmail(errors ValidationErrors, field string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	if _, err := mail.ParseAddress(value); err != nil {
		errors[field] = "Geçerli bir e-posta adresi giriniz."
	}
}

func validateDate(errors ValidationErrors, field string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	if _, err := time.Parse("2006-01-02", value); err != nil {
		errors[field] = "Tarih YYYY-AA-GG formatında olmalıdır."
	}
}

func validateCorporateSector(errors ValidationErrors, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	if _, ok := corporateSectorOptions[value]; !ok {
		errors["corporate_sector"] = "Geçerli bir sektör seçiniz."
	}
}

func phoneValueForCustomerType(input domain.CreateCustomerInput) string {
	if input.Type == "bireysel" {
		return input.Cep
	}

	return input.Telefon
}

func phoneFieldForCustomerType(customerType string) string {
	if customerType == "bireysel" {
		return "cep"
	}

	return "telefon"
}
