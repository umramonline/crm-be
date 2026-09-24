package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/auth/application"
	branchapp "github.com/umran/new.crm/backend/internal/authorization/application"
)

type fakeOTPRequestService struct {
	requestErr    error
	verifyErr     error
	phone         string
	otpCode       string
	password      string
	mfaToken      string
	loginData     map[string]any
	requestResult application.RequestOTPResult
}

type fakeSessionTokenService struct {
	accessToken  string
	refreshToken string
	userID       uint64
	fullName     string
	roleID       uint64
	roleName     string
	validateErr  error
	issueErr     error
}

func (f *fakeOTPRequestService) RequestOTP(_ context.Context, phone string, password string) (application.RequestOTPResult, error) {
	f.phone = phone
	f.password = password

	if f.requestErr != nil {
		return application.RequestOTPResult{}, f.requestErr
	}

	if f.requestResult.MFARequired || f.requestResult.LoginData != nil {
		return f.requestResult, nil
	}

	return application.RequestOTPResult{
		MFARequired: true,
		MFAToken:    "mfa-token",
		MFAChannel:  "sms",
	}, nil
}

func (f *fakeOTPRequestService) VerifyOTP(_ context.Context, mfaToken string, otpCode string) (map[string]any, error) {
	f.mfaToken = mfaToken
	f.otpCode = otpCode

	if f.verifyErr != nil {
		return nil, f.verifyErr
	}

	return f.loginData, nil
}

func (f *fakeSessionTokenService) Issue(userID uint64, tokenType string, _ time.Duration, roleID uint64, roleName string, fullName string, _ []branchapp.Branch, _ string) (string, error) {
	if f.issueErr != nil {
		return "", f.issueErr
	}

	f.userID = userID
	f.fullName = fullName
	f.roleID = roleID
	f.roleName = roleName

	if tokenType == application.TokenTypeRefresh {
		return f.refreshToken, nil
	}

	return f.accessToken, nil
}

func (f *fakeSessionTokenService) Validate(_ string, expectedType string) (application.SessionTokenClaims, error) {
	if f.validateErr != nil {
		return application.SessionTokenClaims{}, f.validateErr
	}

	return application.SessionTokenClaims{
		UserId:       f.userID,
		UserFullName: f.fullName,
		TokenType:    expectedType,
		ExpiresAt:    time.Now().Add(time.Minute).Unix(),
		RoleID:       f.roleID,
		RoleName:     f.roleName,
	}, nil
}

func TestOTPHandlerReturnsValidationErrorForInvalidPhone(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: application.ErrInvalidPhone}
	app := newTestApp(service)

	response := performRequest(t, app, `{"phone":"5551234567","password":"secret"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"phone"`) {
		t.Fatalf("expected validation envelope, got %s", body)
	}
}

func TestOTPHandlerReturnsSuccessEnvelope(t *testing.T) {
	service := &fakeOTPRequestService{}
	app := newTestApp(service)

	response := performRequest(t, app, `{"phone":"05551234567","password":"secret"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":true`) || !strings.Contains(body, `"mfa_token"`) {
		t.Fatalf("expected success envelope with mfa token, got %s", body)
	}

	if service.phone != "05551234567" || service.password != "secret" {
		t.Fatalf("expected payload to be passed to service, got phone=%s password=%s", service.phone, service.password)
	}
}

func TestOTPHandlerDoesNotLeakRequesterErrors(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: errors.New("secret upstream error")}
	app := newTestApp(service)

	response := performRequest(t, app, `{"phone":"05551234567","password":"secret"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if strings.Contains(body, "secret upstream error") {
		t.Fatalf("handler leaked internal error: %s", body)
	}
}

func TestOTPHandlerReturnsValidationErrorForInvalidMFAToken(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: application.ErrInvalidMFAToken}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"mfa_token":" ","otp_code":"123456"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"mfa_token"`) {
		t.Fatalf("expected mfa token validation envelope, got %s", body)
	}
}

