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

	pix, _ := New(opts...)
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
