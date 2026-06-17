package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/auth/application"
)

type fakeOTPRequestService struct {
	requestErr error
	verifyErr  error
	loginErr   error
	phone      string
	otpCode    string
	password   string
	loginData  map[string]any
}

func (f *fakeOTPRequestService) RequestOTP(_ context.Context, phone string) error {
	f.phone = phone

	return f.requestErr
}

func (f *fakeOTPRequestService) VerifyOTP(_ context.Context, phone string, otpCode string) error {
	f.phone = phone
	f.otpCode = otpCode

	return f.verifyErr
}

func (f *fakeOTPRequestService) LoginWithPassword(_ context.Context, phone string, password string) (map[string]any, error) {
	f.phone = phone
	f.password = password

	return f.loginData, f.loginErr
}

func TestOTPHandlerReturnsValidationErrorForInvalidPhone(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: application.ErrInvalidPhone}
	app := newTestApp(service)

	response := performRequest(t, app, `{"phone":"5551234567"}`)
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

	response := performRequest(t, app, `{"phone":"05551234567"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":true`) {
		t.Fatalf("expected success envelope, got %s", body)
	}

	if service.phone != "05551234567" {
		t.Fatalf("expected phone to be passed to service, got %s", service.phone)
	}
}

func TestOTPHandlerDoesNotLeakRequesterErrors(t *testing.T) {
	service := &fakeOTPRequestService{requestErr: errors.New("secret upstream error")}
	app := newTestApp(service)

	response := performRequest(t, app, `{"phone":"05551234567"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if strings.Contains(body, "secret upstream error") {
		t.Fatalf("handler leaked internal error: %s", body)
	}
}

func TestOTPHandlerReturnsValidationErrorForInvalidVerifyPhone(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: application.ErrInvalidPhone}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"phone":"5551234567","otp_code":"123456"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"phone"`) {
		t.Fatalf("expected phone validation envelope, got %s", body)
	}
}

func TestOTPHandlerReturnsValidationErrorForInvalidOTPCode(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: application.ErrInvalidOTPCode}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"phone":"05551234567","otp_code":"12345"}`)
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
	service := &fakeOTPRequestService{}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"phone":"05551234567","otp_code":"123456"}`)
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response.Body)
	if !strings.Contains(body, `"success":true`) || !strings.Contains(body, `"OTP doğrulandı."`) {
		t.Fatalf("expected success envelope, got %s", body)
	}

	if service.phone != "05551234567" || service.otpCode != "123456" {
		t.Fatalf("expected payload to be passed to service, got phone=%s otp=%s", service.phone, service.otpCode)
	}
}

func TestOTPHandlerReturnsRejectedForWrongOTPCode(t *testing.T) {
	service := &fakeOTPRequestService{verifyErr: application.ErrOTPVerifyRejected}
	app := newTestApp(service)

	response := performVerifyRequest(t, app, `{"phone":"05551234567","otp_code":"654321"}`)
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

	response := performVerifyRequest(t, app, `{"phone":"05551234567","otp_code":"123456"}`)
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
	service := &fakeOTPRequestService{loginErr: application.ErrInvalidPhone}
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
	service := &fakeOTPRequestService{loginErr: application.ErrInvalidPassword}
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
	service := &fakeOTPRequestService{loginData: map[string]any{"user": map[string]any{"id": float64(1)}}}
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
}

func TestOTPHandlerReturnsRejectedForWrongPassword(t *testing.T) {
	service := &fakeOTPRequestService{loginErr: application.ErrPasswordRejected}
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
	service := &fakeOTPRequestService{loginErr: errors.New("secret upstream error")}
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

func newTestApp(service *fakeOTPRequestService) *fiber.App {
	app := fiber.New()
	handler := NewOTPHandler(service)
	app.Post("/api/v1/auth/otp/request", handler.RequestOTP)
	app.Post("/api/v1/auth/otp/verify", handler.VerifyOTP)
	app.Post("/api/v1/auth/password/login", handler.LoginWithPassword)

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

func readBody(t *testing.T, reader io.Reader) string {
	t.Helper()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}

	return string(body)
}