func TestOTPHandlerReturnsValidationErrorForInvalidOTPCode(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: application.ErrInvalidOTPCode}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"mfa_token":"mfa-token","otp_code":"12345"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"otp_code"`) {
		t.Fatalf("expected otp code validation envelope, got %s", body)
	}
}

func TestOTPHandlerReturnsVerifySuccessEnvelope(t *testing.T) {
	service := &fakeOTPRequestService{loginData: map[string]any{"user": map[string]any{"id": float64(1), "role_id": float64(30)}}}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"mfa_token":"mfa-token","otp_code":"123456"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":true`) || !strings.Contains(body, `"Giriş başarılı."`) {
		t.Fatalf("expected success envelope, got %s", body)
	}

	if service.mfaToken != "mfa-token" || service.otpCode != "123456" {
		t.Fatalf("expected payload to be passed to service, got mfa=%s otp=%s", service.mfaToken, service.otpCode)
	}

	if !hasCookie(response.Cookies(), "access_token") || !hasCookie(response.Cookies(), "refresh_token") {
		t.Fatalf("expected auth cookies, got %#v", response.Cookies())
	}
}

func TestOTPHandlerReturnsRejectedForWrongOTPCode(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: application.ErrOTPVerifyRejected}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"mfa_token":"mfa-token","otp_code":"654321"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"Güvenlik kodu hatalı."`) {
		t.Fatalf("expected rejected envelope, got %s", body)
	}
}

func TestOTPHandlerDoesNotLeakVerifyRequesterErrors(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: errors.New("secret upstream error")}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"mfa_token":"mfa-token","otp_code":"123456"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if strings.Contains(body, "secret upstream error") {
		t.Fatalf("handler leaked internal error: %s", body)
	}
}

func TestOTPHandlerReturnsValidationErrorForInvalidPasswordLoginPhone(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: application.ErrInvalidPhone}
	app := newTestApp(service)

	response := performPasswordLoginRequest(t, app, `{"phone":"5551234567","password":"secret"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"phone"`) {
		t.Fatalf("expected phone validation envelope, got %s", body)
	}
}

func TestOTPHandlerReturnsValidationErrorForEmptyPassword(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: application.ErrInvalidPassword}
	app := newTestApp(service)

	response := performPasswordLoginRequest(t, app, `{"phone":"05551234567","password":""}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"password"`) {
		t.Fatalf("expected password validation envelope, got %s", body)
	}
}

func TestOTPHandlerReturnsPasswordLoginSuccessEnvelope(t *testing.T) {
	service := &fakeOTPRequestService{
		requestResult: application.RequestOTPResult{
			LoginData: map[string]any{"user": map[string]any{"id": float64(1), "role_id": float64(30)}},
		},
	}
	app := newTestApp(service)

	response := performPasswordLoginRequest(t, app, `{"phone":"05551234567","password":"secret"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":true`) || !strings.Contains(body, `"Giriş başarılı."`) {
		t.Fatalf("expected success envelope, got %s", body)
	}

	if service.phone != "05551234567" || service.password != "secret" {
		t.Fatalf("expected payload to be passed to service, got phone=%s password=%s", service.phone, service.password)
	}

	if !hasCookie(response.Cookies(), "access_token") || !hasCookie(response.Cookies(), "refresh_token") {
		t.Fatalf("expected auth cookies, got %#v", response.Cookies())
	}
}

func TestOTPHandlerReturnsRejectedForWrongPassword(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: application.ErrPasswordRejected}
	app := newTestApp(service)

	response := performPasswordLoginRequest(t, app, `{"phone":"05551234567","password":"wrong"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"Kimlik bilgileri hatalı."`) {
		t.Fatalf("expected rejected envelope, got %s", body)
	}
}

