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
	id          string
	pixKey      string
	description string
	amount      string
	merchant    merchant
	kind        PixKind
}

type merchant struct {
	name string
	city string
}

func OptKind(kind PixKind) Options {
	return func(o *OptionsParams) error {
		o.kind = kind
		return nil
	}
}

func OptId(id string) Options {
	return func(o *OptionsParams) error {
		o.id = id
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

func (o *OptionsParams) GetId() string {
	return o.id
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
