package application

import (
	"context"
	"errors"

	"github.com/umran/new.crm/backend/internal/auth/domain"
)

var (
	ErrInvalidPhone     = domain.ErrInvalidPhone
	ErrOTPRequestFailed = errors.New("otp request failed")
)

type OTPRequester interface {
	RequestOTP(ctx context.Context, phone string) error
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
