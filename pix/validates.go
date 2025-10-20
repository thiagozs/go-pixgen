package pix

import (
	"errors"
	"fmt"
	"net/mail"
	urlpkg "net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	uuidPattern   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	txidPattern   = regexp.MustCompile(`^[A-Za-z0-9]{1,25}$`)
	amountPattern = regexp.MustCompile(`^\d{1,10}\.\d{2}$`)
)

// Validates ensures the payload meets BACEN Pix requirements
func (p *Pix) Validates() error {
	if p.params == nil {
		return errors.New("pix params must not be nil")
	}

	key := strings.TrimSpace(p.params.pixKey)
	if key == "" {
		if p.params.kind != DYNAMIC {
			return errors.New("pixKey must not be empty")
		}
	} else {
		normalizedKey, err := normalizePixKey(key)
		if err != nil {
			return err
		}
		if utf8.RuneCountInString(normalizedKey) > 77 {
			return errors.New("pixKey must be at most 77 characters")
		}
		p.params.pixKey = normalizedKey
	}

	name := strings.TrimSpace(p.params.merchant.name)
	if name == "" {
		return errors.New("merchant name must not be empty")
	}
	if utf8.RuneCountInString(name) > 25 {
		return errors.New("merchant name must be at most 25 characters")
	}
	p.params.merchant.name = name

	city := strings.TrimSpace(p.params.merchant.city)
	if city == "" {
		return errors.New("merchant city must not be empty")
	}
	if utf8.RuneCountInString(city) > 15 {
		return errors.New("merchant city must be at most 15 characters")
	}
	p.params.merchant.city = city

	if desc := strings.TrimSpace(p.params.description); desc != "" {
		if utf8.RuneCountInString(desc) > 72 {
			return errors.New("description must be at most 72 characters")
		}
		p.params.description = desc
	}

	if add := strings.TrimSpace(p.params.additional); add != "" {
		if utf8.RuneCountInString(add) > 72 {
			return errors.New("additional info must be at most 72 characters")
		}
		p.params.additional = add
	}

	if amount := strings.TrimSpace(p.params.amount); amount != "" {
		if !amountPattern.MatchString(amount) {
			return fmt.Errorf("invalid amount format: %s", amount)
		}
		p.params.amount = amount
	}

	if p.params.kind != STATIC && p.params.kind != DYNAMIC {
		return errors.New("pix kind must be static or dynamic")
	}

	if p.params.kind == DYNAMIC {
		rawURL := strings.TrimSpace(p.params.url)
		if rawURL == "" {
			return errors.New("dynamic Pix requires URL")
		}
		parsedURL, err := urlpkg.Parse(rawURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return errors.New("dynamic Pix requires valid URL")
		}
		if parsedURL.Scheme != "https" {
			host := parsedURL.Hostname()
			if host != "localhost" && !strings.HasPrefix(host, "127.") {
				return errors.New("dynamic Pix requires HTTPS URL")
			}
		}
		p.params.url = rawURL
	}

	txid := strings.TrimSpace(p.params.txId)
	switch p.params.kind {
	case STATIC:
		if txid != "" {
			if !txidPattern.MatchString(txid) {
				return fmt.Errorf("txid must be alphanumeric up to 25 characters")
			}
			p.params.txId = strings.ToUpper(txid)
		}
	case DYNAMIC:
		if txid == "" {
			return errors.New("dynamic Pix requires txid")
		}
		if !txidPattern.MatchString(txid) {
			return fmt.Errorf("dynamic txid must be alphanumeric up to 25 characters")
		}
		p.params.txId = strings.ToUpper(txid)
	}

	return nil
}

func normalizePixKey(key string) (string, error) {
	if uuidPattern.MatchString(key) {
		return strings.ToLower(key), nil
	}

	if normalized, ok := normalizePhoneKey(key); ok {
		return normalized, nil
	}

	if normalized, ok := normalizeCPFKey(key); ok {
		return normalized, nil
	}

	if normalized, ok := normalizeCNPJKey(key); ok {
		return normalized, nil
	}

	if normalized, ok := normalizeEmailKey(key); ok {
		return normalized, nil
	}

	return "", errors.New("invalid pix key format")
}

func normalizeEmailKey(key string) (string, bool) {
	addr, err := mail.ParseAddress(key)
	if err != nil {
		return "", false
	}
	if addr.Address != key {
		return "", false
	}
	if utf8.RuneCountInString(key) > 77 {
		return "", false
	}
	return key, true
}

func normalizePhoneKey(key string) (string, bool) {
	digits := digitsOnly(key)
	switch {
	case len(digits) == 13 && strings.HasPrefix(digits, "55"):
		return "+55" + digits[2:], true
	case len(digits) == 12 && strings.HasPrefix(digits, "55"):
		return "+55" + digits[2:], true
	case len(digits) == 11:
		return "+55" + digits, true
	case len(digits) == 10:
		return "+55" + digits, true
	default:
		return "", false
	}
}

func normalizeCPFKey(key string) (string, bool) {
	digits := digitsOnly(key)
	if len(digits) != 11 {
		return "", false
	}
	if isAllSameRune(digits) {
		return "", false
	}

	firstDigit := calculateCPFCheckDigit(digits[:9])
	secondDigit := calculateCPFCheckDigit(digits[:10])

	if firstDigit != digits[9] || secondDigit != digits[10] {
		return "", false
	}
	return digits, true
}

func normalizeCNPJKey(key string) (string, bool) {
	digits := digitsOnly(key)
	if len(digits) != 14 {
		return "", false
	}
	if isAllSameRune(digits) {
		return "", false
	}

	firstDigit := calculateCNPJCheckDigit(digits[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	secondDigit := calculateCNPJCheckDigit(digits[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})

	if firstDigit != digits[12] || secondDigit != digits[13] {
		return "", false
	}
	return digits, true
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isAllSameRune(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s[1:] {
		if r != rune(s[0]) {
			return false
		}
	}
	return true
}

func calculateCPFCheckDigit(digits string) byte {
	sum := 0
	for i, r := range digits {
		sum += int(r-'0') * (len(digits) + 1 - i)
	}
	mod := sum % 11
	if mod < 2 {
		return '0'
	}
	return byte(11-mod) + '0'
}

func calculateCNPJCheckDigit(digits string, weights []int) byte {
	sum := 0
	for i, r := range digits {
		sum += int(r-'0') * weights[i]
	}
	mod := sum % 11
	if mod < 2 {
		return '0'
	}
	return byte(11-mod) + '0'
}
