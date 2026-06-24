package umramonline

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientVerifyOTPReturnsTrueForSuccessfulResponse(t *testing.T) {
	server := newTestServer(t, http.StatusOK, `{"success":true,"message":"OTP doğrulandı."}`)
	client := newTestClient(server)

	verified, err := client.VerifyOTP(context.Background(), "05551234567", "123456")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !verified {
		t.Fatal("expected otp to be verified")
	}
}

func TestClientVerifyOTPReturnsFalseForRejectedOTP(t *testing.T) {
	server := newTestServer(t, http.StatusUnprocessableEntity, `{"success":false,"message":"Güvenlik kodu hatalı."}`)
	client := newTestClient(server)

	verified, err := client.VerifyOTP(context.Background(), "05551234567", "654321")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if verified {
		t.Fatal("expected otp to be rejected")
	}
}

func TestClientVerifyOTPReturnsErrorForServerFailure(t *testing.T) {
	server := newTestServer(t, http.StatusInternalServerError, `{"success":false,"message":"failed"}`)
	client := newTestClient(server)

	verified, err := client.VerifyOTP(context.Background(), "05551234567", "123456")
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("expected ErrRequestFailed, got %v", err)
	}

	if verified {
		t.Fatal("expected otp to be unverified")
	}
}

func TestClientLoginWithPasswordReturnsDataForSuccessfulResponse(t *testing.T) {
	server := newPasswordLoginTestServer(t, http.StatusOK, `{"success":true,"message":"Giriş başarılı.","data":{"user":{"id":1,"name":"Test User","phone":"05551234567"}}}`)
	client := newTestClient(server)

	data, err := client.LoginWithPassword(context.Background(), "05551234567", "secret")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if data == nil {
		t.Fatal("expected login data")
	}
}

func TestClientLoginWithPasswordReturnsNilForRejectedCredentials(t *testing.T) {
	server := newPasswordLoginTestServer(t, http.StatusUnprocessableEntity, `{"success":false,"message":"Kimlik bilgileri hatalı.","data":null}`)
	client := newTestClient(server)

	data, err := client.LoginWithPassword(context.Background(), "05551234567", "wrong")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if data != nil {
		t.Fatalf("expected nil data for rejected credentials, got %#v", data)
	}
}

func TestClientLoginWithPasswordReturnsErrorForServerFailure(t *testing.T) {
	server := newPasswordLoginTestServer(t, http.StatusInternalServerError, `{"success":false,"message":"failed"}`)
	client := newTestClient(server)

	data, err := client.LoginWithPassword(context.Background(), "05551234567", "secret")
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("expected ErrRequestFailed, got %v", err)
	}

	if data != nil {
		t.Fatalf("expected nil data, got %#v", data)
	}
}

func TestClientListCustomersReturnsItemsForSuccessfulResponse(t *testing.T) {
	server := newCustomersTestServer(t, http.StatusOK, `{"success":true,"items":[{"situation":"Aktif Müşteri","unvan":"Test A.Ş.","credit":10}],"pagination":{"current_page":1,"last_page":1,"per_page":10,"total":1,"from":1,"to":1}}`)
	client := newCustomersTestClient(server)

	result, err := client.ListCustomers(context.Background(), CustomerListQuery{
		Page:    1,
		PerPage: 10,
		Unvan:   "Test",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].Unvan != "Test A.Ş." {
		t.Fatalf("unexpected unvan: %s", result.Items[0].Unvan)
	}
}

func TestClientListCustomersForwardsQueryParameters(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"items":[],"pagination":{"current_page":1,"last_page":1,"per_page":10,"total":0}}`))
	}))
	t.Cleanup(server.Close)

	client := newCustomersTestClient(server)
	_, err := client.ListCustomers(context.Background(), CustomerListQuery{
		Page:       2,
		PerPage:    25,
		Situation:  "Aktif Müşteri",
		SortBy:     "credit",
		SortOrder:  "asc",
		BranchName: "Merkez",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if capturedQuery == "" {
		t.Fatal("expected query parameters to be forwarded")
	}
}

func newCustomersTestClient(server *httptest.Server) *Client {
	return NewClient(Config{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		CustomersPath: "/api/v1/crm/customers",
	})
}

func newCustomersTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/crm/customers" {
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

func newTestClient(server *httptest.Server) *Client {
	return NewClient(Config{
		BaseURL:           server.URL,
		APIKey:            "test-key",
		OTPRequestPath:    "/api/v1/crm/auth/otp/request",
		OTPVerifyPath:     "/api/v1/crm/auth/otp/verify",
		PasswordLoginPath: "/api/v1/crm/auth/password/login",
	})
}

func newTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/crm/auth/otp/verify" {
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

func newPasswordLoginTestServer(t *testing.T, status int, responseBody string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/crm/auth/password/login" {
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
