package application

import (
	"context"
	"errors"
	"testing"
)

type fakeOTPRequester struct {
	loginCalled  bool
	verifyCalled bool
	phone        string
	password     string
	mfaToken     string
	otpCode      string
	loginResult  RequestOTPResult
	loginData    map[string]any
	err          error
}

func (f *fakeOTPRequester) AdminLogin(_ context.Context, phone string, password string) (RequestOTPResult, error) {
	f.loginCalled = true
	f.phone = phone
	f.password = password

	if f.err != nil {
		return RequestOTPResult{}, f.err
	}

	if f.loginResult.MFARequired || f.loginResult.LoginData != nil {
		return f.loginResult, nil
	}

	return RequestOTPResult{MFARequired: true, MFAToken: "mfa-token"}, nil
}

func (f *fakeOTPRequester) AdminLoginVerify(_ context.Context, mfaToken string, otpCode string) (map[string]any, error) {
	f.verifyCalled = true
	f.mfaToken = mfaToken
	f.otpCode = otpCode

	if f.err != nil {
		return nil, f.err
	}

	return f.loginData, nil
}

func TestOTPRequestServiceRejectsInvalidPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	_, err := service.RequestOTP(context.Background(), "5551234567", "secret")
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

	_, err := service.RequestOTP(context.Background(), "05551234567", " ")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}

	if requester.loginCalled {
		t.Fatal("requester should not be called for empty password")
	}
}

func TestOTPRequestServiceRequestsOTPForValidCredentials(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	result, err := service.RequestOTP(context.Background(), "05551234567", "secret")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !requester.loginCalled {
		t.Fatal("requester should be called")
	}

	if requester.phone != "05551234567" || requester.password != "secret" {
		t.Fatalf("expected payload to be passed through, got phone=%s password=%s", requester.phone, requester.password)
	}

	if !result.MFARequired || result.MFAToken != "mfa-token" {
		t.Fatalf("expected mfa challenge, got %#v", result)
	}
}

func TestOTPRequestServiceWrapsRequesterFailure(t *testing.T) {
	requester := &fakeOTPRequester{err: errors.New("upstream failed")}
	service := NewOTPRequestService(requester)

	_, err := service.RequestOTP(context.Background(), "05551234567", "secret")
	if !errors.Is(err, ErrOTPRequestFailed) {
		t.Fatalf("expected ErrOTPRequestFailed, got %v", err)
	}
}

func TestOTPRequestServiceReturnsRejectedCredentials(t *testing.T) {
	requester := &fakeOTPRequester{err: ErrPasswordRejected}
	service := NewOTPRequestService(requester)

	_, err := service.RequestOTP(context.Background(), "05551234567", "wrong")
	if !errors.Is(err, ErrPasswordRejected) {
		t.Fatalf("expected ErrPasswordRejected, got %v", err)
	}
}

func TestOTPRequestServiceRejectsInvalidMFAToken(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	_, err := service.VerifyOTP(context.Background(), " ", "123456")
	if !errors.Is(err, ErrInvalidMFAToken) {
		t.Fatalf("expected ErrInvalidMFAToken, got %v", err)
	}

	if requester.verifyCalled {
		t.Fatal("requester should not be called for invalid mfa token")
	}
}

func TestOTPRequestServiceRejectsInvalidOTPCode(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	_, err := service.VerifyOTP(context.Background(), "mfa-token", "12345")
	if !errors.Is(err, ErrInvalidOTPCode) {
		t.Fatalf("expected ErrInvalidOTPCode, got %v", err)
	}

	if requester.verifyCalled {
		t.Fatal("requester should not be called for invalid otp code")
	}
}

func TestOTPRequestServiceVerifiesOTPForValidPayload(t *testing.T) {
	requester := &fakeOTPRequester{loginData: map[string]any{"user": map[string]any{"id": float64(1)}}}
	service := NewOTPRequestService(requester)

	data, err := service.VerifyOTP(context.Background(), "mfa-token", "123456")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !requester.verifyCalled {
		t.Fatal("requester should be called")
	}

	if requester.mfaToken != "mfa-token" || requester.otpCode != "123456" {
		t.Fatalf("expected payload to be passed through, got mfa=%s otp=%s", requester.mfaToken, requester.otpCode)
	}

	if data == nil {
		t.Fatal("expected login data")
	}
}

func TestOTPRequestServiceReturnsRejectedWhenOTPDoesNotMatch(t *testing.T) {
	requester := &fakeOTPRequester{err: ErrOTPVerifyRejected}
	service := NewOTPRequestService(requester)

	_, err := service.VerifyOTP(context.Background(), "mfa-token", "123456")
	if !errors.Is(err, ErrOTPVerifyRejected) {
		t.Fatalf("expected ErrOTPVerifyRejected, got %v", err)
	}
}

func TestOTPRequestServiceWrapsOTPVerifyFailure(t *testing.T) {
	requester := &fakeOTPRequester{err: errors.New("upstream failed")}
	service := NewOTPRequestService(requester)

	_, err := service.VerifyOTP(context.Background(), "mfa-token", "123456")
	if !errors.Is(err, ErrOTPVerifyFailed) {
		t.Fatalf("expected ErrOTPVerifyFailed, got %v", err)
	}
}
