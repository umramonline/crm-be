package http

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/auth/application"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type OTPRequestService interface {
	RequestOTP(ctx context.Context, phone string) error
}

type OTPHandler struct {
	service OTPRequestService
}

type otpRequest struct {
	Phone string `json:"phone"`
}

func NewOTPHandler(service OTPRequestService) *OTPHandler {
	return &OTPHandler{service: service}
}

func (h *OTPHandler) RequestOTP(c *fiber.Ctx) error {
	var request otpRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", map[string]string{
			"body": "JSON formatı geçersiz.",
		})
	}

	if err := h.service.RequestOTP(c.UserContext(), request.Phone); err != nil {
		if errors.Is(err, application.ErrInvalidPhone) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Doğrulama hatası.", map[string]string{
				"phone": "Telefon numarası 05XXXXXXXXX formatında olmalıdır.",
			})
		}

		return response.Error(c, fiber.StatusInternalServerError, "OTP isteği şu anda tamamlanamadı.", nil)
	}

	return response.Success(c, fiber.StatusOK, "OTP kodu gönderildi.", fiber.Map{})
}
