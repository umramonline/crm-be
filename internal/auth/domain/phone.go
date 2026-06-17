package domain

import "errors"

var ErrInvalidPhone = errors.New("phone must be in 05XXXXXXXXX format")

func ValidatePhone(phone string) error {
	if len(phone) != 11 {
		return ErrInvalidPhone
	}

	if phone[0] != '0' || phone[1] != '5' {
		return ErrInvalidPhone
	}

	for _, digit := range phone[2:] {
		if digit < '0' || digit > '9' {
			return ErrInvalidPhone
		}
	}

	return nil
}
