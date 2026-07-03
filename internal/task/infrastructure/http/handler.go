package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/shared/response"
	"github.com/umran/new.crm/backend/internal/task/application"
	"github.com/umran/new.crm/backend/internal/task/domain"
)

type Handler struct {
	service *application.Service
}

type createTaskRequest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	AssignedUserID uint64   `json:"assigned_user_id"`
	BranchID       uint64   `json:"branch_id"`
	VisitDate      string   `json:"visit_date"`
	DueDate        string   `json:"due_date"`
	Priority       string   `json:"priority"`
	CustomerIDs    []uint64 `json:"customer_ids"`
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Post("/tasks", authRequired, h.CreateTask)
}

func (h *Handler) CreateTask(c *fiber.Ctx) error {
	var request createTaskRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Görev bilgileri geçersiz.", fiber.Map{
			"request": "Görev bilgileri geçersiz.",
		})
	}

	task, validationErrors, err := h.service.CreateTask(c.UserContext(), domain.CreateTaskInput{
		Title:          request.Title,
		Description:    request.Description,
		AssignedUserID: request.AssignedUserID,
		BranchID:       request.BranchID,
		VisitDate:      request.VisitDate,
		DueDate:        request.DueDate,
		Priority:       request.Priority,
		CustomerIDs:    request.CustomerIDs,
	})
	if err != nil {
		if err == application.ErrInvalidTaskCreateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Görev bilgileri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Görev kaydı şu anda oluşturulamadı.", nil)
	}

	return response.Success(c, fiber.StatusCreated, "Görev kaydedildi.", task)
}