func TestOTPHandlerDoesNotLeakPasswordLoginRequesterErrors(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: errors.New("secret upstream error")}
	app := newTestApp(service)

	response := performPasswordLoginRequest(t, app, `{"phone":"05551234567","password":"secret"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if strings.Contains(body, "secret upstream error") {
		t.Fatalf("handler leaked internal error: %s", body)
	}
}

func TestOTPHandlerRefreshesAccessCookie(t *testing.T) {
	service := &fakeOTPRequestService{}
	tokenService := &fakeSessionTokenService{
		accessToken:  "new-access-token",
		refreshToken: "refresh-token",
		userID:       1,
		fullName:     "Test User",
	}
	app := newTestAppWithTokenService(service, tokenService)

	response := performRefreshRequest(t, app, "refresh-token")
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	if !hasCookie(response.Cookies(), "access_token") {
		t.Fatalf("expected refreshed access cookie, got %#v", response.Cookies())
	}
}

func TestOTPHandlerRejectsRefreshWithoutCookie(t *testing.T) {
	service := &fakeOTPRequestService{}
	app := newTestApp(service)

	response := performRefreshRequest(t, app, "")
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.StatusCode)
	}
}

func TestOTPHandlerClearsCookiesOnLogout(t *testing.T) {
	service := &fakeOTPRequestService{}
	app := newTestApp(service)

	response := performLogoutRequest(t, app)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	if !hasExpiredCookie(response.Cookies(), "access_token") || !hasExpiredCookie(response.Cookies(), "refresh_token") {
		t.Fatalf("expected expired auth cookies, got %#v", response.Cookies())
	}
}

func TestOTPHandlerReturnsSessionForValidAccessCookie(t *testing.T) {
	service := &fakeOTPRequestService{}
	tokenService := &fakeSessionTokenService{userID: 1, fullName: "Test User"}
	app := newTestAppWithTokenService(service, tokenService)

	response := performSessionRequest(t, app, "access-token")
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"user_id":1`) {
		t.Fatalf("expected session user id, got %s", body)
	}
}

func newTestApp(service *fakeOTPRequestService) *fiber.App {
	return newTestAppWithTokenService(service, &fakeSessionTokenService{
		accessToken:  "access-token",
		refreshToken: "refresh-token",
		userID:       1,
		fullName:     "Test User",
	})
}

func newTestAppWithTokenService(service *fakeOTPRequestService, tokenService *fakeSessionTokenService) *fiber.App {
	app := fiber.New()
	handler := NewOTPHandler(service, tokenService, SessionConfig{
		AccessTTL:      time.Minute,
		RefreshTTL:     time.Hour,
		CookieSameSite: "Lax",
	})
	app.Post("/api/v1/auth/otp/request", handler.RequestOTP)
	app.Post("/api/v1/auth/otp/verify", handler.VerifyOTP)
	app.Post("/api/v1/auth/password/login", handler.LoginWithPassword)
	app.Post("/api/v1/auth/refresh", handler.RefreshSession)
	app.Post("/api/v1/auth/logout", handler.Logout)
	app.Get("/api/v1/auth/session", handler.Session)

	return app
}

func performRequest(t *testing.T, app *fiber.App, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest("POST", "/api/v1/auth/otp/request", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func performVerifyRequest(t *testing.T, app *fiber.App, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest("POST", "/api/v1/auth/otp/verify", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func performPasswordLoginRequest(t *testing.T, app *fiber.App, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest("POST", "/api/v1/auth/password/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func performRefreshRequest(t *testing.T, app *fiber.App, refreshToken string) *http.Response {
	t.Helper()

	request := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	if refreshToken != "" {
		request.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	}

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func performLogoutRequest(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()

	request := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func performSessionRequest(t *testing.T, app *fiber.App, accessToken string) *http.Response {
	t.Helper()

	request := httptest.NewRequest("GET", "/api/v1/auth/session", nil)
	if accessToken != "" {
		request.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	}

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return response
}

func readBody(t *testing.T, reader io.Reader) string {
	t.Helper()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}

	return string(body)
}

func hasCookie(cookies []*http.Cookie, name string) bool {
	for _, cookie := range cookies {
		if cookie.Name == name && cookie.Value != "" {
			return true
		}
	}

	return false
}

func hasExpiredCookie(cookies []*http.Cookie, name string) bool {
	for _, cookie := range cookies {
		if cookie.Name == name && cookie.Value == "" {
			return true
		}
	}

	return false
}
