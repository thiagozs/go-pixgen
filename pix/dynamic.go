package pix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DynamicPayload represents the parsed response from a dynamic Pix endpoint.
type DynamicPayload struct {
	Raw       string
	Parsed    *ParsedPayload
	ExpiresAt *time.Time
}

// IsExpired returns whether the payload is expired at the provided instant.
func (d DynamicPayload) IsExpired(at time.Time) bool {
	if d.ExpiresAt == nil {
		return false
	}
	return !d.ExpiresAt.After(at)
}

// FetchDynamicPayload retrieves, parses and validates a Pix dynamic payload from the configured URL.
// When client is nil, http.DefaultClient is used. The response body can be either raw payload text
// or JSON containing a Pix payload field (pixCopyPaste, pix, payload, pixCopiaECola).
func (p *Pix) FetchDynamicPayload(ctx context.Context, client *http.Client) (*DynamicPayload, error) {
	if err := p.Validates(); err != nil {
		return nil, err
	}
	if p.params.kind != DYNAMIC {
		return nil, errors.New("FetchDynamicPayload is supported only for dynamic Pix")
	}

	url := p.params.GetUrl()
	if url == "" {
		return nil, errors.New("dynamic Pix URL must not be empty")
	}

	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch dynamic payload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("dynamic payload request failed with status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	payload, expiresAt, err := interpretDynamicResponse(resp.Header.Get("Content-Type"), body)
	if err != nil {
		return nil, err
	}

	if payload == "" {
		return nil, errors.New("dynamic endpoint did not return a Pix payload")
	}

	parsed, err := ParsePayload(payload)
	if err != nil {
		return nil, fmt.Errorf("remote payload parsing failed: %w", err)
	}

	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return nil, errors.New("dynamic payload is expired")
	}

	return &DynamicPayload{
		Raw:       payload,
		Parsed:    parsed,
		ExpiresAt: expiresAt,
	}, nil
}

func interpretDynamicResponse(contentType string, body []byte) (string, *time.Time, error) {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return "", nil, nil
	}

	if strings.Contains(strings.ToLower(contentType), "json") || looksLikeJSON(raw) {
		payload, expiresAt, err := extractPayloadFromJSON(raw)
		return payload, expiresAt, err
	}

	return raw, nil, nil
}

func looksLikeJSON(raw string) bool {
	return strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "[")
}

func extractPayloadFromJSON(raw string) (string, *time.Time, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		return "", nil, fmt.Errorf("decode dynamic payload json: %w", err)
	}

	payloadKeys := []string{
		"pixCopyPaste",
		"pix_copy_paste",
		"pixCopiaECola",
		"pix_copia_cola",
		"payload",
		"pix",
		"emv",
	}

	var payload string
	for _, key := range payloadKeys {
		if v, ok := lookupStringCaseInsensitive(generic, key); ok {
			payload = v
			break
		}
	}
	if payload == "" {
		return "", nil, errors.New("json response did not contain a Pix payload field")
	}

	expirationKeys := []string{
		"expiresAt",
		"expires_at",
		"expiration",
		"expiraEm",
		"expira_em",
		"expiry",
	}

	for _, key := range expirationKeys {
		if v, ok := lookupStringCaseInsensitive(generic, key); ok && v != "" {
			if ts, err := parseTime(v); err == nil {
				return payload, &ts, nil
			}
		}
	}

	return payload, nil, nil
}

func lookupStringCaseInsensitive(m map[string]interface{}, key string) (string, bool) {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			if s, ok := v.(string); ok {
				return s, true
			}
			return "", false
		}
	}
	return "", false
}

func parseTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", value)
}
