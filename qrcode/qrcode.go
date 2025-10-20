package qrcode

import (
	"bytes"
	"errors"
	"strings"

	"github.com/mdp/qrterminal"
	"github.com/skip2/go-qrcode"
)

type QRCodeOptions struct {
	Content string
	Size    int
}

func New(options QRCodeOptions) ([]byte, error) {
	if options.Size == 0 {
		options.Size = 256
	}
	return qrcode.Encode(options.Content, qrcode.Medium, options.Size)
}

type ASCIIOptions struct {
	Content      string
	BlackChar    string
	WhiteChar    string
	Scale        int
	QuietZone    bool
	QuietZoneSet bool
}

func NewASCII(opts ASCIIOptions) (string, error) {
	if strings.TrimSpace(opts.Content) == "" {
		return "", errors.New("qrcode: content must not be empty")
	}

	var buf bytes.Buffer

	cfg := qrterminal.Config{
		Level:  qrterminal.M,
		Writer: &buf,
	}

	if opts.QuietZoneSet {
		if opts.QuietZone {
			cfg.QuietZone = qrterminal.QUIET_ZONE
		} else {
			cfg.QuietZone = 1
		}
	} else {
		cfg.QuietZone = 1
	}

	scale := opts.Scale
	if scale < 1 {
		scale = 1
	}

	black := opts.BlackChar
	if black == "" {
		black = qrterminal.BLACK
	}
	white := opts.WhiteChar
	if white == "" {
		white = qrterminal.WHITE
	}

	cfg.BlackChar = strings.Repeat(black, scale)
	cfg.WhiteChar = strings.Repeat(white, scale)

	qrterminal.GenerateWithConfig(opts.Content, cfg)

	return strings.TrimRight(buf.String(), "\n"), nil
}
