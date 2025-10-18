package pix

import (
	"strings"
	"testing"
)

func TestBacenConformance(t *testing.T) {
	opts := []Options{
		OptPixKey("123e4567-e12b-12d1-a456-426655440000"),
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("ARACAJU"),
		OptAmount("123.45"),
		OptTxId("W1234567890123456789"),
		OptKind(STATIC),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pix.Validates(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	payload := pix.GenPayload()

	if !strings.HasPrefix(payload, "000201010211") {
		t.Errorf("invalid start: %s", payload[:20])
	}
	if !strings.Contains(payload, "br.gov.bcb.pix") {
		t.Errorf("missing BACEN GUI")
	}
	if !strings.Contains(payload, "6304") {
		t.Errorf("missing CRC16 tag")
	}

	t.Logf("✅ Payload OK: %s", payload)
}

func TestPhoneKeyNormalization(t *testing.T) {
	opts := []Options{
		OptPixKey("11999821234"),
		OptMerchantName("Pix Merchant"),
		OptMerchantCity("CURITIBA"),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "+5511999821234"
	if got := p.params.GetPixKey(); got != want {
		t.Fatalf("expected normalized phone key %q, got %q", want, got)
	}
}

func TestDescriptionFallbackInMAI(t *testing.T) {
	opts := []Options{
		OptPixKey("123e4567-e12b-12d1-a456-426655440000"),
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("ARACAJU"),
		OptDescription("Produto ABC"),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if add := p.params.GetAdditionalInfo(); add != "" {
		t.Fatalf("expected no additional info, got %q", add)
	}

	mai := p.generateMAI()
	if !strings.Contains(mai, "Produto ABC") {
		t.Fatalf("MAI should contain description fallback, got %s", mai)
	}
}

func TestNewValidationErrors(t *testing.T) {
	base := []Options{
		OptPixKey("123e4567-e12b-12d1-a456-426655440000"),
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("CURITIBA"),
	}

	build := func(extra ...Options) []Options {
		opts := make([]Options, len(base))
		copy(opts, base)
		return append(opts, extra...)
	}

	tests := []struct {
		name string
		opts []Options
	}{
		{
			name: "invalid pix key",
			opts: build(OptPixKey("invalid-key")),
		},
		{
			name: "empty merchant name",
			opts: build(OptMerchantName("")),
		},
		{
			name: "merchant city too long",
			opts: build(OptMerchantCity("CIDADEMAIORQUE15")),
		},
		{
			name: "dynamic without url",
			opts: build(OptKind(DYNAMIC)),
		},
		{
			name: "dynamic without https",
			opts: build(OptKind(DYNAMIC), OptUrl("http://example.com")),
		},
		{
			name: "invalid amount",
			opts: build(OptAmount("12.345")),
		},
		{
			name: "invalid txid",
			opts: build(OptTxId("INVALID TXID")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New(tc.opts...); err == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
		})
	}
}

func TestDynamicPixValidation(t *testing.T) {
	opts := []Options{
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("ARACAJU"),
		OptKind(DYNAMIC),
		OptUrl("https://example.com/invoice"),
	}

	if _, err := New(opts...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDynamicPixWithoutPixKey(t *testing.T) {
	opts := []Options{
		OptMerchantName("Fulano de Tal"),
		OptMerchantCity("ARACAJU"),
		OptKind(DYNAMIC),
		OptUrl("https://example.com/invoice"),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key := p.params.GetPixKey(); key != "" {
		t.Fatalf("expected empty pix key, got %s", key)
	}
}
