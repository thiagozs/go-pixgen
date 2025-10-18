package pix

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Validates ensures the payload meets BACEN Pix requirements
func (p *Pix) Validates() error {
	if p.params.pixKey == "" {
		return errors.New("pixKey must not be empty")
	}
	if p.params.merchant.name == "" {
		return errors.New("merchant name must not be empty")
	}
	if p.params.merchant.city == "" {
		return errors.New("merchant city must not be empty")
	}

	if utf8.RuneCountInString(p.params.merchant.name) > 25 {
		return errors.New("merchant name must be at most 25 characters")
	}
	if utf8.RuneCountInString(p.params.merchant.city) > 15 {
		return errors.New("merchant city must be at most 15 characters")
	}

	if p.params.kind == DYNAMIC && !strings.HasPrefix(p.params.url, "https://") {
		return errors.New("dynamic Pix requires HTTPS URL")
	}

	return nil
}
