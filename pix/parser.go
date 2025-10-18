package pix

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/snksoft/crc"
)

var (
	templateTags = map[string]bool{
		TAG_MAI:             true,
		TAG_ADDITIONAL_DATA: true,
	}
)

// TLV represents an EMV tag-length-value entry.
type TLV struct {
	Tag     string
	Value   string
	Entries []*TLV
}

// MerchantAccount describes the parsed Merchant Account Information template.
type MerchantAccount struct {
	ID             string
	GUI            string
	PixKey         string
	AdditionalInfo string
	URL            string
	Raw            map[string]string
}

// AdditionalData captures parsed Additional Data Field Template values.
type AdditionalData struct {
	Raw  map[string]string
	TxID string
}

// ParsedPayload contains the structured Pix payload after parsing.
type ParsedPayload struct {
	Raw                     string
	PayloadFormatIndicator  string
	PointOfInitiationMethod string
	MerchantCategoryCode    string
	TransactionCurrency     string
	TransactionAmount       string
	CountryCode             string
	MerchantName            string
	MerchantCity            string
	CRC                     string
	MerchantAccounts        []MerchantAccount
	AdditionalDataField     AdditionalData
	Tags                    map[string]*TLV
}

// Kind returns the Pix kind inferred from the payload.
func (p ParsedPayload) Kind() PixKind {
	switch p.PointOfInitiationMethod {
	case "12":
		return DYNAMIC
	default:
		return STATIC
	}
}

// ParsePayload converts a Pix EMV payload string into a structured representation.
func ParsePayload(payload string) (*ParsedPayload, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, errors.New("payload must not be empty")
	}

	tlvs, err := parseTLVStream(payload)
	if err != nil {
		return nil, err
	}

	topLevel := make(map[string]*TLV, len(tlvs))
	for _, tlv := range tlvs {
		// keep the first occurrence for deterministic lookups
		if _, exists := topLevel[tlv.Tag]; !exists {
			topLevel[tlv.Tag] = tlv
		}
	}

	crcTLV, ok := topLevel[TAG_CRC]
	if !ok {
		return nil, errors.New("payload missing CRC tag (63)")
	}
	if len(crcTLV.Value) != 4 {
		return nil, errors.New("crc tag must have length 4")
	}

	expectedCRC := strings.ToUpper(crcTLV.Value)
	recalculatedCRC := strings.ToUpper(fmt.Sprintf("%04X", crc.CalculateCRC(crc.CCITT, []byte(payload[:len(payload)-4]))))
	if expectedCRC != recalculatedCRC {
		return nil, fmt.Errorf("crc mismatch: expected %s got %s", expectedCRC, recalculatedCRC)
	}

	result := &ParsedPayload{
		Raw:  payload,
		CRC:  expectedCRC,
		Tags: topLevel,
	}

	if v, ok := topLevel[TAG_INIT]; ok {
		result.PayloadFormatIndicator = v.Value
	}
	if v, ok := topLevel[TAG_INIT_METHOD]; ok {
		result.PointOfInitiationMethod = v.Value
	}
	if v, ok := topLevel[TAG_MCC]; ok {
		result.MerchantCategoryCode = v.Value
	}
	if v, ok := topLevel[TAG_TRANSACTION_CURRENCY]; ok {
		result.TransactionCurrency = v.Value
	}
	if v, ok := topLevel[TAG_TRANSACTION_AMOUNT]; ok {
		result.TransactionAmount = v.Value
	}
	if v, ok := topLevel[TAG_COUNTRY_CODE]; ok {
		result.CountryCode = v.Value
	}
	if v, ok := topLevel[TAG_MERCHANT_NAME]; ok {
		result.MerchantName = v.Value
	}
	if v, ok := topLevel[TAG_MERCHANT_CITY]; ok {
		result.MerchantCity = v.Value
	}

	var merchantAccounts []MerchantAccount
	for tag, tlv := range topLevel {
		if isMerchantAccountTag(tag) {
			account := MerchantAccount{
				ID:  tag,
				Raw: make(map[string]string),
			}

			for _, entry := range tlv.Entries {
				account.Raw[entry.Tag] = entry.Value
				switch entry.Tag {
				case TAG_MAI_GUI:
					account.GUI = entry.Value
				case TAG_MAI_PIXKEY:
					account.PixKey = entry.Value
				case TAG_MAI_INFO_ADD:
					account.AdditionalInfo = entry.Value
				case TAG_MAI_URL:
					account.URL = entry.Value
				}
			}
			merchantAccounts = append(merchantAccounts, account)
		}
	}
	result.MerchantAccounts = merchantAccounts

	if ad, ok := topLevel[TAG_ADDITIONAL_DATA]; ok {
		additional := AdditionalData{
			Raw: make(map[string]string),
		}
		for _, entry := range ad.Entries {
			additional.Raw[entry.Tag] = entry.Value
			if entry.Tag == TAG_TXID {
				additional.TxID = entry.Value
			}
		}
		result.AdditionalDataField = additional
	}

	if err := result.validateRequiredFields(); err != nil {
		return nil, err
	}

	return result, nil
}

// ValidatePayload parses and validates a Pix payload, returning an error if the payload
// does not conform to the Pix EMV rules enforced by this library.
func ValidatePayload(payload string) error {
	_, err := ParsePayload(payload)
	return err
}

func (p ParsedPayload) validateRequiredFields() error {
	if p.PayloadFormatIndicator == "" {
		return errors.New("payload format indicator (tag 00) is required")
	}
	if len(p.MerchantAccounts) == 0 {
		return errors.New("at least one merchant account information (tag 26-51) is required")
	}
	if p.CountryCode == "" {
		return errors.New("country code (tag 58) is required")
	}
	if p.MerchantName == "" {
		return errors.New("merchant name (tag 59) is required")
	}
	if p.MerchantCity == "" {
		return errors.New("merchant city (tag 60) is required")
	}
	return nil
}

func parseTLVStream(payload string) ([]*TLV, error) {
	var entries []*TLV
	cursor := 0

	for cursor < len(payload) {
		if len(payload[cursor:]) < 4 {
			return nil, fmt.Errorf("unexpected end of payload near index %d", cursor)
		}

		tag := payload[cursor : cursor+2]
		lengthField := payload[cursor+2 : cursor+4]
		length, err := strconv.Atoi(lengthField)
		if err != nil {
			return nil, fmt.Errorf("invalid length for tag %s: %w", tag, err)
		}

		start := cursor + 4
		end := start + length
		if end > len(payload) {
			return nil, fmt.Errorf("tag %s length exceeds payload size", tag)
		}

		value := payload[start:end]
		entry := &TLV{
			Tag:   tag,
			Value: value,
		}

		if shouldParseNested(tag) {
			nested, err := parseTLVStream(value)
			if err != nil {
				return nil, err
			}
			entry.Entries = nested
		}

		entries = append(entries, entry)
		cursor = end
	}

	return entries, nil
}

func shouldParseNested(tag string) bool {
	if templateTags[tag] {
		return true
	}
	// Dynamic merchant account templates are defined in EMV spec as tags 26-51.
	tagValue, err := strconv.Atoi(tag)
	if err != nil {
		return false
	}
	return tagValue >= 26 && tagValue <= 51
}

func isMerchantAccountTag(tag string) bool {
	tagValue, err := strconv.Atoi(tag)
	if err != nil {
		return false
	}
	return tagValue >= 26 && tagValue <= 51
}
