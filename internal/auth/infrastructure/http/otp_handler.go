package http

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/auth/application"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type OTPRequestService interface {
	RequestOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone string, otpCode string) error
	LoginWithPassword(ctx context.Context, phone string, password string) (map[string]any, error)
}

type SessionTokenService interface {
	Issue(subject string, tokenType string, ttl time.Duration) (string, error)
	Validate(token string, expectedType string) (application.SessionTokenClaims, error)
}

type SessionConfig struct {
	AccessCookieName  string
	RefreshCookieName string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	CookieSecure      bool
	CookieSameSite    string
}

type OTPHandler struct {
	service       OTPRequestService
	tokenService  SessionTokenService
	sessionConfig SessionConfig
}

type otpRequest struct {
	Phone string `json:"phone"`
}

type otpVerifyRequest struct {
	Phone   string `json:"phone"`
	OTPCode string `json:"otp_code"`
}

type passwordLoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func NewOTPHandler(service OTPRequestService, tokenService SessionTokenService, sessionConfig SessionConfig) *OTPHandler {
	if sessionConfig.AccessCookieName == "" {
		sessionConfig.AccessCookieName = "access_token"
	}

	if sessionConfig.RefreshCookieName == "" {
		sessionConfig.RefreshCookieName = "refresh_token"
	}

	if sessionConfig.AccessTTL <= 0 {
		sessionConfig.AccessTTL = 15 * time.Minute
	}

	if sessionConfig.RefreshTTL <= 0 {
		sessionConfig.RefreshTTL = 30 * 24 * time.Hour
	}

	if sessionConfig.CookieSameSite == "" {
		sessionConfig.CookieSameSite = "Lax"
	}

	return &OTPHandler{service: service, tokenService: tokenService, sessionConfig: sessionConfig}
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

func (h *OTPHandler) VerifyOTP(c *fiber.Ctx) error {
	var request otpVerifyRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", map[string]string{
			"body": "JSON formatı geçersiz.",
		})
	}

	if err := h.service.VerifyOTP(c.UserContext(), request.Phone, request.OTPCode); err != nil {
		if errors.Is(err, application.ErrInvalidPhone) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Doğrulama hatası.", map[string]string{
				"phone": "Telefon numarası 05XXXXXXXXX formatında olmalıdır.",
			})
		}

		if errors.Is(err, application.ErrInvalidOTPCode) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Doğrulama hatası.", map[string]string{
				"otp_code": "OTP kodu 6 haneli olmalıdır.",
			})
		}

		if errors.Is(err, application.ErrOTPVerifyRejected) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Güvenlik kodu hatalı.", nil)
		}

		return response.Error(c, fiber.StatusInternalServerError, "OTP doğrulama şu anda tamamlanamadı.", nil)
	}

	return response.Success(c, fiber.StatusOK, "OTP doğrulandı.", fiber.Map{})
}

func (h *OTPHandler) LoginWithPassword(c *fiber.Ctx) error {
	var request passwordLoginRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", map[string]string{
			"body": "JSON formatı geçersiz.",
		})
	}

	data, err := h.service.LoginWithPassword(c.UserContext(), request.Phone, request.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidPhone) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Doğrulama hatası.", map[string]string{
				"phone": "Telefon numarası 05XXXXXXXXX formatında olmalıdır.",
			})
		}

		if errors.Is(err, application.ErrInvalidPassword) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Doğrulama hatası.", map[string]string{
				"password": "Şifre zorunludur.",
			})
		}

		if errors.Is(err, application.ErrPasswordRejected) {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Kimlik bilgileri hatalı.", nil)
		}

		return response.Error(c, fiber.StatusInternalServerError, "Giriş işlemi şu anda tamamlanamadı.", nil)
	}

	userID, err := extractUserID(data)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Giriş işlemi şu anda tamamlanamadı.", nil)
	}

	if err := h.setSessionCookies(c, userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Giriş işlemi şu anda tamamlanamadı.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Giriş başarılı.", data)
}

func (h *OTPHandler) RefreshSession(c *fiber.Ctx) error {
	refreshToken := c.Cookies(h.sessionConfig.RefreshCookieName)
	if refreshToken == "" {
		return response.Error(c, fiber.StatusUnauthorized, "Oturum geçersiz.", nil)
	}

	claims, err := h.tokenService.Validate(refreshToken, application.TokenTypeRefresh)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Oturum geçersiz.", nil)
	}

	accessToken, err := h.tokenService.Issue(claims.Subject, application.TokenTypeAccess, h.sessionConfig.AccessTTL)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Oturum yenilenemedi.", nil)
	}

	h.setCookie(c, h.sessionConfig.AccessCookieName, accessToken, h.sessionConfig.AccessTTL, "/")

	return response.Success(c, fiber.StatusOK, "Oturum yenilendi.", fiber.Map{})
}

func (h *OTPHandler) Logout(c *fiber.Ctx) error {
	h.clearCookie(c, h.sessionConfig.AccessCookieName, "/")
	h.clearCookie(c, h.sessionConfig.RefreshCookieName, "/api/v1/auth/refresh")

	return response.Success(c, fiber.StatusOK, "Çıkış yapıldı.", fiber.Map{})
}

func (h *OTPHandler) Session(c *fiber.Ctx) error {
	accessToken := c.Cookies(h.sessionConfig.AccessCookieName)
	if accessToken == "" {
		return response.Error(c, fiber.StatusUnauthorized, "Oturum geçersiz.", nil)
	}

	claims, err := h.tokenService.Validate(accessToken, application.TokenTypeAccess)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Oturum geçersiz.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Oturum geçerli.", fiber.Map{
		"user_id": claims.Subject,
	})
}

func (h *OTPHandler) setSessionCookies(c *fiber.Ctx, userID string) error {
	accessToken, err := h.tokenService.Issue(userID, application.TokenTypeAccess, h.sessionConfig.AccessTTL)
	if err != nil {
		return err
	}

	refreshToken, err := h.tokenService.Issue(userID, application.TokenTypeRefresh, h.sessionConfig.RefreshTTL)
	if err != nil {
		return err
	}

	h.setCookie(c, h.sessionConfig.AccessCookieName, accessToken, h.sessionConfig.AccessTTL, "/")
	h.setCookie(c, h.sessionConfig.RefreshCookieName, refreshToken, h.sessionConfig.RefreshTTL, "/api/v1/auth/refresh")

	return nil
}

func (h *OTPHandler) setCookie(c *fiber.Ctx, name string, value string, ttl time.Duration, path string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   int(ttl.Seconds()),
		Secure:   h.sessionConfig.CookieSecure,
		HTTPOnly: true,
		SameSite: h.sessionConfig.CookieSameSite,
	})
}

func (h *OTPHandler) clearCookie(c *fiber.Ctx, name string, path string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		Secure:   h.sessionConfig.CookieSecure,
		HTTPOnly: true,
		SameSite: h.sessionConfig.CookieSameSite,
	})
}

func extractUserID(data map[string]any) (string, error) {
	user, ok := data["user"].(map[string]any)
	if !ok {
		return "", errors.New("missing user data")
	}

	switch id := user["id"].(type) {
	case float64:
		return strconv.FormatInt(int64(id), 10), nil
	case int:
		return strconv.Itoa(id), nil
	case string:
		if id == "" {
			return "", errors.New("empty user id")
		}

		return id, nil
	default:
		return "", fmt.Errorf("unsupported user id type: %T", id)
	}
}
