package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

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
	router.Get("/customers/:id", authRequired, h.GetCustomer)
	router.Get("/zones", authRequired, h.ListZones)
	router.Get("/cities", authRequired, h.ListCities)
	router.Get("/towns", authRequired, h.ListTowns)
	router.Get("/branches", authRequired, h.ListBranches)
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

func (h *Handler) ListCustomers(c *fiber.Ctx) error {
	query := domain.ListQuery{
		Page:       queryInt(c, "page", 1),
		PerPage:    queryInt(c, "per_page", 10),
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
	}

	result, err := h.service.ListCustomers(c.UserContext(), query)
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Müşteri listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Müşteriler getirildi.", result)
}

func (h *Handler) GetCustomer(c *fiber.Ctx) error {
	customer, err := h.service.GetCustomer(c.UserContext(), paramUint64(c, "id", 0))
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
