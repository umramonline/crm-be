package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	authapp "github.com/umran/new.crm/backend/internal/auth/application"
	"github.com/umran/new.crm/backend/internal/customer/application"
	"github.com/umran/new.crm/backend/internal/customer/domain"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Get("/customers", authRequired, h.ListCustomers)
	router.Post("/customers", authRequired, h.CreateCustomer)
	router.Get("/customers/search", authRequired, h.SearchCustomer)
	router.Get("/customers/full-registration/:id/phone-exists", authRequired, h.FullRegistrationPhoneExists)
	router.Get("/customers/full-registration/:id", authRequired, h.GetFullRegistrationCustomer)
	router.Put("/customers/full-registration/:id", authRequired, h.CompleteFullRegistration)
	router.Get("/customers/backend/my-branches", authRequired, h.ListBackendCustomersMyBranches)
	router.Get("/customers/backend", authRequired, h.ListBackendCustomers)
	router.Get("/customers/backend/:id", authRequired, h.GetBackendCustomer)
	router.Get("/customers/umramonline/my-branches", authRequired, h.ListUmramonlineCustomersMyBranches)
	router.Get("/customers/umramonline", authRequired, h.ListUmramonlineCustomers)
	router.Get("/customers/umramonline/:id", authRequired, h.GetUmramonlineCustomer)
	router.Get("/customers/:id", authRequired, h.GetCustomer)
	router.Get("/zones", authRequired, h.ListZones)
	router.Get("/cities", authRequired, h.ListCities)
	router.Get("/towns", authRequired, h.ListTowns)
	router.Get("/branches", authRequired, h.ListBranches)
	router.Get("/branches/:id/users", authRequired, h.ListBranchUsers)
}

type createCustomerRequest struct {
	Type       string `json:"type"`
	Ad         string `json:"ad"`
	Soyad      string `json:"soyad"`
	Cep        string `json:"cep"`
	Unvan      string `json:"unvan"`
	YetkiliAdi string `json:"yetkili_adi"`
	Telefon    string `json:"telefon"`
	IlKodu     string `json:"il_kodu"`
	IlceKodu   string `json:"ilce_kodu"`
	Mahalle    string `json:"mahalle"`
	BranchID   int32  `json:"branch_id"`
}

type fullRegistrationTelephoneRequest struct {
	PhoneNumber string `json:"phone_number"`
	Title       string `json:"title"`
}

type fullRegistrationRequest struct {
	Type                   string                             `json:"type"`
	Cep                    string                             `json:"cep"`
	Ad                     string                             `json:"ad"`
	Soyad                  string                             `json:"soyad"`
	Unvan                  string                             `json:"unvan"`
	CorporateSector        string                             `json:"corporate_sector"`
	TCNo                   string                             `json:"tc_no"`
	DogumTarihi            string                             `json:"dogum_tarihi"`
	Eposta                 string                             `json:"eposta"`
	Website                string                             `json:"website"`
	GoogleMapLink          string                             `json:"google_map_link"`
	ClassifiedsWebsiteLink string                             `json:"classifieds_website_link"`
	VehicleStockCount      int32                              `json:"vehicle_stock_count"`
	BranchID               int32                              `json:"branch_id"`
	VergiNo                string                             `json:"vergi_no"`
	VergiDairesi           string                             `json:"vergi_dairesi"`
	Telephones             []fullRegistrationTelephoneRequest `json:"telephones"`
	IlKodu                 string                             `json:"il_kodu"`
	IlceKodu               string                             `json:"ilce_kodu"`
	Mahalle                string                             `json:"mahalle"`
	AddressDetail          string                             `json:"address_detail"`
}

func (h *Handler) ListCustomers(c *fiber.Ctx) error {
	return h.listCustomers(c, c.Query("data_source"), nil)
}

func (h *Handler) ListBackendCustomers(c *fiber.Ctx) error {
	return h.listCustomers(c, "backend", nil)
}

func (h *Handler) ListUmramonlineCustomers(c *fiber.Ctx) error {
	return h.listCustomers(c, "umramonline", nil)
}

func (h *Handler) ListBackendCustomersMyBranches(c *fiber.Ctx) error {
	return h.listCustomersScopedToClaimsBranches(c, "backend")
}

func (h *Handler) ListUmramonlineCustomersMyBranches(c *fiber.Ctx) error {
	return h.listCustomersScopedToClaimsBranches(c, "umramonline")
}

