package qrcode

import (
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
