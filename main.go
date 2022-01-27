package main

import (
	"fmt"

	"github.com/thiagozs/go-pixgen/pix"
)

func main() {

	opts := []pix.Options{
		pix.OptPixKey("11955555555"),
		pix.OptDescription("Teste"),
		pix.OptMerchantName("Thiago Zilli Sarmento"),
		pix.OptMerchantCity("Ararangua"),
		pix.OptAmount("1.00"),
		pix.OptAditionalInfo("gerado por go-pixgen"),
		pix.OptKind(pix.STATIC),
	}

	p, err := pix.New(opts...)
	if err != nil {
		panic(err)
	}

	if err := p.Validates(); err != nil {
		fmt.Println(err.Error())
		return
	}

	cpy := p.GenPayload()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Copy and Paste: %s\n", cpy)

	bts, err := p.GenQRCode()
	if err != nil {
		panic(err)
	}
	fmt.Printf("QRCode: %b\n", bts)

}
