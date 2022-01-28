package pix

type PixKind int

const (
	STATIC PixKind = iota
	DYNAMIC
)

func (pix PixKind) String() string {
	return [...]string{"STATIC", "DYNAMIC"}[pix]
}

type Options func(o *OptionsParams) error

type OptionsParams struct {
	txId          string
	pixKey        string
	description   string
	amount        string
	aditionalInfo string
	merchant      merchant
	kind          PixKind
	url           string
	qrcodeContent string
	qrcodeSize    int
}

type merchant struct {
	name string
	city string
}

func OptQRCodeSize(value int) Options {
	return func(o *OptionsParams) error {
		o.qrcodeSize = value
		return nil
	}
}

func OptUrl(value string) Options {
	return func(o *OptionsParams) error {
		o.url = value
		return nil
	}
}

func OptAditionalInfo(value string) Options {
	return func(o *OptionsParams) error {
		o.aditionalInfo = value
		return nil
	}
}

func OptKind(kind PixKind) Options {
	return func(o *OptionsParams) error {
		o.kind = kind
		return nil
	}
}

func OptTxId(id string) Options {
	return func(o *OptionsParams) error {
		o.txId = id
		return nil
	}
}

func OptPixKey(pixkey string) Options {
	return func(o *OptionsParams) error {
		o.pixKey = pixkey
		return nil
	}
}

func OptDescription(desc string) Options {
	return func(o *OptionsParams) error {
		o.description = desc
		return nil
	}
}

func OptMerchantName(name string) Options {
	return func(o *OptionsParams) error {
		o.merchant.name = name
		return nil
	}
}

func OptMerchantCity(city string) Options {
	return func(o *OptionsParams) error {
		o.merchant.city = city
		return nil
	}
}

func OptAmount(amount string) Options {
	return func(o *OptionsParams) error {
		o.amount = amount
		return nil
	}
}

// ------------- getters

func (o *OptionsParams) GetTxId() string {
	return o.txId
}

func (o *OptionsParams) GetPixKey() string {
	return o.pixKey
}

func (o *OptionsParams) GetDescription() string {
	return o.description
}

func (o *OptionsParams) GetMerchantName() string {
	return o.merchant.name
}

func (o *OptionsParams) GetMerchantCity() string {
	return o.merchant.city
}

func (o *OptionsParams) GetAmount() string {
	return o.amount
}

func (o *OptionsParams) GetKind() PixKind {
	return o.kind
}

func (o *OptionsParams) GetAditionalInfo() string {
	return o.aditionalInfo
}

func (o *OptionsParams) GetUrl() string {
	return o.url
}

func (o *OptionsParams) GetQRCodeSize() int {
	return o.qrcodeSize
}

func (o *OptionsParams) GetQRCodeContent() string {
	return o.qrcodeContent
}

// ------------- setters

func (o *OptionsParams) SetQRCodeContent(value string) {
	o.qrcodeContent = value
}
