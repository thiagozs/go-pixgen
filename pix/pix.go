package pix

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/snksoft/crc"
	"github.com/thiagozs/go-pixgen/qrcode"
)

type Pix struct {
	params *OptionsParams
}

func New(opts ...Options) (*Pix, error) {
	p := &OptionsParams{}
	for _, op := range opts {
		if err := op(p); err != nil {
			return nil, err
		}
	}
	px := &Pix{params: p}
	if err := px.Validates(); err != nil {
		return nil, err
	}
	return px, nil
}

// GenPayload builds the EMV-compliant Pix payload string
func (p *Pix) GenPayload() string {
	var tags []string

	initMethod := map[PixKind]string{STATIC: "11", DYNAMIC: "12"}[p.params.GetKind()]
	tags = append(tags,
		p.tlv(TAG_INIT, "01"),
		p.tlv(TAG_INIT_METHOD, initMethod),
		p.tlv(TAG_MAI, p.generateMAI()),
		p.tlv(TAG_MCC, "0000"),
		p.tlv(TAG_TRANSACTION_CURRENCY, "986"),
	)

	if amt := p.params.GetAmount(); amt != "" {
		tags = append(tags, p.tlv(TAG_TRANSACTION_AMOUNT, amt))
	}

	tags = append(tags,
		p.tlv(TAG_COUNTRY_CODE, "BR"),
		p.tlv(TAG_MERCHANT_NAME, p.params.GetMerchantName()),
		p.tlv(TAG_MERCHANT_CITY, p.params.GetMerchantCity()),
		p.tlv(TAG_ADDITIONAL_DATA, p.generateAdditionalData()),
		p.tlv(TAG_CRC, "0000"),
	)

	payload := strings.Join(tags, "")
	payload = p.replaceCRC(payload)
	p.params.SetQRCodeContent(payload)
	return payload
}

// Generate QRCode bytes from payload
func (p *Pix) GenQRCode() ([]byte, error) {
	if p.params.GetQRCodeContent() == "" {
		_ = p.GenPayload()
	}
	return qrcode.New(qrcode.QRCodeOptions{
		Size:    p.params.GetQRCodeSize(),
		Content: p.params.GetQRCodeContent(),
	})
}

// Private helpers
func (p *Pix) tlv(tag, content string) string {
	return fmt.Sprintf("%s%02d%s", tag, len(content), content)
}

func (p *Pix) generateMAI() string {
	switch p.params.GetKind() {
	case STATIC:
		parts := []string{
			p.tlv(TAG_MAI_GUI, BC_GUI),
			p.tlv(TAG_MAI_PIXKEY, p.params.GetPixKey()),
		}
		if add := p.params.GetAdditionalInfo(); add != "" {
			parts = append(parts, p.tlv(TAG_MAI_INFO_ADD, add))
		} else if desc := p.params.GetDescription(); desc != "" {
			parts = append(parts, p.tlv(TAG_MAI_INFO_ADD, desc))
		}
		return strings.Join(parts, "")
	case DYNAMIC:
		return strings.Join([]string{
			p.tlv(TAG_MAI_GUI, BC_GUI),
			p.tlv(TAG_MAI_URL, p.params.GetUrl()),
		}, "")
	default:
		return ""
	}
}

func (p *Pix) generateAdditionalData() string {
	txid := p.params.GetTxId()
	if txid == "" {
		txid = "***"
	}
	return p.tlv(TAG_TXID, txid)
}

func (p *Pix) crc16(payload string) string {
	return fmt.Sprintf("%04X", crc.CalculateCRC(crc.CCITT, []byte(payload)))
}

func (p *Pix) replaceCRC(payload string) string {
	re := regexp.MustCompile(`\w{4}$`)
	payload = re.ReplaceAllString(payload, "")
	return payload + p.crc16(payload)
}
