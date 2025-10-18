package pix

// PixKind defines if the QR Code is static or dynamic.
type PixKind int

const (
	STATIC PixKind = iota
	DYNAMIC
)

func (k PixKind) String() string {
	switch k {
	case STATIC:
		return "STATIC"
	case DYNAMIC:
		return "DYNAMIC"
	default:
		return "UNKNOWN"
	}
}

// Options pattern for configuring Pix parameters.
type Options func(o *OptionsParams) error

type Merchant struct {
	name string
	city string
}

type OptionsParams struct {
	txId          string
	pixKey        string
	description   string
	amount        string
	additional    string
	merchant      Merchant
	kind          PixKind
	url           string
	qrcodeContent string
	qrcodeSize    int
}

// Functional options (setters)
func OptQRCodeSize(v int) Options {
	return func(o *OptionsParams) error { o.qrcodeSize = v; return nil }
}
func OptUrl(v string) Options { return func(o *OptionsParams) error { o.url = v; return nil } }
func OptAdditionalInfo(v string) Options {
	return func(o *OptionsParams) error { o.additional = v; return nil }
}
func OptKind(k PixKind) Options  { return func(o *OptionsParams) error { o.kind = k; return nil } }
func OptTxId(v string) Options   { return func(o *OptionsParams) error { o.txId = v; return nil } }
func OptPixKey(v string) Options { return func(o *OptionsParams) error { o.pixKey = v; return nil } }
func OptDescription(v string) Options {
	return func(o *OptionsParams) error { o.description = v; return nil }
}
func OptMerchantName(v string) Options {
	return func(o *OptionsParams) error { o.merchant.name = v; return nil }
}
func OptMerchantCity(v string) Options {
	return func(o *OptionsParams) error { o.merchant.city = v; return nil }
}
func OptAmount(v string) Options                   { return func(o *OptionsParams) error { o.amount = v; return nil } }
func (o *OptionsParams) SetQRCodeContent(v string) { o.qrcodeContent = v }

// Getters
func (o *OptionsParams) GetTxId() string           { return o.txId }
func (o *OptionsParams) GetPixKey() string         { return o.pixKey }
func (o *OptionsParams) GetDescription() string    { return o.description }
func (o *OptionsParams) GetMerchantName() string   { return o.merchant.name }
func (o *OptionsParams) GetMerchantCity() string   { return o.merchant.city }
func (o *OptionsParams) GetAmount() string         { return o.amount }
func (o *OptionsParams) GetKind() PixKind          { return o.kind }
func (o *OptionsParams) GetAdditionalInfo() string { return o.additional }
func (o *OptionsParams) GetUrl() string            { return o.url }
func (o *OptionsParams) GetQRCodeSize() int        { return o.qrcodeSize }
func (o *OptionsParams) GetQRCodeContent() string  { return o.qrcodeContent }
