package pix

import (
	"fmt"
	"strings"
	"testing"

	"github.com/snksoft/crc"
)

func TestParsePayloadStatic(t *testing.T) {
	opts := []Options{
		OptPixKey("11999887766"),
		OptMerchantName("FULANO DE TAL"),
		OptMerchantCity("SAO PAULO"),
		OptAmount("123.45"),
		OptTxId("AA12345678901234567890"),
	}

	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error creating pix: %v", err)
	}

	payload := p.GenPayload()

	parsed, err := ParsePayload(payload)
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}

	if parsed.PayloadFormatIndicator != "01" {
		t.Fatalf("unexpected format indicator: %s", parsed.PayloadFormatIndicator)
	}
	if parsed.PointOfInitiationMethod != "11" {
		t.Fatalf("expected static initiation method 11 got %s", parsed.PointOfInitiationMethod)
	}
	if parsed.TransactionAmount != "123.45" {
		t.Fatalf("expected amount 123.45 got %s", parsed.TransactionAmount)
	}
	if parsed.AdditionalDataField.TxID != "AA12345678901234567890" {
		t.Fatalf("expected txid to be AA123..., got %s", parsed.AdditionalDataField.TxID)
	}

	if len(parsed.MerchantAccounts) != 1 {
		t.Fatalf("expected a single merchant account, got %d", len(parsed.MerchantAccounts))
	}

	account := parsed.MerchantAccounts[0]
	if account.GUI != BC_GUI {
		t.Fatalf("expected gui %s got %s", BC_GUI, account.GUI)
	}
	if account.PixKey == "" {
		t.Fatalf("merchant account pix key must not be empty")
	}
}

func TestParsePayloadInvalidCRC(t *testing.T) {
	opts := []Options{
		OptPixKey("11999887766"),
		OptMerchantName("FULANO DE TAL"),
		OptMerchantCity("SAO PAULO"),
	}
	p, err := New(opts...)
	if err != nil {
		t.Fatalf("unexpected error creating pix: %v", err)
	}

	payload := p.GenPayload()
	tampered := payload[:len(payload)-1] + "0"

	if _, err := ParsePayload(tampered); err == nil || !strings.Contains(err.Error(), "crc mismatch") {
		t.Fatalf("expected crc mismatch error, got %v", err)
	}
}

func TestParsePayloadMissingFields(t *testing.T) {
	payload := buildPayloadWithoutMerchantName()

	if _, err := ParsePayload(payload); err == nil {
		t.Fatalf("expected error due to missing merchant data")
	}
}

func buildPayloadWithoutMerchantName() string {
	tlv := func(tag, value string) string {
		return fmt.Sprintf("%s%02d%s", tag, len(value), value)
	}

	mai := strings.Join([]string{
		tlv(TAG_MAI_GUI, BC_GUI),
		tlv(TAG_MAI_PIXKEY, "12345678901"),
	}, "")

	additional := tlv(TAG_TXID, "TX123")

	payload := strings.Join([]string{
		tlv(TAG_INIT, "01"),
		tlv(TAG_INIT_METHOD, "11"),
		tlv(TAG_MAI, mai),
		tlv(TAG_MCC, "0000"),
		tlv(TAG_TRANSACTION_CURRENCY, "986"),
		tlv(TAG_COUNTRY_CODE, "BR"),
		tlv(TAG_MERCHANT_CITY, "SAO PAULO"),
		tlv(TAG_ADDITIONAL_DATA, additional),
		tlv(TAG_CRC, "0000"),
	}, "")

	base := payload[:len(payload)-4]
	checksum := fmt.Sprintf("%04X", crc.CalculateCRC(crc.CCITT, []byte(base)))
	return base + checksum
}
