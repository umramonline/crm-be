package application

import (
	"context"
	"errors"
	"regexp"
	"strings"

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

type CustomerProvider interface {
	ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	ListZones(ctx context.Context) ([]domain.Zone, error)
	SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error)
	GetCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	ListCities(ctx context.Context) ([]domain.City, error)
	ListTowns(ctx context.Context, cityID uint64) ([]domain.Town, error)
	ListBranches(ctx context.Context) ([]domain.Branch, error)
}

type CustomerRepository interface {
	SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	CreateCustomer(ctx context.Context, input domain.CreateCustomerInput) (domain.CustomerDetail, error)
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
	if s == nil || s.provider == nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	return s.provider.ListCustomers(ctx, query)
}

func (s *Service) ListZones(ctx context.Context) ([]domain.Zone, error) {
	if s == nil || s.provider == nil {
		return nil, ErrZoneListUnavailable
	}

	return s.provider.ListZones(ctx)
}

func (s *Service) GetCustomer(ctx context.Context, id uint64) (domain.CustomerDetail, error) {
	if s == nil || s.provider == nil || id == 0 {
		return domain.CustomerDetail{}, ErrCustomerSearchUnavailable
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

	// ilk başta veritabanında arama yapıyoruz
	customer, found, err := s.repository.SearchCustomer(ctx, normalizedQuery)
	if err != nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	if found {
		return domain.CustomerSearchResult{
			Found:    true,
			Source:   "backend",
			Customer: &customer,
		}, nil
	}

	// eğer veritabanında bulunamadıysa, umramonline'dan arama yapıyoruz
	customer, found, err = s.provider.SearchCustomer(ctx, normalizedQuery)
	if err != nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	if !found {
		return domain.CustomerSearchResult{Found: false}, nil
	}

	return domain.CustomerSearchResult{
		Found:    true,
		Source:   "umramonline",
		Customer: &customer,
	}, nil
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

func (s *Service) ListBranches(ctx context.Context) ([]domain.Branch, error) {
	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListBranches(ctx)
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