func (h *Handler) listCustomersScopedToClaimsBranches(c *fiber.Ctx, dataSource string) error {
	claims := c.Locals("claims").(authapp.SessionTokenClaims)
	branchIDs := uint64SliceToInt32(claims.BranchIds)
	if len(branchIDs) == 0 {
		page := queryInt(c, "page", 1)
		perPage := queryInt(c, "per_page", 10)
		return response.Success(c, fiber.StatusOK, "Müşteriler getirildi.", domain.ListResult{
			Items: []domain.Customer{},
			Pagination: domain.Pagination{
				CurrentPage: page,
				LastPage:    1,
				PerPage:     perPage,
				Total:       0,
			},
		})
	}

	return h.listCustomers(c, dataSource, branchIDs)
}

func (h *Handler) listCustomers(c *fiber.Ctx, dataSource string, branchIDs []int32) error {
	query := domain.ListQuery{
		Page:       queryInt(c, "page", 1),
		PerPage:    queryInt(c, "per_page", 10),
		DataSource: dataSource,
		Situation:  c.Query("situation"),
		Unvan:      c.Query("unvan"),
		Cep:        c.Query("cep"),
		Ad:         c.Query("ad"),
		Soyad:      c.Query("soyad"),
		BranchName: c.Query("branch_name"),
		PlusCardNo: firstNonEmpty(c.Query("plus_card_no"), c.Query("no")),
		Source:     c.Query("source"),
		City:       firstNonEmpty(c.Query("city"), c.Query("title")),
		Town:       firstNonEmpty(c.Query("town"), c.Query("ilce_title")),
		CreatedAt:  c.Query("created_at"),
		Type:       c.Query("type"),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
		ZoneID:     queryInt(c, "zone_id", 0),
		BranchIDs:  branchIDs,
	}

	result, err := h.service.ListCustomers(c.UserContext(), query)
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Müşteriler getirildi.", result)
}

func uint64SliceToInt32(values []uint64) []int32 {
	if len(values) == 0 {
		return nil
	}

	result := make([]int32, 0, len(values))
	for _, value := range values {
		if value == 0 || value > uint64(^uint32(0)>>1) {
			continue
		}
		result = append(result, int32(value))
	}

	return result
}

func (h *Handler) GetCustomer(c *fiber.Ctx) error {
	customer, err := h.service.GetCustomer(c.UserContext(), paramUint64(c, "id", 0), c.Query("data_source"))
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri detayı şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Müşteri detayı getirildi.", customer)
}

func (h *Handler) GetFullRegistrationCustomer(c *fiber.Ctx) error {
	customer, err := h.service.GetFullRegistrationCustomer(c.UserContext(), paramUint64(c, "id", 0))
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Tam kayıt bilgileri şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Tam kayıt bilgileri getirildi.", customer)
}

func (h *Handler) FullRegistrationPhoneExists(c *fiber.Ctx) error {
	exists, err := h.service.FullRegistrationPhoneExists(c.UserContext(), paramUint64(c, "id", 0), c.Query("cep"))
	if err != nil {
		if err == application.ErrInvalidCustomerCreateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Cep telefonu geçersiz.", fiber.Map{
				"cep": "Telefon 05XXXXXXXXX formatında, toplam 11 hane olmalıdır.",
			})
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Cep telefonu kontrolü şu anda yapılamıyor.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Cep telefonu kontrolü tamamlandı.", fiber.Map{
		"exists": exists,
	})
}

func (h *Handler) CompleteFullRegistration(c *fiber.Ctx) error {
	var request fullRegistrationRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Tam kayıt bilgileri geçersiz.", fiber.Map{
			"request": "Tam kayıt bilgileri geçersiz.",
		})
	}

	telephones := make([]domain.CustomerTelephone, 0, len(request.Telephones))
	for _, telephone := range request.Telephones {
		telephones = append(telephones, domain.CustomerTelephone{
			PhoneNumber: telephone.PhoneNumber,
			Title:       telephone.Title,
		})
	}

	customer, validationErrors, err := h.service.CompleteFullRegistration(c.UserContext(), paramUint64(c, "id", 0), domain.FullRegistrationInput{
		Type:                   request.Type,
		Cep:                    request.Cep,
		Ad:                     request.Ad,
		Soyad:                  request.Soyad,
		Unvan:                  request.Unvan,
		CorporateSector:        request.CorporateSector,
		TCNo:                   request.TCNo,
		DogumTarihi:            request.DogumTarihi,
		Eposta:                 request.Eposta,
		Website:                request.Website,
		GoogleMapLink:          request.GoogleMapLink,
		ClassifiedsWebsiteLink: request.ClassifiedsWebsiteLink,
		VehicleStockCount:      request.VehicleStockCount,
		BranchID:               request.BranchID,
		VergiNo:                request.VergiNo,
		VergiDairesi:           request.VergiDairesi,
		Telephones:             telephones,
		IlKodu:                 request.IlKodu,
		IlceKodu:               request.IlceKodu,
		Mahalle:                request.Mahalle,
		AddressDetail:          request.AddressDetail,
	})
	if err != nil {
		if err == application.ErrInvalidCustomerCreateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Tam kayıt bilgileri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Tam kayıt şu anda tamamlanamadı.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Tam kayıt tamamlandı.", customer)
}

