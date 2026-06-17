package domain

import "errors"

var ErrInvalidOTPCode = errors.New("otp code must be 6 digits")

func ValidateOTPCode(code string) error {
	if len(code) != 6 {
		return ErrInvalidOTPCode
	}

	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return ErrInvalidOTPCode
		}
	}

	return nil
}
