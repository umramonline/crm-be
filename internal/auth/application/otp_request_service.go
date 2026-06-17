package application

import (
	"context"
	"errors"

	"github.com/umran/new.crm/backend/internal/auth/domain"
)

var (
	ErrInvalidPhone      = domain.ErrInvalidPhone
	ErrInvalidOTPCode    = domain.ErrInvalidOTPCode
	ErrInvalidPassword   = domain.ErrInvalidPassword
	ErrOTPRequestFailed  = errors.New("otp request failed")
	ErrOTPVerifyRejected = errors.New("otp verification rejected")
	ErrOTPVerifyFailed   = errors.New("otp verification failed")
	ErrPasswordRejected  = errors.New("password login rejected")
	ErrPasswordFailed    = errors.New("password login failed")
)

type OTPRequester interface {
	RequestOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone string, otpCode string) (bool, error)
	LoginWithPassword(ctx context.Context, phone string, password string) (map[string]any, error)
}

type OTPRequestService struct {
	requester OTPRequester
}

func NewOTPRequestService(requester OTPRequester) *OTPRequestService {
	return &OTPRequestService{requester: requester}
}

func (s *OTPRequestService) RequestOTP(ctx context.Context, phone string) error {
	if err := domain.ValidatePhone(phone); err != nil {
		return err
	}

	if err := s.requester.RequestOTP(ctx, phone); err != nil {
		return ErrOTPRequestFailed
	}

	return nil
}

func (s *OTPRequestService) VerifyOTP(ctx context.Context, phone string, otpCode string) error {
	if err := domain.ValidatePhone(phone); err != nil {
		return err
	}

	if err := domain.ValidateOTPCode(otpCode); err != nil {
		return err
	}

	verified, err := s.requester.VerifyOTP(ctx, phone, otpCode)
	if err != nil {
		return ErrOTPVerifyFailed
	}

	if !verified {
		return ErrOTPVerifyRejected
	}

	return nil
}

func (s *OTPRequestService) LoginWithPassword(ctx context.Context, phone string, password string) (map[string]any, error) {
	if err := domain.ValidatePhone(phone); err != nil {
		return nil, err
	}

	if err := domain.ValidatePassword(password); err != nil {
		return nil, err
	}

	data, err := s.requester.LoginWithPassword(ctx, phone, password)
	if err != nil {
		return nil, ErrPasswordFailed
	}

	if data == nil {
		return nil, ErrPasswordRejected
	}

	return data, nil
}
