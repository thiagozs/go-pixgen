package pix

const (
	//EmvSchema
	TAG_INIT                 = "00"
	TAG_INIT_METHOD          = "01"
	TAG_MAI                  = "26"
	TAG_MCC                  = "52"
	TAG_TRANSACTION_CURRENCY = "53"
	TAG_TRANSACTION_AMOUNT   = "54"
	TAG_COUNTRY_CODE         = "58"
	TAG_MERCHANT_NAME        = "59"
	TAG_MERCHANT_CITY        = "60"
	TAG_ADDITIONAL_DATA      = "62"
	TAG_CRC                  = "63"

	//EmvAdditionalDataSchema
	TAG_TXID = "05"

	//EmvMaiSchema
	TAG_MAI_GUI      = "00"
	TAG_MAI_PIXKEY   = "01"
	TAG_MAI_INFO_ADD = "02"
	TAG_MAI_URL      = "25"
	BC_GUI           = "br.gov.bcb.pix"
)
