package pix

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/snksoft/crc"
	"github.com/thiagozs/go-pixgen/qrcode"
)

// Pix representa o construtor principal do QR Pix
type Pix struct {
	params *OptionsParams
}

// New cria uma nova instância de Pix
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

// GenPayload monta o payload EMV-compliant Pix Copia e Cola
func (p *Pix) GenPayload() (string, error) {
	var tags []string

	initMethod := map[PixKind]string{STATIC: "11", DYNAMIC: "12"}[p.params.GetKind()]

	mai, err := p.generateMAI()
	if err != nil {
		return "", err
	}

	tags = append(tags,
		p.tlv(TAG_INIT, "01"),
		p.tlv(TAG_INIT_METHOD, initMethod),
		p.tlv(TAG_MAI, mai),
		p.tlv(TAG_MCC, "0000"),
		p.tlv(TAG_TRANSACTION_CURRENCY, "986"),
	)

	// Transaction Amount deve ter 2 casas decimais
	var moneyRe = regexp.MustCompile(`^\d{1,10}\.\d{2}$`)

	if amt := p.params.GetAmount(); amt != "" && moneyRe.MatchString(amt) {
		tags = append(tags, p.tlv(TAG_TRANSACTION_AMOUNT, amt))
	}

	tags = append(tags,
		p.tlv(TAG_COUNTRY_CODE, "BR"),
		p.tlv(TAG_MERCHANT_NAME, normalizeChars(p.params.GetMerchantName())),
		p.tlv(TAG_MERCHANT_CITY, normalizeChars(p.params.GetMerchantCity())),
	)

	additionalData, err := p.generateAdditionalData()
	if err != nil {
		return "", err
	}

	tags = append(tags,
		p.tlv(TAG_ADDITIONAL_DATA, additionalData),
		p.tlv(TAG_CRC, "0000"),
	)

	payload := strings.Join(tags, "")
	payload = sanitizePayload(payload)
	payload = p.replaceCRC(payload)
	p.params.SetQRCodeContent(payload)
	return payload, nil
}

// GenQRCode gera o QR Code em bytes
func (p *Pix) GenQRCode() ([]byte, error) {
	if p.params.GetQRCodeContent() == "" {
		if _, err := p.GenPayload(); err != nil {
			return nil, err
		}
	}
	size := p.params.GetQRCodeSize()
	if size == 0 {
		size = 256
	}
	return qrcode.New(qrcode.QRCodeOptions{
		Size:    size,
		Content: p.params.GetQRCodeContent(),
	})
}

// GenQRCodeASCII renderiza o QR Code em arte ASCII para uso no terminal.
func (p *Pix) GenQRCodeASCII() (string, error) {
	if p.params.GetQRCodeContent() == "" {
		if _, err := p.GenPayload(); err != nil {
			return "", err
		}
	}

	scale := p.params.GetASCIIQrScale()
	if scale <= 0 {
		scale = 1
	}

	quiet := false
	if p.params.HasASCIIQuietZone() {
		quiet = p.params.GetASCIIQuietZone()
	}

	return qrcode.NewASCII(qrcode.ASCIIOptions{
		Content:      p.params.GetQRCodeContent(),
		Scale:        scale,
		BlackChar:    p.params.GetASCIIQrBlack(),
		WhiteChar:    p.params.GetASCIIQrWhite(),
		QuietZone:    quiet,
		QuietZoneSet: true,
	})
}

// -------- Helpers --------

// tlv gera um campo TLV formatado
func (p *Pix) tlv(tag, content string) string {
	content = strings.TrimSpace(content)
	return fmt.Sprintf("%s%02d%s", tag, len(content), content)
}

