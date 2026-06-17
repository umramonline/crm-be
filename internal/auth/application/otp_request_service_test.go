package application

import (
	"context"
	"errors"
	"testing"
)

type fakeOTPRequester struct {
	called bool
	phone  string
	err    error
}

func (f *fakeOTPRequester) RequestOTP(_ context.Context, phone string) error {
	f.called = true
	f.phone = phone

	return f.err
}

func TestOTPRequestServiceRejectsInvalidPhone(t *testing.T) {
	requester := &fakeOTPRequester{}
	service := NewOTPRequestService(requester)

	err := service.RequestOTP(context.Background(), "5551234567")
	if !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}

	if requester.called {
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

	if !requester.called {
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
