package http

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
	maxEmailLength    = 254
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password exceeds maximum length")
	ErrPasswordWeak     = errors.New("password must contain at least one letter and one number")
	ErrEmptyField       = errors.New("field cannot be empty")
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmptyField
	}
	if len(email) > maxEmailLength {
		return errors.New("email exceeds maximum length")
	}
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyField
	}
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}
	
	hasLetter := false
	hasNumber := false
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsNumber(r) {
			hasNumber = true
		}
	}
	
	if !hasLetter || !hasNumber {
		return ErrPasswordWeak
	}
	
	return nil
}