// generateMAI monta o bloco Merchant Account Information (26-51)
func (p *Pix) generateMAI() (string, error) {
	gui := p.tlv(TAG_MAI_GUI, BC_GUI)
	totalLen := len(gui)

	var parts []string
	parts = append(parts, gui)

	if p.params.GetKind() == DYNAMIC {
		rawURL := strings.TrimSpace(p.params.GetUrl())
		if rawURL == "" {
			return "", fmt.Errorf("dynamic pix requires url for MAI")
		}

		urlNoScheme := stripURLScheme(rawURL)
		if urlNoScheme == "" {
			return "", fmt.Errorf("dynamic pix url must not be empty after stripping scheme")
		}
		if len(urlNoScheme) > 77 {
			return "", fmt.Errorf("dynamic pix url must be at most 77 characters without scheme")
		}

		urlTLV := p.tlv(TAG_MAI_URL, urlNoScheme)
		if totalLen+len(urlTLV) > 99 {
			return "", fmt.Errorf("dynamic pix merchant account exceeds 99 characters")
		}
		parts = append(parts, urlTLV)
		return strings.Join(parts, ""), nil
	}

	key := strings.TrimSpace(p.params.GetPixKey())
	keyTLV := p.tlv(TAG_MAI_PIXKEY, key)
	if totalLen+len(keyTLV) > 99 {
		return "", fmt.Errorf("pix key length exceeds EMV 99 character limit")
	}
	parts = append(parts, keyTLV)
	totalLen += len(keyTLV)

	info := normalizeChars(p.params.GetAdditionalInfo())
	if info == "" {
		if desc := strings.TrimSpace(p.params.GetDescription()); desc != "" {
			info = normalizeChars(desc)
		}
	}

	if info != "" {
		// Limite de 72 caracteres para informação adicional (valor)
		if len(info) > 72 {
			info = info[:72]
		}

		infoTLVPrefixLen := len(TAG_MAI_INFO_ADD) + 2 // tag + length indicator
		remaining := 99 - (totalLen + infoTLVPrefixLen)
		if remaining > 72 {
			remaining = 72
		}

		if remaining <= 0 {
			// Sem espaço para incluir info adicional
		} else {
			if len(info) > remaining {
				info = info[:remaining]
			}
			infoTLV := p.tlv(TAG_MAI_INFO_ADD, info)
			if totalLen+len(infoTLV) <= 99 {
				parts = append(parts, infoTLV)
				totalLen += len(infoTLV)
			}
		}
	}

	return strings.Join(parts, ""), nil
}

func (p *Pix) generateAdditionalData() (string, error) {
	if p.params.GetKind() == DYNAMIC {
		return p.tlv(TAG_TXID, "***"), nil
	}

	txid := strings.TrimSpace(p.params.GetTxId())
	if txid == "" {
		txid = "***"
	} else {
		txid = strings.ToUpper(txid)
	}

	return p.tlv(TAG_TXID, txid), nil
}

// replaceCRC calcula e substitui o CRC16 corretamente
func (p *Pix) replaceCRC(payload string) string {
	re := regexp.MustCompile(`\w{4}$`)
	payload = sanitizePayload(re.ReplaceAllString(payload, ""))
	crcValue := fmt.Sprintf("%04X", crc.CalculateCRC(crc.CCITT, []byte(payload)))
	return payload + crcValue
}

// -------- Normalização --------

// sanitizePayload remove caracteres invisíveis e espaços extras
func sanitizePayload(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.TrimSpace(s)
	return s
}

// normalizeChars remove acentos e caracteres fora de ASCII básico
func normalizeChars(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä', 'Á', 'À', 'Ã', 'Â', 'Ä':
			return 'A'
		case 'é', 'è', 'ê', 'ë', 'É', 'È', 'Ê', 'Ë':
			return 'E'
		case 'í', 'ì', 'î', 'ï', 'Í', 'Ì', 'Î', 'Ï':
			return 'I'
		case 'ó', 'ò', 'õ', 'ô', 'ö', 'Ó', 'Ò', 'Õ', 'Ô', 'Ö':
			return 'O'
		case 'ú', 'ù', 'û', 'ü', 'Ú', 'Ù', 'Û', 'Ü':
			return 'U'
		case 'ç', 'Ç':
			return 'C'
		case 'ñ', 'Ñ':
			return 'N'
		default:
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}
	}, s)
	return strings.ToUpper(s)
}

func stripURLScheme(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	u, err := url.Parse(trimmed)
	if err == nil && u.Host != "" {
		result := u.Host
		if path := u.EscapedPath(); path != "" {
			result += path
		}
		if u.RawQuery != "" {
			result += "?" + u.RawQuery
		}
		if u.Fragment != "" {
			result += "#" + u.Fragment
		}
		return result
	}

	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	return trimmed
}
