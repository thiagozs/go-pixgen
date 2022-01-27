# Go-PixGen

<p align="center"><img alt="pix-utils" src="https://raw.githubusercontent.com/thiagozs/go-pixgen/main/assets/logo-pix.png" width="128px" /></p>

generate and validate payments of Brazil Instant Payment System (Pix), making fast and simple to handle charges and proccess then in your project.

**UNDER DEVELOPMENT** *** Not ready for production

Example implementation.
```golang
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
```

## Contributing

Please contribute using [GitHub Flow](https://guides.github.com/introduction/flow). Create a branch, add commits, and [open a pull request](https://github.com/thiagozs/go-genpix/compare).

## Versioning and license

Our version numbers follow the [semantic versioning specification](http://semver.org/). You can see the available versions by checking the [tags on this repository](https://github.com/thiagozs/go-pixgen/tags). For more details about our license model, please take a look at the [LICENSE](LICENSE) file.

**2022**, thiagozs