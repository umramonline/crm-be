package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	branchapp "github.com/umran/new.crm/backend/internal/authorization/application"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

type SessionTokenClaims struct {
	UserId       uint64             `json:"user_id"`
	BranchIds    []uint64           `json:"branch_ids,omitempty"`
	Branches     []branchapp.Branch `json:"branches,omitempty"`
	UserFullName string             `json:"user_full_name,omitempty"`
	TokenType    string             `json:"typ"`
	ExpiresAt    int64              `json:"exp"`
	RoleID       uint64             `json:"role_id,omitempty"`
	RoleName     string             `json:"role_name,omitempty"`
}

type SessionTokenService struct {
	secret []byte
	now    func() time.Time
}

func NewSessionTokenService(secret string) *SessionTokenService {
	return &SessionTokenService{
		secret: []byte(secret),
		now:    time.Now,
	}
}

func (s *SessionTokenService) Issue(userId uint64, tokenType string, ttl time.Duration, roleID uint64, roleName string, fullName string, branches []branchapp.Branch) (string, error) {
	if userId == 0 || strings.TrimSpace(tokenType) == "" || strings.TrimSpace(fullName) == "" || len(s.secret) == 0 || ttl <= 0 {
		return "", ErrTokenInvalid
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	branchIds := make([]uint64, 0, len(branches))
	for _, branch := range branches {
		branchIds = append(branchIds, branch.ID)
	}
	claims := SessionTokenClaims{
		UserId:       userId,
		BranchIds:    branchIds,
		Branches:     branches,
		UserFullName: fullName,
		TokenType:    tokenType,
		ExpiresAt:    s.now().Add(ttl).Unix(),
		RoleID:       roleID,
		RoleName:     roleName,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", ErrTokenInvalid
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", ErrTokenInvalid
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	unsignedToken := encodedHeader + "." + encodedClaims

	return unsignedToken + "." + s.sign(unsignedToken), nil
}

func (s *SessionTokenService) Validate(token string, expectedType string) (SessionTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(s.secret) == 0 {
		return SessionTokenClaims{}, ErrTokenInvalid
	}

	unsignedToken := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.sign(unsignedToken))) {
		return SessionTokenClaims{}, ErrTokenInvalid
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return SessionTokenClaims{}, ErrTokenInvalid
	}

	var claims SessionTokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return SessionTokenClaims{}, ErrTokenInvalid
	}

	if claims.UserId == 0 || claims.TokenType != expectedType {
		return SessionTokenClaims{}, ErrTokenInvalid
	}

	if s.now().Unix() >= claims.ExpiresAt {
		return SessionTokenClaims{}, ErrTokenExpired
	}

	return claims, nil
}

func (s *SessionTokenService) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
