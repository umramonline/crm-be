package http

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	authApp "github.com/umran/new.crm/backend/internal/auth/application"
	"github.com/umran/new.crm/backend/internal/shared/response"
	"github.com/umran/new.crm/backend/internal/task/application"
	"github.com/umran/new.crm/backend/internal/task/domain"
)

type Handler struct {
	service     *application.Service
	smsNotifier TaskCreatedSMSNotifier
	logger      *log.Logger
}

type TaskCreatedSMSNotifier interface {
	SendTaskCreatedSMS(ctx context.Context, input domain.TaskCreatedSMSInput) error
}

type createTaskRequest struct {
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	AssignedUserID       uint64   `json:"assigned_user_id"`
	AssignedUserFullName string   `json:"assigned_user_full_name"`
	BranchID             uint64   `json:"branch_id"`
	BranchName           string   `json:"branch_name"`
	VisitDate            string   `json:"visit_date"`
	DueDate              string   `json:"due_date"`
	Priority             string   `json:"priority"`
	CustomerIDs          []uint64 `json:"customer_ids"`
}

func NewHandler(service *application.Service, smsNotifiers ...TaskCreatedSMSNotifier) *Handler {
	var smsNotifier TaskCreatedSMSNotifier
	if len(smsNotifiers) > 0 {
		smsNotifier = smsNotifiers[0]
	}

	return &Handler{
		service:     service,
		smsNotifier: smsNotifier,
		logger:      log.Default(),
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Get("/tasks", authRequired, h.ListTasks)
	router.Get("/tasks/:uuid", authRequired, h.GetTask)
	router.Patch("/tasks/:uuid/cancel", authRequired, h.CancelTask)
	router.Post("/tasks", authRequired, h.CreateTask)
}

func (h *Handler) ListTasks(c *fiber.Ctx) error {
	result, err := h.service.ListTasks(c.UserContext(), domain.ListQuery{
		Page:                  queryInt(c, "page", 1),
		PerPage:               queryInt(c, "per_page", 10),
		Title:                 c.Query("title"),
		Customer:              c.Query("customer"),
		AssignedUserFullName:  c.Query("assigned_user_full_name"),
		BranchName:            c.Query("branch_name"),
		VisitDate:             c.Query("visit_date"),
		DueDate:               c.Query("due_date"),
		Priority:              c.Query("priority"),
		Status:                c.Query("status"),
		CreatedByUserFullName: c.Query("created_by_user_full_name"),
		SortBy:                c.Query("sort_by"),
		SortOrder:             c.Query("sort_order"),
	})
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Görev listesi şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Görevler getirildi.", result)
}

func (h *Handler) GetTask(c *fiber.Ctx) error {
	task, err := h.service.GetTask(c.UserContext(), c.Params("uuid"), queryUint64(c, "customer_id"))
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Görev detayı şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Görev detayı getirildi.", task)
}

func (h *Handler) CancelTask(c *fiber.Ctx) error {
	customerID := queryUint64(c, "customer_id")
	if customerID == 0 {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Müşteri bilgisi zorunludur.", fiber.Map{
			"customer_id": "Müşteri bilgisi zorunludur.",
		})
	}

	task, err := h.service.CancelTask(c.UserContext(), c.Params("uuid"), customerID)
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Görev şu anda iptal edilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Görev iptal edildi.", task)
}

func (h *Handler) CreateTask(c *fiber.Ctx) error {
	var request createTaskRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Görev bilgileri geçersiz.", fiber.Map{
			"request": "Görev bilgileri geçersiz.",
		})
	}

	claims := c.Locals("claims").(authApp.SessionTokenClaims)

	task, validationErrors, err := h.service.CreateTask(c.UserContext(), domain.CreateTaskInput{
		Title:                 request.Title,
		Description:           request.Description,
		AssignedUserID:        request.AssignedUserID,
		AssignedUserFullName:  request.AssignedUserFullName,
		CreatedByUserID:       claims.UserId,
		CreatedByUserFullName: claims.UserFullName,
		BranchID:              request.BranchID,
		BranchName:            request.BranchName,
		VisitDate:             request.VisitDate,
		DueDate:               request.DueDate,
		Priority:              request.Priority,
		CustomerIDs:           request.CustomerIDs,
	})
	if err != nil {
		if err == application.ErrInvalidTaskCreateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Görev bilgileri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Görev kaydı şu anda oluşturulamadı.", nil)
	}

	h.dispatchTaskCreatedSMS(task)

	return response.Success(c, fiber.StatusCreated, "Görev kaydedildi.", task)
}

func (h *Handler) dispatchTaskCreatedSMS(task domain.Task) {
	if h == nil || h.smsNotifier == nil || strings.TrimSpace(task.AssignedUserPhone) == "" {
		return
	}

	input := domain.TaskCreatedSMSInput{
		Phone:                task.AssignedUserPhone,
		TaskUUID:             task.UUID,
		Title:                task.Title,
		AssignedUserFullName: task.AssignedUserFullName,
		BranchName:           task.BranchName,
		VisitDate:            task.VisitDate,
		DueDate:              task.DueDate,
		Priority:             task.Priority,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := h.smsNotifier.SendTaskCreatedSMS(ctx, input); err != nil && h.logger != nil {
			h.logger.Printf("task created SMS failed task_uuid=%s phone=%s error=%v", input.TaskUUID, maskPhone(input.Phone), err)
		}
	}()
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

func queryUint64(c *fiber.Ctx, key string) uint64 {
	value := c.Query(key)
	if value == "" {
		return 0
	}

	parsedValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}

	return parsedValue
}

func maskPhone(phone string) string {
	trimmedPhone := strings.TrimSpace(phone)
	if len(trimmedPhone) <= 4 {
		return "****"
	}

	return trimmedPhone[:2] + "*****" + trimmedPhone[len(trimmedPhone)-2:]
}
