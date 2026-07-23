package http

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/consume/application"
	"github.com/umran/new.crm/backend/internal/consume/domain"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, apiKeyRequired fiber.Handler) {
	router.Post("/consume", apiKeyRequired, h.Consume)
}

type consumeEnvelope struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
}

func (h *Handler) Consume(c *fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", fiber.Map{
			"body": "JSON parse edilemedi.",
		})
	}

	var envelope consumeEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", fiber.Map{
			"body": "JSON parse edilemedi.",
		})
	}

	result, err := h.service.Consume(c.UserContext(), domain.ConsumeCommand{
		EventID:   envelope.EventID,
		EventType: envelope.EventType,
		Payload:   append([]byte(nil), body...),
	})
	if err != nil {
		switch err {
		case application.ErrInvalidEventPayload:
			return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz event payload.", fiber.Map{
				"event_id":   "event_id zorunludur.",
				"uo_id":      "uo_id zorunludur.",
				"event_type": "event_type zorunludur.",
			})
		case application.ErrUnsupportedEventType:
			return response.Error(c, fiber.StatusUnprocessableEntity, "Desteklenmeyen event_type.", fiber.Map{
				"event_type": "Şu an yalnızca customer.created desteklenmektedir.",
			})
		default:
			return response.Error(c, fiber.StatusInternalServerError, "Event işlenemedi.", nil)
		}
	}

	status := fiber.StatusOK
	message := "Event consumed."
	if result.Action == "created" {
		status = fiber.StatusCreated
		message = "Customer created."
	} else if result.Action == "updated" {
		message = "Customer updated."
	} else if result.Action == "already_processed" {
		message = "Event already processed."
	}

	return response.Success(c, status, message, result)
}
