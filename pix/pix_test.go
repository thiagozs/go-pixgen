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

	payload, err := pix.GenPayload()
	if err != nil {
		t.Fatalf("generate payload: %v", err)
	}

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

	mai, err := p.generateMAI()
	if err != nil {
		t.Fatalf("generate MAI: %v", err)
	}
	if !strings.Contains(mai, normalizeChars("Produto ABC")) {
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
		name               string
		opts               []Options
		expectNewError     bool
		expectPayloadError bool
	}{
		{
			name:           "invalid pix key",
			opts:           build(OptPixKey("invalid-key")),
			expectNewError: true,
		},
		{
			name:           "empty merchant name",
			opts:           build(OptMerchantName("")),
			expectNewError: true,
		},
		{
			name:           "merchant city too long",
			opts:           build(OptMerchantCity("CIDADEMAIORQUE15")),
			expectNewError: true,
		},
		{
			name:           "dynamic without url",
			opts:           build(OptKind(DYNAMIC), OptTxId(strings.Repeat("A", 25))),
			expectNewError: true,
		},
		{
			name:           "dynamic without https",
			opts:           build(OptKind(DYNAMIC), OptUrl("http://example.com"), OptTxId(strings.Repeat("A", 25))),
			expectNewError: true,
		},
		{
			name:           "invalid amount",
			opts:           build(OptAmount("12.345")),
			expectNewError: true,
		},
		{
			name:           "invalid txid characters",
			opts:           build(OptTxId("INVALID TXID")),
			expectNewError: true,
		},
		{
			name:           "txid too long",
			opts:           build(OptTxId(strings.Repeat("A", 26))),
			expectNewError: true,
		},
		{
			name:           "dynamic without txid",
			opts:           build(OptKind(DYNAMIC), OptUrl("https://example.com/cobranca")),
			expectNewError: true,
		},
		{
			name:           "dynamic txid invalid length",
			opts:           build(OptKind(DYNAMIC), OptUrl("https://example.com/cobranca"), OptTxId(strings.Repeat("A", 26))),
			expectNewError: true,
		},
		{
			name:               "dynamic url exceeds limit",
			opts:               build(OptKind(DYNAMIC), OptUrl("https://example.com/"+strings.Repeat("a", 80)), OptTxId(strings.Repeat("A", 25))),
			expectPayloadError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := New(tc.opts...)
			if tc.expectNewError {
				if err == nil {
					t.Fatalf("expected validation error for %s", tc.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error creating pix: %v", err)
			}

			if tc.expectPayloadError {
				if _, err := p.GenPayload(); err == nil {
					t.Fatalf("expected payload generation error for %s", tc.name)
				}
				return
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
		OptTxId(strings.Repeat("A", 25)),
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
		OptTxId(strings.Repeat("B", 25)),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key := p.params.GetPixKey(); key != "" {
		t.Fatalf("expected empty pix key, got %s", key)
	}
}

func TestGenQRCodeASCII(t *testing.T) {
	opts := []Options{
		OptPixKey("11999887766"),
		OptMerchantName("FULANO DE TAL"),
		OptMerchantCity("SAO PAULO"),
		OptAmount("10.00"),
		OptTxId("TESTE123"),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error creating pix: %v", err)
	}

	ascii, err := p.GenQRCodeASCII()
	if err != nil {
		t.Fatalf("generate ascii qrcode: %v", err)
	}

	if len(strings.TrimSpace(ascii)) == 0 {
		t.Fatalf("ascii qrcode should not be empty")
	}
	if !strings.Contains(ascii, "\n") {
		t.Fatalf("expected multiline ascii qrcode, got %q", ascii)
	}
}
