package pix

// EMV schema tags (BACEN Pix)
const (
	TAG_INIT                 = "00" // Payload Format Indicator
	TAG_INIT_METHOD          = "01" // Point of Initiation Method
	TAG_MAI                  = "26" // Merchant Account Information
	TAG_MCC                  = "52" // Merchant Category Code
	TAG_TRANSACTION_CURRENCY = "53" // Transaction Currency
	TAG_TRANSACTION_AMOUNT   = "54" // Transaction Amount
	TAG_COUNTRY_CODE         = "58" // Country Code
	TAG_MERCHANT_NAME        = "59" // Merchant Name
	TAG_MERCHANT_CITY        = "60" // Merchant City
	TAG_ADDITIONAL_DATA      = "62" // Additional Data Field Template
	TAG_CRC                  = "63" // CRC16 Checksum

	// Additional Data Template
	TAG_TXID = "05"

	// Merchant Account Info subtags
	TAG_MAI_GUI      = "00"
	TAG_MAI_PIXKEY   = "01"
	TAG_MAI_INFO_ADD = "02"
	TAG_MAI_URL      = "25"

	// BACEN domain for Pix
	BC_GUI = "br.gov.bcb.pix"
)
