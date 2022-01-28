package main

import (
	"fmt"

	"github.com/thiagozs/go-pixgen/pix"
)

func main() {
	opts := []pix.Options{
		pix.OptPixKey("11999821234"),
		pix.OptMerchantName("Thiago Zilli Sarmento"),
		pix.OptMerchantCity("Ararangua"),
		pix.OptKind(pix.STATIC),
	}

	p, err := pix.New(opts...)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := p.Validates(); err != nil {
		fmt.Println(err.Error())
		return
	}

	cpy := p.GenPayload()

	fmt.Printf("Copy and Paste: %s\n", cpy)

	// bts, err := p.GenQRCode()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Printf("QRCode: %b\n", bts)
}
