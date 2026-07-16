package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/ietts/application"
	"github.com/umran/new.crm/backend/internal/ietts/domain"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Get("/ietts", authRequired, h.ListRecords)
}

func (h *Handler) ListRecords(c *fiber.Ctx) error {
	result, err := h.service.ListRecords(c.UserContext(), domain.ListQuery{
		Page:              queryInt(c, "page", 1),
		PerPage:           queryInt(c, "per_page", 20),
		DocumentNumber:    c.Query("document_number"),
		CompanyName:       c.Query("company_name"),
		BusinessName:      c.Query("business_name"),
		BusinessAddress:   c.Query("business_address"),
		DocumentIssueDate: c.Query("document_issue_date"),
		DocumentStatus:    c.Query("document_status"),
		City:              c.Query("city"),
		District:          c.Query("district"),
		CreatedAt:         c.Query("created_at"),
		SortBy:            c.Query("sort_by"),
		SortOrder:         c.Query("sort_order"),
	})
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "IETTS kayıtları şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "IETTS kayıtları getirildi.", result)
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
