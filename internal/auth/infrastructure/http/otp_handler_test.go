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
	err   error
	phone string
}

func (f *fakeOTPRequestService) RequestOTP(_ context.Context, phone string) error {
	f.phone = phone

	return f.err
}

func TestOTPHandlerReturnsValidationErrorForInvalidPhone(t *testing.T) {
	service := &fakeOTPRequestService{err: application.ErrInvalidPhone}
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
	service := &fakeOTPRequestService{err: errors.New("secret upstream error")}
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

func newTestApp(service *fakeOTPRequestService) *fiber.App {
	app := fiber.New()
	handler := NewOTPHandler(service)
	app.Post("/api/v1/auth/otp/request", handler.RequestOTP)

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

func readBody(t *testing.T, reader io.Reader) string {
	t.Helper()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}

	return string(body)
}
