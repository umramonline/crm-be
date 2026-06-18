package application

import (
	"errors"
	"testing"
	"time"
)

func TestSessionTokenServiceIssuesAndValidatesToken(t *testing.T) {
	service := NewSessionTokenService("test-secret")
	service.now = func() time.Time {
		return time.Unix(1000, 0)
	}

	token, err := service.Issue("1", TokenTypeAccess, time.Minute)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	claims, err := service.Validate(token, TokenTypeAccess)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if claims.Subject != "1" || claims.TokenType != TokenTypeAccess {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestSessionTokenServiceRejectsWrongTokenType(t *testing.T) {
	service := NewSessionTokenService("test-secret")
	service.now = func() time.Time {
		return time.Unix(1000, 0)
	}

	token, err := service.Issue("1", TokenTypeAccess, time.Minute)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = service.Validate(token, TokenTypeRefresh)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestSessionTokenServiceRejectsExpiredToken(t *testing.T) {
	currentTime := time.Unix(1000, 0)
	service := NewSessionTokenService("test-secret")
	service.now = func() time.Time {
		return currentTime
	}

	token, err := service.Issue("1", TokenTypeAccess, time.Minute)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	currentTime = currentTime.Add(2 * time.Minute)

	_, err = service.Validate(token, TokenTypeAccess)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
