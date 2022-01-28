package pix

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/snksoft/crc"
	"github.com/thiagozs/go-pixgen/qrcode"
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
			p.getValue(TAG_INIT_METHOD, "11"),
			p.getValue(TAG_MAI, p.generateMAI()),
			p.getValue(TAG_MCC, "0000"),
			p.getValue(TAG_TRANSACTION_CURRENCY, "986")}
		if len(p.params.amount) > 0 {
			tags = append(tags, p.getValue(TAG_TRANSACTION_AMOUNT, p.params.amount))
		}
		tags = append(tags, p.getValue(TAG_COUNTRY_CODE, "BR"),
			p.getValue(TAG_MERCHANT_NAME, p.params.merchant.name),
			p.getValue(TAG_MERCHANT_CITY, p.params.merchant.city),
			p.getValue(TAG_ADDITIONAL_DATA, p.generateAdditionalData()),
			p.getValue(TAG_CRC, "0000"),
		)
	case DYNAMIC:
		tags = []string{
			p.getValue(TAG_INIT, "01"),
			p.getValue(TAG_INIT_METHOD, "11"),
			p.getValue(TAG_MAI, p.generateMAI()),
			p.getValue(TAG_MCC, "0000"),
			p.getValue(TAG_TRANSACTION_CURRENCY, "986"),
		}
		if len(p.params.amount) > 0 {
			tags = append(tags, p.getValue(TAG_TRANSACTION_AMOUNT, p.params.amount))
		}
		tags = append(tags, p.getValue(TAG_COUNTRY_CODE, "BR"),
			p.getValue(TAG_MERCHANT_NAME, p.params.merchant.name),
			p.getValue(TAG_MERCHANT_CITY, p.params.merchant.city),
			p.getValue(TAG_ADDITIONAL_DATA, p.generateAdditionalData()),
			p.getValue(TAG_CRC, "0000"),
		)
	}

	payload := strings.Join(tags, "")

	payload = p.FindAndReplaceCRC(payload)

	p.params.SetQRCodeContent(payload)

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
		}
		if len(p.params.aditionalInfo) > 0 {
			tags = append(tags, p.getValue(TAG_MAI_INFO_ADD, p.params.aditionalInfo))
		}
		return strings.Join(tags, "")
	case DYNAMIC:
		tags := []string{
			p.getValue(TAG_MAI_GUI, BC_GUI),
			p.getValue(TAG_MAI_URL, p.params.url),
		}
		return strings.Join(tags, "")
	default:
		return ""
	}
}

func (p *Pix) generateAdditionalData() string {
	txid := "***"
	if len(p.params.txId) > 0 {
		txid = p.params.txId
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

func (p *Pix) Validates() error {
	if p.params.pixKey == "" {
		return errors.New("pixkey must not be empty")
	}

	if p.params.merchant.name == "" {
		return errors.New("name must not be empty")
	}

	if p.params.merchant.city == "" {
		return errors.New("city must not be empty")
	}

	if utf8.RuneCountInString(p.params.merchant.name) > 25 {
		return errors.New("name must be at least 25 characters long")
	}

	if utf8.RuneCountInString(p.params.merchant.city) > 15 {
		return errors.New("city must be at least 15 characters long")
	}

	return nil
}

func (p *Pix) GenQRCode() ([]byte, error) {
	return qrcode.New(qrcode.QRCodeOptions{
		Size:    p.params.GetQRCodeSize(),
		Content: p.params.GetQRCodeContent(),
	})
}
