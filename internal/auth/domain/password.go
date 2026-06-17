package domain

import (
	"errors"
	"strings"
)

var ErrInvalidPassword = errors.New("password must not be empty")

func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrInvalidPassword
	}

	return nil
}
