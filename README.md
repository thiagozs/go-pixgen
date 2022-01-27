# Go-PixGen

<p align="center"><img alt="pix-utils" src="https://raw.githubusercontent.com/thiagozs/go-pixgen/main/assets/logo-pix.png" width="128px" /></p>

Generate ~~and validate~~ payments of Brazil Instant Payment System (Pix), making fast and simple to handle charges and proccess then in your project.

**UNDER DEVELOPMENT** *** Becareful ***

Example implementation.
```golang
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
		fmt.Println(err.Error())
    return
	}

	if err := p.Validates(); err != nil {
		fmt.Println(err.Error())
		return
	}

	cpy := p.GenPayload()

	fmt.Printf("Copy and Paste: %s\n", cpy)

	bts, err := p.GenQRCode()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("QRCode: %b\n", bts)
```

## Roadmap

- [x] Generate payments based on parameters
  - [x] Static
  - [x] Dynamic
- [ ] Parse and validate EMV Codes
- [x] Export generated/parsed payment to Image
- [x] Export generated/parsed payment to EMV Code
- [ ] Fetch, parse and validate remote payloads from dynamic payments
  - [ ] Verify if has already expired
- [ ] Improve tests
- [ ] Doccumentation with all methods, parameters and some examples
- [ ] Add dynamic payment tests

## Contributing

Please contribute using [GitHub Flow](https://guides.github.com/introduction/flow). Create a branch, add commits, and [open a pull request](https://github.com/thiagozs/go-genpix/compare).

## Versioning and license

Our version numbers follow the [semantic versioning specification](http://semver.org/). You can see the available versions by checking the [tags on this repository](https://github.com/thiagozs/go-pixgen/tags). For more details about our license model, please take a look at the [LICENSE](LICENSE) file.

**2022**, thiagozs