func (h *Handler) GetBackendCustomer(c *fiber.Ctx) error {
	customer, err := h.service.GetCustomer(c.UserContext(), paramUint64(c, "id", 0), "backend")
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri detayı şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Müşteri detayı getirildi.", customer)
}

func (h *Handler) GetUmramonlineCustomer(c *fiber.Ctx) error {
	customer, err := h.service.GetCustomer(c.UserContext(), paramUint64(c, "id", 0), "umramonline")
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri detayı şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Müşteri detayı getirildi.", customer)
}

func (h *Handler) CreateCustomer(c *fiber.Ctx) error {
	var request createCustomerRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Müşteri bilgileri geçersiz.", fiber.Map{
			"request": "Müşteri bilgileri geçersiz.",
		})
	}

	customer, validationErrors, err := h.service.CreateCustomer(c.UserContext(), domain.CreateCustomerInput{
		Type:       request.Type,
		Ad:         request.Ad,
		Soyad:      request.Soyad,
		Cep:        request.Cep,
		Unvan:      request.Unvan,
		YetkiliAdi: request.YetkiliAdi,
		Telefon:    request.Telefon,
		IlKodu:     request.IlKodu,
		IlceKodu:   request.IlceKodu,
		Mahalle:    request.Mahalle,
		BranchID:   request.BranchID,
	})
	if err != nil {
		if err == application.ErrInvalidCustomerCreateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Müşteri bilgileri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri kaydı şu anda oluşturulamadı.", nil)
	}

	return response.Success(c, fiber.StatusCreated, "Müşteri kaydedildi.", customer)
}

func (h *Handler) SearchCustomer(c *fiber.Ctx) error {
	result, err := h.service.SearchCustomer(c.UserContext(), c.Query("q"))
	if err != nil {
		if err == application.ErrInvalidCustomerSearchQuery {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Arama metni zorunludur.", fiber.Map{
				"q": "Arama metni zorunludur.",
			})
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri araması şu anda yapılamıyor.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Müşteri araması tamamlandı.", result)
}

func (h *Handler) ListZones(c *fiber.Ctx) error {
	zones, err := h.service.ListZones(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Bölge listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Bölgeler getirildi.", fiber.Map{"items": zones})
}

func (h *Handler) ListCities(c *fiber.Ctx) error {
	cities, err := h.service.ListCities(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Şehir listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Şehirler getirildi.", fiber.Map{"items": cities})
}

func (h *Handler) ListTowns(c *fiber.Ctx) error {
	towns, err := h.service.ListTowns(c.UserContext(), queryUint64(c, "city_id", 0))
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "İlçe listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "İlçeler getirildi.", fiber.Map{"items": towns})
}

func (h *Handler) ListBranches(c *fiber.Ctx) error {
	branches, err := h.service.ListBranches(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Bayi listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Bayiler getirildi.", fiber.Map{"items": branches})
}

func (h *Handler) ListBranchUsers(c *fiber.Ctx) error {
	users, err := h.service.ListBranchUsers(c.UserContext(), paramUint64(c, "id", 0))
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Bayi kullanıcıları şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Bayi kullanıcıları getirildi.", fiber.Map{"items": users})
}

func queryInt(c *fiber.Ctx, name string, defaultValue int) int {
	value := c.Query(name)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if name == "zone_id" {
		if parsed < 0 {
			return defaultValue
		}

		return parsed
	}

	if parsed <= 0 {
		return defaultValue
	}

	return parsed
}

func queryUint64(c *fiber.Ctx, name string, defaultValue uint64) uint64 {
	value := c.Query(name)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func paramUint64(c *fiber.Ctx, name string, defaultValue uint64) uint64 {
	value := c.Params(name)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
