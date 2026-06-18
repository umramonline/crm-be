package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
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
	Subject   string `json:"sub"`
	TokenType string `json:"typ"`
	ExpiresAt int64  `json:"exp"`
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

func (s *SessionTokenService) Issue(subject string, tokenType string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(tokenType) == "" || len(s.secret) == 0 || ttl <= 0 {
		return "", ErrTokenInvalid
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := SessionTokenClaims{
		Subject:   subject,
		TokenType: tokenType,
		ExpiresAt: s.now().Add(ttl).Unix(),
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

	if claims.Subject == "" || claims.TokenType != expectedType {
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
