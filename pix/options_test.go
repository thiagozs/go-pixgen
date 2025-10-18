package pix

import (
	"testing"
)

func TestPixKindString(t *testing.T) {
	tests := []struct {
		input PixKind
		want  string
	}{
		{STATIC, "STATIC"},
		{DYNAMIC, "DYNAMIC"},
	}

	for _, test := range tests {
		got := test.input.String()
		if got != test.want {
			t.Errorf("PixKind.String() = %v; want %v", got, test.want)
		}
	}
}

func TestOptions(t *testing.T) {
	params := &OptionsParams{}
	tests := []struct {
		opt  Options
		get  func() interface{}
		want interface{}
	}{
		{OptQRCodeSize(256), func() interface{} { return params.GetQRCodeSize() }, 256},
		{OptUrl("https://example.com"), func() interface{} { return params.GetUrl() }, "https://example.com"},
		{OptAdditionalInfo("Additional Info"), func() interface{} { return params.GetAdditionalInfo() }, "Additional Info"},
		{OptKind(DYNAMIC), func() interface{} { return params.GetKind() }, DYNAMIC},
		{OptTxId("Transaction123"), func() interface{} { return params.GetTxId() }, "Transaction123"},
		{OptPixKey("123456"), func() interface{} { return params.GetPixKey() }, "123456"},
		{OptDescription("Test Description"), func() interface{} { return params.GetDescription() }, "Test Description"},
		{OptMerchantName("Test Merchant"), func() interface{} { return params.GetMerchantName() }, "Test Merchant"},
		{OptMerchantCity("Test City"), func() interface{} { return params.GetMerchantCity() }, "Test City"},
		{OptAmount("100.00"), func() interface{} { return params.GetAmount() }, "100.00"},
	}

	for _, test := range tests {
		if err := test.opt(params); err != nil {
			t.Fatalf("Option function returned error: %v", err)
		}
		got := test.get()
		if got != test.want {
			t.Errorf("Option function did not set value correctly: got %v; want %v", got, test.want)
		}
	}
}

func TestSetQRCodeContent(t *testing.T) {
	params := &OptionsParams{}
	params.SetQRCodeContent("QR Code Content")

	if params.GetQRCodeContent() != "QR Code Content" {
		t.Errorf("SetQRCodeContent did not set value correctly: got %v; want 'QR Code Content'", params.GetQRCodeContent())
	}
}
