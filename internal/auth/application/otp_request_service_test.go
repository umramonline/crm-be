package application

import (
	"context"
	"errors"
	"testing"
)

type fakeOTPRequester struct {
	requestCalled bool
	verifyCalled  bool
	loginCalled   bool
	phone         string
	otpCode       string
	password      string
	verified      bool
	loginData     map[string]any
	err           error
}

func (f *fakeOTPRequester) RequestOTP(_ context.Context, phone string) error {
	f.requestCalled = true
	f.phone = phone

	return f.err
}

func (f *fakeOTPRequester) VerifyOTP(_ context.Context, phone string, otpCode string) (bool, error) {
	f.verifyCalled = true
	f.phone = phone
	f.otpCode = otpCode

	return f.verified, f.err
}

func (f *fakeOTPRequester) LoginWithPassword(_ context.Context, phone string, password string) (map[string]any, error) {
	f.loginCalled = true
	f.phone = phone
	f.password = password

	return f.loginData, f.err
}

func TestOTPRequestServiceRejectsInvalidPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	err := service.RequestOTP(context.Background(), "5551234567")
	if !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}

	if requester.requestCalled {
		t.Fatal("requester should not be called for invalid phone")
	}
}

func TestOTPRequestServiceRequestsOTPForValidPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	err := service.RequestOTP(context.Background(), "05551234567")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !requester.requestCalled {
		t.Fatal("requester should be called")
	}

	if requester.phone != "05551234567" {
		t.Fatalf("expected phone to be passed through, got %s", requester.phone)
	}
}

func TestOTPRequestServiceWrapsRequesterFailure(t *testing.T) {
	requester := &fakeOTPRequester{err: errors.New("upstream failed")}
	service := NewOTPRequestService(requester)

	err := service.RequestOTP(context.Background(), "05551234567")
	if !errors.Is(err, ErrOTPRequestFailed) {
		t.Fatalf("expected ErrOTPRequestFailed, got %v", err)
	}
}

func TestOTPRequestServiceRejectsInvalidOTPVerifyPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	err := service.VerifyOTP(context.Background(), "5551234567", "123456")
	if !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}

	if requester.verifyCalled {
		t.Fatal("requester should not be called for invalid phone")
	}
}

func TestOTPRequestServiceRejectsInvalidOTPCode(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	err := service.VerifyOTP(context.Background(), "05551234567", "12345")
	if !errors.Is(err, ErrInvalidOTPCode) {
		t.Fatalf("expected ErrInvalidOTPCode, got %v", err)
	}

	if requester.verifyCalled {
		t.Fatal("requester should not be called for invalid otp code")
	}
}

func TestOTPRequestServiceVerifiesOTPForValidPayload(t *testing.T) {
	requester := &fakeOTPRequester{verified: true}
	service := NewOTPRequestService(requester)

	err := service.VerifyOTP(context.Background(), "05551234567", "123456")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !requester.verifyCalled {
		t.Fatal("requester should be called")
	}

	if requester.phone != "05551234567" || requester.otpCode != "123456" {
		t.Fatalf("expected payload to be passed through, got phone=%s otp=%s", requester.phone, requester.otpCode)
	}
}

func TestOTPRequestServiceReturnsRejectedWhenOTPDoesNotMatch(t *testing.T) {
	requester := &fakeOTPRequester{verified: false}
	service := NewOTPRequestService(requester)

	err := service.VerifyOTP(context.Background(), "05551234567", "123456")
	if !errors.Is(err, ErrOTPVerifyRejected) {
		t.Fatalf("expected ErrOTPVerifyRejected, got %v", err)
	}
}

func TestOTPRequestServiceWrapsOTPVerifyFailure(t *testing.T) {
	requester := &fakeOTPRequester{err: errors.New("upstream failed")}
	service := NewOTPRequestService(requester)

	err := service.VerifyOTP(context.Background(), "05551234567", "123456")
	if !errors.Is(err, ErrOTPVerifyFailed) {
		t.Fatalf("expected ErrOTPVerifyFailed, got %v", err)
	}
}

func TestOTPRequestServiceRejectsInvalidPasswordLoginPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	_, err := service.LoginWithPassword(context.Background(), "5551234567", "secret")
	if !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}

	if requester.loginCalled {
		t.Fatal("requester should not be called for invalid phone")
	}
}

func TestOTPRequestServiceRejectsEmptyPassword(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	_, err := service.LoginWithPassword(context.Background(), "05551234567", " ")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}

	if requester.loginCalled {
		t.Fatal("requester should not be called for empty password")
	}
}

func TestOTPRequestServiceLogsInWithPasswordForValidPayload(t *testing.T) {
	requester := &fakeOTPRequester{loginData: map[string]any{"user": map[string]any{"id": float64(1)}}}
	service := NewOTPRequestService(requester)

	data, err := service.LoginWithPassword(context.Background(), "05551234567", "secret")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !requester.loginCalled {
		t.Fatal("requester should be called")
	}

	if requester.phone != "05551234567" || requester.password != "secret" {
		t.Fatalf("expected payload to be passed through, got phone=%s password=%s", requester.phone, requester.password)
	}

	if data == nil {
		t.Fatal("expected login data")
	}
}

func TestOTPRequestServiceReturnsRejectedWhenPasswordDoesNotMatch(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	_, err := service.LoginWithPassword(context.Background(), "05551234567", "wrong")
	if !errors.Is(err, ErrPasswordRejected) {
		t.Fatalf("expected ErrPasswordRejected, got %v", err)
	}
}

func TestOTPRequestServiceWrapsPasswordLoginFailure(t *testing.T) {
	requester := &fakeOTPRequester{err: errors.New("upstream failed")}
	service := NewOTPRequestService(requester)

	_, err := service.LoginWithPassword(context.Background(), "05551234567", "secret")
	if !errors.Is(err, ErrPasswordFailed) {
		t.Fatalf("expected ErrPasswordFailed, got %v", err)
	}
}
