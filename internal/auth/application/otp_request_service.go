package application

import (
	"context"
	"errors"
	"strings"

	"github.com/umran/new.crm/backend/internal/auth/domain"
)

var (
	ErrInvalidPhone      = domain.ErrInvalidPhone
	ErrInvalidOTPCode    = domain.ErrInvalidOTPCode
	ErrInvalidPassword   = domain.ErrInvalidPassword
	ErrInvalidMFAToken   = errors.New("mfa token is required")
	ErrOTPRequestFailed  = errors.New("otp request failed")
	ErrOTPVerifyRejected = errors.New("otp verification rejected")
	ErrOTPVerifyFailed   = errors.New("otp verification failed")
	ErrPasswordRejected  = errors.New("password login rejected")
	ErrPasswordFailed    = errors.New("password login failed")
)

type OTPRequester interface {
	AdminLogin(ctx context.Context, phone string, password string) (RequestOTPResult, error)
	AdminLoginVerify(ctx context.Context, mfaToken string, otpCode string) (map[string]any, error)
}

type OTPRequestService struct {
	requester OTPRequester
}

func NewOTPRequestService(requester OTPRequester) *OTPRequestService {
	return &OTPRequestService{requester: requester}
}

func (s *OTPRequestService) RequestOTP(ctx context.Context, phone string, password string) (RequestOTPResult, error) {
	if err := domain.ValidatePhone(phone); err != nil {
		return RequestOTPResult{}, err
	}

	if err := domain.ValidatePassword(password); err != nil {
		return RequestOTPResult{}, err
	}

	result, err := s.requester.AdminLogin(ctx, phone, password)
	if err != nil {
		if errors.Is(err, ErrPasswordRejected) {
			return RequestOTPResult{}, ErrPasswordRejected
		}

		return RequestOTPResult{}, ErrOTPRequestFailed
	}

	return result, nil
}

func (s *OTPRequestService) VerifyOTP(ctx context.Context, mfaToken string, otpCode string) (map[string]any, error) {
	if err := validateMFAToken(mfaToken); err != nil {
		return nil, err
	}

	if err := domain.ValidateOTPCode(otpCode); err != nil {
		return nil, err
	}

	data, err := s.requester.AdminLoginVerify(ctx, mfaToken, otpCode)
	if err != nil {
		if errors.Is(err, ErrOTPVerifyRejected) {
			return nil, ErrOTPVerifyRejected
		}

		return nil, ErrOTPVerifyFailed
	}

	if data == nil {
		return nil, ErrOTPVerifyRejected
	}

	return data, nil
}

func validateMFAToken(mfaToken string) error {
	if strings.TrimSpace(mfaToken) == "" {
		return ErrInvalidMFAToken
	}

	return nil
}
