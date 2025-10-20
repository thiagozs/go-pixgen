package pix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFetchDynamicPayloadJSON(t *testing.T) {
	opts := []Options{
		OptKind(DYNAMIC),
		OptUrl("https://example.com/pix/123"),
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("CURITIBA"),
		OptTxId(strings.Repeat("A", 25)),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error creating dynamic pix: %v", err)
	}

	payload, err := p.GenPayload()
	if err != nil {
		t.Fatalf("generate payload: %v", err)
	}

	respBody, _ := json.Marshal(map[string]string{
		"pixCopyPaste": payload,
		"expiresAt":    time.Now().Add(10 * time.Minute).Format(time.RFC3339),
	})

	client := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != p.params.url {
			return nil, fmt.Errorf("unexpected url: %s", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(respBody)),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}, nil
	})

	result, err := p.FetchDynamicPayload(context.Background(), client)
	if err != nil {
		t.Fatalf("fetch dynamic payload failed: %v", err)
	}

	if result.Parsed == nil || result.Parsed.Raw != payload {
		t.Fatalf("expected parsed payload")
	}
	expectedURL := stripURLScheme("https://example.com/pix/123")
	if len(result.Parsed.MerchantAccounts) == 0 || result.Parsed.MerchantAccounts[0].URL != expectedURL {
		t.Fatalf("expected merchant account url %q, got %+v", expectedURL, result.Parsed.MerchantAccounts)
	}
	if result.Parsed.MerchantAccounts[0].PixKey != "" {
		t.Fatalf("expected empty pix key for dynamic payload")
	}
	if result.Parsed.AdditionalDataField.TxID != "***" {
		t.Fatalf("expected txid placeholder '***', got %s", result.Parsed.AdditionalDataField.TxID)
	}
	if result.Parsed.Kind() != DYNAMIC {
		t.Fatalf("expected dynamic kind")
	}
	if result.ExpiresAt == nil {
		t.Fatalf("expected expiration time")
	}
	if result.IsExpired(time.Now()) {
		t.Fatalf("payload should not be expired")
	}
}

func TestFetchDynamicPayloadPlainText(t *testing.T) {
	opts := []Options{
		OptKind(DYNAMIC),
		OptUrl("https://example.com/pix/321"),
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("CURITIBA"),
		OptTxId(strings.Repeat("B", 25)),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error creating dynamic pix: %v", err)
	}

	payload, err := p.GenPayload()
	if err != nil {
		t.Fatalf("generate payload: %v", err)
	}

	client := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(payload)),
		}, nil
	})

	result, err := p.FetchDynamicPayload(context.Background(), client)
	if err != nil {
		t.Fatalf("fetch dynamic payload failed: %v", err)
	}

	if result.Parsed == nil || result.Parsed.Raw != payload {
		t.Fatalf("unexpected parsed payload")
	}
	expectedURL := stripURLScheme("https://example.com/pix/321")
	if len(result.Parsed.MerchantAccounts) == 0 || result.Parsed.MerchantAccounts[0].URL != expectedURL {
		t.Fatalf("expected merchant account url %q, got %+v", expectedURL, result.Parsed.MerchantAccounts)
	}
}

func TestFetchDynamicPayloadExpired(t *testing.T) {
	opts := []Options{
		OptKind(DYNAMIC),
		OptUrl("https://example.com/pix/expired"),
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("CURITIBA"),
		OptTxId(strings.Repeat("C", 25)),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error creating dynamic pix: %v", err)
	}

	payload, err := p.GenPayload()
	if err != nil {
		t.Fatalf("generate payload: %v", err)
	}

	respBody, _ := json.Marshal(map[string]string{
		"pix":       payload,
		"expiresAt": time.Now().Add(-time.Minute).Format(time.RFC3339),
	})

	client := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(respBody)),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}, nil
	})

	if _, err := p.FetchDynamicPayload(context.Background(), client); err == nil {
		t.Fatalf("expected expiration error")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newMockHTTPClient(fn func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: roundTripperFunc(fn)}
}
