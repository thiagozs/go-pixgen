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
		pix.OptAmount("0.00"),
		pix.OptKind(pix.STATIC),
	}

	p, err := pix.New(opts...)
	if err != nil {
		panic(err)
	}
	cpy := p.GenPayload()
	if err != nil {
		panic(err)
	}
	fmt.Println(cpy)

}
