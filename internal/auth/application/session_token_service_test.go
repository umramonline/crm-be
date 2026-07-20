package application

import (
	"errors"
	"testing"
	"time"

	branchapp "github.com/umran/new.crm/backend/internal/authorization/application"
	sharedauth "github.com/umran/new.crm/backend/internal/shared/auth"
)

func TestSessionTokenServiceIssuesAndValidatesToken(t *testing.T) {
	service := NewSessionTokenService("test-secret")
	service.now = func() time.Time {
		return time.Unix(1000, 0)
	}

	token, err := service.Issue(1, TokenTypeAccess, time.Minute, 30, "Admin", "Test User", []branchapp.Branch{{ID: 5, KisaAd: "A"}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	claims, err := service.Validate(token, TokenTypeAccess)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if claims.UserId != 1 || claims.UserFullName != "Test User" || claims.TokenType != TokenTypeAccess || claims.RoleID != 30 || claims.RoleName != "Admin" {
		t.Fatalf("unexpected claims: %#v", claims)
	}

	if len(claims.BranchIds) != 0 || len(claims.Branches) != 0 {
		t.Fatalf("expected admin claims without branches, got %#v", claims)
	}
}

func TestSessionTokenServiceIncludesBranchesForNonAdmin(t *testing.T) {
	service := NewSessionTokenService("test-secret")
	service.now = func() time.Time {
		return time.Unix(1000, 0)
	}

	branches := []branchapp.Branch{{ID: 5, KisaAd: "A"}}
	token, err := service.Issue(1, TokenTypeAccess, time.Minute, sharedauth.AdminRoleID+1, "User", "Test User", branches)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	claims, err := service.Validate(token, TokenTypeAccess)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(claims.BranchIds) != 1 || claims.BranchIds[0] != 5 {
		t.Fatalf("unexpected branch ids: %#v", claims.BranchIds)
	}

	if len(claims.Branches) != 1 || claims.Branches[0].ID != 5 {
		t.Fatalf("unexpected branches: %#v", claims.Branches)
	}
}

func TestSessionTokenServiceRejectsWrongTokenType(t *testing.T) {
	service := NewSessionTokenService("test-secret")
	service.now = func() time.Time {
		return time.Unix(1000, 0)
	}

	token, err := service.Issue(1, TokenTypeAccess, time.Minute, 0, "", "Test User", nil)
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

	token, err := service.Issue(1, TokenTypeAccess, time.Minute, 0, "", "Test User", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	currentTime = currentTime.Add(2 * time.Minute)

	_, err = service.Validate(token, TokenTypeAccess)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
