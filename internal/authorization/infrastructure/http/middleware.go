package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/auth/application"
	authzapp "github.com/umran/new.crm/backend/internal/authorization/application"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type TokenValidator interface {
	Validate(token string, expectedType string) (application.SessionTokenClaims, error)
}

type AuthMiddlewareConfig struct {
	AccessCookieName string
}

func RequirePermission(service *authzapp.Service, tokenValidator TokenValidator, config AuthMiddlewareConfig) fiber.Handler {
	if config.AccessCookieName == "" {
		config.AccessCookieName = "access_token"
	}

	return func(c *fiber.Ctx) error {
		token := c.Cookies(config.AccessCookieName)
		if token == "" {
			return response.Error(c, fiber.StatusUnauthorized, "Oturum geçersiz.", nil)
		}

		claims, err := tokenValidator.Validate(token, application.TokenTypeAccess)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Oturum geçersiz.", nil)
		}

		if claims.RoleID == 0 {
			return response.Error(c, fiber.StatusForbidden, "Yetkiniz bulunmuyor.", nil)
		}

		allowed, err := service.RoleHasAccess(c.UserContext(), claims.RoleID, c.Method(), normalizedRoutePath(c))
		if err != nil || !allowed {
			return response.Error(c, fiber.StatusForbidden, "Yetkiniz bulunmuyor.", nil)
		}

		c.Locals("claims", claims)

		return c.Next()
	}
}

func normalizedRoutePath(c *fiber.Ctx) string {
	path := c.Route().Path
	if path == "" {
		path = c.Path()
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
