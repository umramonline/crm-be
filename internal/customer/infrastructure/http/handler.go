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
	router.Get("/zones", authRequired, h.ListZones)
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

func (h *Handler) ListZones(c *fiber.Ctx) error {
	zones, err := h.service.ListZones(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Bölge listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Bölgeler getirildi.", fiber.Map{"items": zones})
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
