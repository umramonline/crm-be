package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/shared/response"
)

func RequireAPIKey(expectedKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if strings.TrimSpace(expectedKey) == "" {
			return response.Error(c, fiber.StatusServiceUnavailable, "Consume API anahtarı yapılandırılmamış.", nil)
		}

		authorization := strings.TrimSpace(c.Get("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") {
			return response.Error(c, fiber.StatusUnauthorized, "Yetkilendirme gerekli.", nil)
		}

		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token == "" || token != expectedKey {
			return response.Error(c, fiber.StatusUnauthorized, "Geçersiz token.", nil)
		}

		return c.Next()
	}
}
