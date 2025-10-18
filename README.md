# Go-PixGen

<p align="center"><img alt="pix-utils" src="https://raw.githubusercontent.com/thiagozs/go-pixgen/main/assets/logo-pix.png" width="128px" /></p>

Generate and validate payments of Brazil Instant Payment System (Pix), making fast and simple to handle charges and proccess then in your project. Além da lib, o repositório oferece CLI e serviço REST para gerar payloads Pix estáticos ou dinâmicos, incluindo QR Code.

## Features

- Static (copy & paste) and dynamic (URL-based) Pix payload generation.
- CLI `pixgen` (comandos `generate` e `serve`) para uso local ou como serviço.
- REST service (`POST /pix`) retornando payload + QR Code (base64) e `GET /healthz`.
- QR Code byte encoding built-in via `github.com/skip2/go-qrcode`.
- EMV parsing helpers to inspect payload tags and metadata.
- Field normalization (Pix keys, amount, TXID) with BACEN-aligned validation.
- Dynamic Pix resolver with JSON/plain-text support, expiration awareness and mock friendly interface.

## Installation

```bash
go get github.com/thiagozs/go-pixgen
```

Go 1.17 or newer is required.

## Quick Start

### CLI – gerar payload no stdout

```bash
make cli

bin/pixgen generate \
  --kind static \
  --key +5511999999999 \
  --merchant-name "Thiago Zilli Sarmento" \
  --merchant-city ARARANGUA \
  --amount 10.00 \
  --description "Pedido 123" \
  --txid PEDIDO-123
```

Saída inclui o código copia-e-cola, campos relevantes e o QR Code em base64.

### REST service

```bash
make run ARGS="serve --addr :8080"

curl -X POST http://localhost:8080/pix \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "static",
    "pixKey": "+5511999999999",
    "merchantName": "Thiago Zilli Sarmento",
    "merchantCity": "ARARANGUA",
    "amount": "10.00",
    "description": "Pedido 123",
    "txid": "PEDIDO-123"
  }'
```

```json
{
  "payload": "000201...",
  "qrCode": "iVBORw0KGgoAAA...",
  "kind": "STATIC",
  "txid": "PEDIDO-123"
}
```

Endpoint `GET /healthz` retorna `200 OK` para checagens.

Exemplo rápido com `curl` + `jq` para visualizar a resposta formatada:

```bash
curl -sS -X POST http://localhost:8080/pix \
  -H 'Content-Type: application/json' \
  -d '{
    "kind": "dynamic",
    "url": "https://example.com/api/pix/abc",
    "merchantName": "Fulano de Tal",
    "merchantCity": "CURITIBA",
    "amount": "25.00",
    "txid": "DINAMICO-001"
  }' | jq
```

## Usage

### Generate a Pix payload programaticamente

```golang
opts := []pix.Options{
	pix.OptPixKey("+5511955555555"),
	pix.OptDescription("Teste"),
	pix.OptMerchantName("Thiago Zilli Sarmento"),
	pix.OptMerchantCity("Ararangua"),
	pix.OptAmount("1.00"),
	pix.OptAdditionalInfo("gerado por go-pixgen"),
	pix.OptKind(pix.STATIC),
}

p, err := pix.New(opts...)
if err != nil {
	fmt.Println(err.Error())
	return
}

cpy := p.GenPayload()
fmt.Printf("Copy and Paste: %s\n", cpy)

qrPNG, err := p.GenQRCode()
if err != nil {
	fmt.Println(err.Error())
	return
}

fmt.Printf("QRCode bytes: %d\n", len(qrPNG))

parsed, err := pix.ParsePayload(cpy)
if err != nil {
	fmt.Println(err.Error())
	return
}

fmt.Printf("Kind: %s | TxID: %s\n", parsed.Kind(), parsed.AdditionalDataField.TxID)
```

### Fetch a dynamic Pix payload

```golang
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

dynamicOpts := []pix.Options{
	pix.OptKind(pix.DYNAMIC),
	pix.OptUrl("https://my-provider.example/api/pix/123"),
	pix.OptMerchantName("Fulano de Tal"),
	pix.OptMerchantCity("CURITIBA"),
}

dyn, err := pix.New(dynamicOpts...)
if err != nil {
	fmt.Println(err)
	return
}

payload, err := dyn.FetchDynamicPayload(ctx, nil)
if err != nil {
	fmt.Println(err)
	return
}

fmt.Printf("Remote Pix expires at: %v\n", payload.ExpiresAt)
```

## API Highlights

- `pix.New(opts...) (*pix.Pix, error)` - build a Pix generator using functional options.
- `(*Pix).GenPayload()` - returns the EMV string and caches it for `GenQRCode()`.
- `pix.ParsePayload(string) (*ParsedPayload, error)` - converts an EMV payload back into structured fields and validates the CRC.
- `(*Pix).FetchDynamicPayload(ctx, client)` - downloads a dynamic Pix payload, parses it and checks expiration. Works with `http.Client` stubs for tests.
- `(*Pix).Validates()` - called automatically by `pix.New`, and can be invoked manually to re-check mutated parameters.

### Supported Pix key formats

- Random EVP UUID (case-normalized to lowercase).
- Phone (`+55DDDUSER`), accepting raw digits with optional `+55` prefix.
- CPF/CNPJ (digits only, validated with check digits).
- Email addresses (validated using `net/mail`).

### Amount & TxID

- Amount accepts up to 13 digits before the separator and up to 2 decimals (`9999999999999.99` cap).
- TxID allows uppercase/lowercase letters, digits, `.` and `-` up to 35 characters. Value is uppercased when stored.

### Dynamic payload fetching

`FetchDynamicPayload` understands responses as:

- Raw EMV payload (any `text/*` content type).
- JSON containing fields `pixCopyPaste`, `pix`, `payload`, `pixCopiaECola`, etc.
- Optional expiration keys such as `expiresAt`, `expiration`, `expiry`. Parsed values support RFC3339 and similar layouts.

Local testing (HTTP URLs) is accepted when targeting `localhost` or `127.*`. Production URLs must be HTTPS.

## Build & Docker

- `make build`, `make test`, `make fmt`
- `make cli` compila o binário em `bin/pixgen`
- `make run ARGS="serve --addr :8080"` executa o servidor REST local
- `make docker-build && make docker-run` constroem e sobem a imagem (porta 8080)

## Roadmap

- [x] Generate payments based on parameters
  - [x] Static
  - [x] Dynamic
- [x] Parse and validate EMV Codes
- [x] Export generated/parsed payment to Image
- [x] Export generated/parsed payment to EMV Code
- [x] Fetch, parse and validate remote payloads from dynamic payments
  - [x] Verify if has already expired
- [x] Improve tests
- [ ] Documentation with all methods, parameters and more examples
- [x] Add dynamic payment tests
- [x] CLI tooling for payload inspection
- [ ] REST API hardening & auth helpers

## Contributing

Please contribute using [GitHub Flow](https://guides.github.com/introduction/flow). Create a branch, add commits, and [open a pull request](https://github.com/thiagozs/go-genpix/compare).

## Versioning and license

Our version numbers follow the [semantic versioning specification](http://semver.org/). You can see the available versions by checking the [tags on this repository](https://github.com/thiagozs/go-pixgen/tags). For more details about our license model, please take a look at the [LICENSE](LICENSE) file.

**2022**, Thiago Zilli Sarmento :heart:
