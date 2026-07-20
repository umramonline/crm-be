package http

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	authapp "github.com/umran/new.crm/backend/internal/auth/application"
	"github.com/umran/new.crm/backend/internal/dashboard/application"
	"github.com/umran/new.crm/backend/internal/dashboard/domain"
	sharedauth "github.com/umran/new.crm/backend/internal/shared/auth"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Get("/dashboard", authRequired, h.GetDashboard)
}

func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	claims := c.Locals("claims").(authapp.SessionTokenClaims)

	filter, validationErrors, err := parseDashboardFilter(c, claims)
	if err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Dashboard filtreleri geçersiz.", validationErrors)
	}

	stats, validationErrors, err := h.service.GetDashboard(c.UserContext(), filter)
	if err != nil {
		if err == application.ErrInvalidDashboardFilter {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Dashboard filtreleri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Dashboard verileri şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Dashboard verileri getirildi.", stats)
}

func parseDashboardFilter(c *fiber.Ctx, claims authapp.SessionTokenClaims) (domain.Filter, application.ValidationErrors, error) {
	startDateRaw := strings.TrimSpace(c.Query("start_date"))
	endDateRaw := strings.TrimSpace(c.Query("end_date"))

	validationErrors := application.ValidationErrors{}
	if startDateRaw == "" {
		validationErrors["start_date"] = "Başlangıç tarihi zorunludur."
	}
	if endDateRaw == "" {
		validationErrors["end_date"] = "Bitiş tarihi zorunludur."
	}
	if len(validationErrors) > 0 {
		return domain.Filter{}, validationErrors, application.ErrInvalidDashboardFilter
	}

	startDate, startErr := time.Parse("2006-01-02", startDateRaw)
	if startErr != nil {
		validationErrors["start_date"] = "Tarih YYYY-AA-GG formatında olmalıdır."
	}

	endDate, endErr := time.Parse("2006-01-02", endDateRaw)
	if endErr != nil {
		validationErrors["end_date"] = "Tarih YYYY-AA-GG formatında olmalıdır."
	}

	if len(validationErrors) > 0 {
		return domain.Filter{}, validationErrors, application.ErrInvalidDashboardFilter
	}

	normalized := application.NormalizeFilterDates(startDate, endDate)
	filter := domain.Filter{
		StartDate:        normalized.StartDate,
		EndDate:          normalized.EndDate,
		BranchIDs:        claims.BranchIds,
		AllowAllBranches: sharedauth.IsAdminRole(claims.RoleID),
	}

	serviceValidationErrors := application.ValidationErrors{}
	if filter.EndDate.Before(filter.StartDate) {
		serviceValidationErrors["end_date"] = "Bitiş tarihi başlangıç tarihinden önce olamaz."
	}

	if len(serviceValidationErrors) > 0 {
		return domain.Filter{}, serviceValidationErrors, application.ErrInvalidDashboardFilter
	}

	return filter, nil, nil
}
