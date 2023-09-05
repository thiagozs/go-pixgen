package pix

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {

	opts := []Options{
		OptQRCodeSize(256),
		OptUrl("https://example.com"),
		OptAditionalInfo("Aditional Info"),
		OptKind(STATIC),
		OptMerchantName("Test Merchant"),
		OptMerchantCity("Test City"),
		OptAmount("100.00"),
		OptDescription("Test Description"),
		OptPixKey("123456"),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("failed to create Pix instance: %v", err)
	}

	if pix.params.pixKey != "123456" {
		t.Errorf("expected pixKey to be '123456', got %v", pix.params.pixKey)
	}

	if pix.params.merchant.name != "Test Merchant" {
		t.Errorf("expected merchant name to be 'Test Merchant', got %v", pix.params.merchant.name)
	}
}

func TestGenPayload(t *testing.T) {
	opts := []Options{
		OptQRCodeSize(256),
		OptUrl("https://example.com"),
		OptAditionalInfo("Aditional Info"),
		OptKind(STATIC),
		OptMerchantName("Test Merchant"),
		OptMerchantCity("Test City"),
		OptAmount("100.00"),
		OptDescription("Test Description"),
		OptPixKey("11912345678"),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("failed to create Pix instance: %v", err)
	}

	payload := pix.GenPayload()
	if payload == "" {
		t.Errorf("expected payload to be non-empty")
	}

}

func TestValidates(t *testing.T) {
	tests := []struct {
		opts     []Options
		hasError bool
	}{
		{
			opts: []Options{
				OptMerchantName("Test Merchant"),
				OptMerchantCity("Test City"),
			},
			hasError: true, // pixKey is missing
		},
		{
			opts: []Options{
				OptPixKey("123456"),
				OptMerchantCity("Test City"),
			},
			hasError: true, // merchant name is missing
		},
		{
			opts: []Options{
				OptPixKey("123456"),
				OptMerchantName("Test Merchant"),
			},
			hasError: true, // merchant city is missing
		},
		{
			opts: []Options{
				OptPixKey("123456"),
				OptMerchantName("Test Merchant"),
				OptMerchantCity("Test City"),
			},
			hasError: false, // all required fields are present
		},
	}

	for _, test := range tests {
		pix, err := New(test.opts...)
		if err != nil {
			t.Fatalf("failed to create Pix instance: %v", err)
		}

		err = pix.Validates()
		if (err != nil) != test.hasError {
			t.Errorf("expected error to be %v, got %v", test.hasError, err != nil)
		}
	}
}

func TestGenQRCode(t *testing.T) {
	opts := []Options{
		OptQRCodeSize(256),
		OptUrl("https://example.com"),
		OptAditionalInfo("Aditional Info"),
		OptKind(STATIC),
		OptMerchantName("Test Merchant"),
		OptMerchantCity("Test City"),
		OptAmount("100.00"),
		OptDescription("Test Description"),
		OptPixKey("11912345678"),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("failed to create Pix instance: %v", err)
	}

	if err := pix.Validates(); err != nil {
		t.Fatalf("failed to validate Pix instance: %v", err)
	}

	qrCode, err := pix.GenQRCode()
	if err != nil {
		t.Errorf("failed to generate QR code: %v", err)
	}

	t.Logf("QR code: %v", qrCode)

	if len(qrCode) == 0 {
		t.Errorf("expected QR code to be non-empty")
	}
}

func TestGenerateMAI(t *testing.T) {
	opts := []Options{
		OptPixKey("11999821234"),
		OptMerchantName("Thiago Zilli Sarmento"),
		OptMerchantCity("Ararangua"),
		OptKind(STATIC),
		OptAmount("100.00"),
		OptDescription("Test Description"),
		OptAditionalInfo("https://example.com"),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("failed to create Pix instance: %v", err)
	}

	if err := pix.Validates(); err != nil {
		t.Fatalf("failed to validate Pix instance: %v", err)
	}

	payload := pix.GenPayload()

	t.Logf("payload: %v", payload)

	if !strings.Contains(payload, "11999821234") || !strings.Contains(payload, "https://example.com") {
		t.Errorf("generateMAI function did not work as expected")
	}
}

func TestGenerateAdditionalData(t *testing.T) {
	opts := []Options{
		OptPixKey("123456"),
		OptTxId("Transaction123"),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("failed to create Pix instance: %v", err)
	}

	payload := pix.GenPayload()
	if !strings.Contains(payload, "Transaction123") {
		t.Errorf("generateAdditionalData function did not work as expected")
	}
}

func TestGetCRC16AndFindAndReplaceCRC(t *testing.T) {
	opts := []Options{
		OptPixKey("123456"),
	}

	pix, err := New(opts...)
	if err != nil {
		t.Fatalf("failed to create Pix instance: %v", err)
	}

	payload := pix.GenPayload()
	if !strings.HasSuffix(payload, pix.getCRC16(payload[:len(payload)-4])) {
		t.Errorf("getCRC16 or FindAndReplaceCRC function did not work as expected")
	}
}
