package umramonline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authapp "github.com/umran/new.crm/backend/internal/auth/application"
)

func TestClientAdminLoginReturnsMFAChallenge(t *testing.T) {
	server := newAdminLoginTestServer(t, http.StatusOK, `{"mfa_required":true,"mfa_token":"abc123","mfa_channel":"sms"}`)
	client := newTestClient(server)

	result, err := client.AdminLogin(context.Background(), "05551234567", "secret")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !result.MFARequired || result.MFAToken != "abc123" || result.MFAChannel != "sms" {
		t.Fatalf("expected mfa challenge, got %#v", result)
	}
}

func TestClientAdminLoginReturnsSessionWhenMFADisabled(t *testing.T) {
	server := newAdminLoginTestServer(t, http.StatusOK, `{"user":{"id":1,"name":"Test User","phone":"05551234567","role_id":30},"token":"1|abc","expires_at":"2026-01-01T00:00:00Z"}`)
	client := newTestClient(server)

	result, err := client.AdminLogin(context.Background(), "05551234567", "secret")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result.MFARequired || result.LoginData == nil {
		t.Fatalf("expected direct login data, got %#v", result)
	}
}

func TestClientAdminLoginReturnsRejectedCredentials(t *testing.T) {
	server := newAdminLoginTestServer(t, http.StatusUnprocessableEntity, `{"message":"Kimlik bilgileri hatalı."}`)
	client := newTestClient(server)

	_, err := client.AdminLogin(context.Background(), "05551234567", "wrong")
	if !errors.Is(err, authapp.ErrPasswordRejected) {
		t.Fatalf("expected ErrPasswordRejected, got %v", err)
	}
}

func TestClientAdminLoginVerifyReturnsLoginData(t *testing.T) {
	server := newAdminLoginVerifyTestServer(t, http.StatusOK, `{"user":{"id":1,"name":"Test User","phone":"05551234567","role_id":30},"token":"1|abc","expires_at":"2026-01-01T00:00:00Z"}`)
	client := newTestClient(server)

	data, err := client.AdminLoginVerify(context.Background(), "abc123", "123456")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if data == nil {
		t.Fatal("expected login data")
	}
}

func TestClientAdminLoginVerifyReturnsRejectedOTP(t *testing.T) {
	server := newAdminLoginVerifyTestServer(t, http.StatusUnprocessableEntity, `{"message":"Güvenlik kodu hatalı."}`)
	client := newTestClient(server)

	_, err := client.AdminLoginVerify(context.Background(), "abc123", "654321")
	if !errors.Is(err, authapp.ErrOTPVerifyRejected) {
		t.Fatalf("expected ErrOTPVerifyRejected, got %v", err)
	}
}

func TestClientAdminLoginVerifyReturnsErrorForServerFailure(t *testing.T) {
	server := newAdminLoginVerifyTestServer(t, http.StatusInternalServerError, `{"message":"failed"}`)
	client := newTestClient(server)

	_, err := client.AdminLoginVerify(context.Background(), "abc123", "123456")
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("expected ErrRequestFailed, got %v", err)
	}
}

func TestClientListCustomersReturnsItemsForSuccessfulResponse(t *testing.T) {
	server := newCustomersTestServer(t, http.StatusOK, `{"success":true,"items":[{"id":100,"plus_card_no":"PC001","il_kodu":"34","ilce_kodu":"001"}],"pagination":{"current_page":1,"last_page":1,"per_page":10,"total":1,"from":1,"to":1}}`)
	client := newCustomersTestClient(server)

	result, err := client.ListCustomers(context.Background(), CustomerListQuery{
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(result.Items) != 1 || result.Items[0].ID != 100 || result.Items[0].PlusCardNo != "PC001" {
		t.Fatalf("unexpected items: %#v", result.Items)
	}
}

func TestClientListCustomersFiltersByIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/crm/customer-list" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"items":[{"id":1,"plus_card_no":"A"},{"id":2,"plus_card_no":"B"}],"pagination":{"current_page":1,"last_page":1,"per_page":500,"total":2}}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(Config{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		APIToken:      "test-token",
		CustomersPath: "/api/v1/admin/crm/customer-list",
	})

	result, err := client.ListCustomers(context.Background(), CustomerListQuery{
		Page:    1,
		PerPage: 10,
		IDs:     []uint64{2},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(result.Items) != 1 || result.Items[0].ID != 2 {
		t.Fatalf("unexpected filtered items: %#v", result.Items)
	}
}

func TestClientListCustomersReturnsErrorForServerFailure(t *testing.T) {
	server := newCustomersTestServer(t, http.StatusInternalServerError, `{"success":false,"message":"failed"}`)
	client := newCustomersTestClient(server)

	_, err := client.ListCustomers(context.Background(), CustomerListQuery{})
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("expected ErrRequestFailed, got %v", err)
	}
}

func TestClientListZonesReturnsItemsForSuccessfulResponse(t *testing.T) {
	server := newZonesTestServer(t, http.StatusOK, `{"success":true,"items":[{"id":1,"name":"Istanbul"}]}`)
	client := newZonesTestClient(server)

	items, err := client.ListZones(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(items))
	}
}

func TestClientListZonesReturnsErrorForServerFailure(t *testing.T) {
	server := newZonesTestServer(t, http.StatusInternalServerError, `{"success":false,"message":"failed"}`)
	client := newZonesTestClient(server)

	_, err := client.ListZones(context.Background(), nil)
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("expected ErrRequestFailed, got %v", err)
	}
}

func newCustomersTestClient(server *httptest.Server) *Client {
	return NewClient(Config{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		APIToken:      "test-token",
		CustomersPath: "/api/v1/admin/crm/customer-list",
	})
}

func newCustomersTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/crm/customer-list" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Fatalf("unexpected api key: %s", r.Header.Get("X-API-KEY"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}))

	t.Cleanup(server.Close)

	return server
}

func newTestClient(server *httptest.Server) *Client {
	return NewClient(Config{
		BaseURL:        server.URL,
		APIKey:         "test-key",
		APIToken:       "test-token",
		OTPRequestPath: "/api/v1/admin/login",
		OTPVerifyPath:  "/api/v1/admin/login/verify",
	})
}

func newAdminLoginTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/admin/login" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Fatalf("unexpected api key: %s", r.Header.Get("X-API-KEY"))
		}

		if r.Header.Get("Authorization") != "" {
			t.Fatal("admin login should not send bearer token")
		}

		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if payload["phone"] == "" || payload["password"] == "" {
			t.Fatalf("expected phone and password in body, got %#v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}))

	t.Cleanup(server.Close)

	return server
}

func newAdminLoginVerifyTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/admin/login/verify" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Fatalf("unexpected api key: %s", r.Header.Get("X-API-KEY"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}))

	t.Cleanup(server.Close)

	return server
}

func newZonesTestClient(server *httptest.Server) *Client {
	return NewClient(Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APIToken:  "test-token",
		ZonesPath: "/api/v1/crm/zones",
	})
}

func newZonesTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/crm/zones" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}))

	t.Cleanup(server.Close)

	return server
}
