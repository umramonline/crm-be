package http

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/auth/application"
	branchapp "github.com/umran/new.crm/backend/internal/authorization/application"
	sharedauth "github.com/umran/new.crm/backend/internal/shared/auth"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type OTPRequestService interface {
	RequestOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone string, otpCode string) error
	LoginWithPassword(ctx context.Context, phone string, password string) (map[string]any, error)
}

type SessionTokenService interface {
	Issue(userId uint64, tokenType string, ttl time.Duration, roleID uint64, roleName string, name string, branches []branchapp.Branch) (string, error)
	Validate(token string, expectedType string) (application.SessionTokenClaims, error)
}

type Permission struct {
	ModuleID       uint64 `json:"module_id"`
	ModuleName     string `json:"module_name"`
	ModuleMethodID uint64 `json:"module_method_id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Method         string `json:"method,omitempty"`
	Path           string `json:"path,omitempty"`
}

type SessionUser struct {
	ID        uint64             `json:"id"`
	FullName  string             `json:"full_name,omitempty"`
	Phone     string             `json:"phone,omitempty"`
	RoleID    uint64             `json:"role_id"`
	RoleName  string             `json:"role_name,omitempty"`
	BranchIds []uint64           `json:"branch_ids,omitempty"`
	Branches  []branchapp.Branch `json:"branches,omitempty"`
}

type SessionData struct {
	UserID      uint64       `json:"user_id"`
	User        SessionUser  `json:"user"`
	Permissions []Permission `json:"permissions"`
}

type AuthorizationService interface {
	SessionFromLoginData(ctx context.Context, data map[string]any) (SessionData, error)
	SessionForUser(ctx context.Context, user SessionUser) (SessionData, error)
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
	authorization AuthorizationService
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

func (h *OTPHandler) SetAuthorizationService(service AuthorizationService) {
	h.authorization = service
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

	sessionData, err := h.sessionDataFromLoginData(c.UserContext(), data)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Giriş işlemi şu anda tamamlanamadı.", nil)
	}

	if err := h.setSessionCookies(c, sessionData.User); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Giriş işlemi şu anda tamamlanamadı.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Giriş başarılı.", sessionData)
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

	accessToken, err := h.tokenService.Issue(claims.UserId, application.TokenTypeAccess, h.sessionConfig.AccessTTL, claims.RoleID, claims.RoleName, claims.UserFullName, claims.Branches)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Oturum yenilenemedi.", nil)
	}

	h.setCookie(c, h.sessionConfig.AccessCookieName, accessToken, h.sessionConfig.AccessTTL, "/")

	sessionData, err := h.sessionDataFromClaims(c.UserContext(), claims)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Oturum yenilenemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Oturum yenilendi.", sessionData)
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

	sessionData, err := h.sessionDataFromClaims(c.UserContext(), claims)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Oturum bilgisi getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Oturum geçerli.", sessionData)
}

func (h *OTPHandler) setSessionCookies(c *fiber.Ctx, user SessionUser) error {
	userID := user.ID
	accessToken, err := h.tokenService.Issue(userID, application.TokenTypeAccess, h.sessionConfig.AccessTTL, user.RoleID, user.RoleName, user.FullName, user.Branches)
	if err != nil {
		return err
	}

	refreshToken, err := h.tokenService.Issue(userID, application.TokenTypeRefresh, h.sessionConfig.RefreshTTL, user.RoleID, user.RoleName, user.FullName, user.Branches)
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

func extractUserID(data map[string]any) (uint64, error) {
	user, ok := data["user"].(map[string]any)
	if !ok {
		return 0, errors.New("missing user data")
	}

	switch id := user["id"].(type) {
	case float64:
		return uint64(id), nil
	case int:
		return uint64(id), nil
	case string:
		if id == "" {
			return 0, errors.New("empty user id")
		}

		return strconv.ParseUint(id, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported user id type: %T", id)
	}
}

func (h *OTPHandler) sessionDataFromLoginData(ctx context.Context, data map[string]any) (SessionData, error) {
	if h.authorization == nil {
		return fallbackSessionDataFromLoginData(data)
	}

	return h.authorization.SessionFromLoginData(ctx, data)
}

func (h *OTPHandler) sessionDataFromClaims(ctx context.Context, claims application.SessionTokenClaims) (SessionData, error) {
	user := SessionUser{
		ID:       claims.UserId,
		FullName: claims.UserFullName,
		RoleID:   claims.RoleID,
		RoleName: claims.RoleName,
	}

	if !sharedauth.IsAdminRole(claims.RoleID) {
		user.BranchIds = claims.BranchIds
		user.Branches = claims.Branches
	}

	if h.authorization == nil {
		return SessionData{
			UserID:      claims.UserId,
			User:        user,
			Permissions: []Permission{},
		}, nil
	}

	return h.authorization.SessionForUser(ctx, user)
}

func fallbackSessionDataFromLoginData(data map[string]any) (SessionData, error) {
	userID, err := extractUserID(data)
	if err != nil {
		return SessionData{}, err
	}

	rawUser, _ := data["user"].(map[string]any)
	roleID, _ := uintFromAny(rawUser["role_id"])

	return SessionData{
		UserID: uint64(userID),
		User: SessionUser{
			ID:       uint64(userID),
			FullName: stringFromAny(rawUser["name"]),
			Phone:    stringFromAny(rawUser["phone"]),
			RoleID:   roleID,
			RoleName: stringFromAny(rawUser["role_name"]),
		},
		Permissions: []Permission{},
	}, nil
}

func uintFromAny(value any) (uint64, error) {
	switch id := value.(type) {
	case float64:
		return uint64(id), nil
	case int:
		return uint64(id), nil
	case uint64:
		return id, nil
	case string:
		return strconv.ParseUint(id, 10, 64)
	default:
		return 0, errors.New("unsupported uint value")
	}
}

func stringFromAny(value any) string {
	if text, ok := value.(string); ok {
		return text
	}

	return ""
}
