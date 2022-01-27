package pix

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/snksoft/crc"
)

type Pix struct {
	params *OptionsParams
}

func New(opts ...Options) (*Pix, error) {
	mts := &OptionsParams{}
	for _, op := range opts {
		err := op(mts)
		if err != nil {
			return nil, err
		}
	}
	return newInstance(mts)
}

func newInstance(params *OptionsParams) (*Pix, error) {
	return &Pix{params}, nil
}

func (p *Pix) GenPayload() string {

	tags := []string{}

	switch p.params.GetKind() {
	case STATIC:
		tags = []string{
			p.getValue(TAG_INIT, "01"),
			p.getValue(TAG_MAI, p.generateMAI()),
			p.getValue(TAG_MCC, "0000"),
			p.getValue(TAG_TRANSACTION_CURRENCY, "986"),
			p.getValue(TAG_COUNTRY_CODE, "BR"),
			p.getValue(TAG_TRANSACTION_AMOUNT, p.params.amount),
			p.getValue(TAG_MERCHANT_NAME, p.params.merchant.name),
			p.getValue(TAG_MERCHANT_CITY, p.params.merchant.city),
			p.getValue(TAG_ADDITIONAL_DATA, p.generateAdditionalData()),
			p.getValue(TAG_CRC, "0000"),
		}
	case DYNAMIC:
		tags = []string{
			p.getValue(TAG_INIT, "01"),
			p.getValue(TAG_MAI, p.generateMAI()),
			p.getValue(TAG_MCC, "0000"),
			p.getValue(TAG_TRANSACTION_CURRENCY, "986"),
			p.getValue(TAG_COUNTRY_CODE, "BR"),
			p.getValue(TAG_TRANSACTION_AMOUNT, p.params.amount),
			p.getValue(TAG_MERCHANT_NAME, p.params.merchant.name),
			p.getValue(TAG_MERCHANT_CITY, p.params.merchant.city),
			p.getValue(TAG_ADDITIONAL_DATA, p.generateAdditionalData()),
			p.getValue(TAG_CRC, "0000"),
		}
	}

	payload := strings.Join(tags, "")

	payload = p.FindAndReplaceCRC(payload)

	return payload
}

func (p *Pix) getValue(id, content string) string {
	return fmt.Sprintf("%s%02d%s", id, len(content), content)
}

func (p *Pix) generateMAI() string {
	switch p.params.GetKind() {
	case STATIC:
		tags := []string{
			p.getValue(TAG_MAI_GUI, BC_GUI),
			p.getValue(TAG_MAI_PIXKEY, p.params.pixKey),
			p.getValue(TAG_MAI_INFO_ADD, "Gerado por Pix-Utils"),
			//p.getValue(TAG_MAI_INFO_ADD, p.params.infoAdicional),
		}
		return strings.Join(tags, "")
	case DYNAMIC:
		tags := []string{
			p.getValue(TAG_MAI_GUI, BC_GUI),
			p.getValue(TAG_MAI_URL, "https://www.pix.com.br/"),
			//p.getValue(TAG_MAI_URL, p.params.url),
		}
		return strings.Join(tags, "")
	default:
		return ""
	}
}

func (p *Pix) generateAdditionalData() string {
	txid := "***"
	if len(p.params.id) > 0 {
		txid = p.params.id
	}
	return p.getValue(TAG_TXID, txid)
}

func (p *Pix) getCRC16(payload string) string {
	crc16 := crc.CalculateCRC(crc.CCITT, []byte(payload))
	return fmt.Sprintf("%04X", crc16)
}

func (p *Pix) FindAndReplaceCRC(payload string) string {
	m := regexp.MustCompile(`\w{4}$`)
	payload = m.ReplaceAllString(payload, "")
	return payload + p.getCRC16(payload)
}
