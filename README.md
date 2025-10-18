# Go-PixGen

<p align="center"><img alt="pix-utils" src="https://raw.githubusercontent.com/thiagozs/go-pixgen/main/assets/logo-pix.png" width="128px" /></p>

Gere e valide pagamentos do Sistema de Pagamentos Instantâneo (Pix) do Banco Central de forma simples. Além da biblioteca, este repositório inclui um CLI e um serviço REST para gerar payloads Pix estáticos ou dinâmicos, bem como seus códigos QR.

## Funcionalidades

- Geração de payloads Pix estáticos (copia e cola) e dinâmicos (com URL).
- CLI `pixgen` com comandos `generate` e `serve` para uso local ou como serviço.
- Serviço REST com `POST /pix` retornando payload + QR Code (base64) e `GET /healthz`.
- Geração de QR Code via `github.com/skip2/go-qrcode`.
- Utilitários de parsing EMV para inspeção de tags e metadados.
- Normalização de dados (chave Pix, valor, TxID) seguindo regras do BACEN.
- Resolução de Pix dinâmico suportando JSON, texto puro e tratamento de expiração.

## Instalação

```bash
go get github.com/thiagozs/go-pixgen
```

Requer Go 1.17 ou superior.

## Guia rápido

### CLI – gerar payload no terminal

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

A saída inclui o código copia-e-cola, campos relevantes e o QR Code em base64.

### Serviço REST

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

O endpoint `GET /healthz` retorna `200 OK` para checagens.

Exemplo com `curl` + `jq` para visualizar a resposta:

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

## Uso em código

### Gerar payload Pix programaticamente

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

## Destaques da API

- `pix.New(opts...) (*pix.Pix, error)` – cria um gerador Pix configurável.
- `(*Pix).GenPayload()` – retorna o payload EMV e o mantém em cache para `GenQRCode()`.
- `pix.ParsePayload(string) (*ParsedPayload, error)` – faz o parsing do payload e valida o CRC.
- `(*Pix).FetchDynamicPayload(ctx, client)` – baixa, valida e parseia payloads dinâmicos remotos.
- `(*Pix).Validates()` – valida os parâmetros (chaves, tamanho de campos, etc.).

### Formatos aceitos de chave Pix

- EVP (UUID) – normalizado para minúsculo.
- Telefone (`+55DDDNÚMERO`) – aceita apenas dígitos com ou sem `+55`.
- CPF/CNPJ – somente dígitos, com validação dos dígitos verificadores.
- E-mail – validado via `net/mail`.

### Valor e TxID

- Valor suporta até 13 dígitos antes da vírgula e 2 casas decimais (`9999999999999.99` limite).
- TxID permite letras, números, `.` e `-` até 35 caracteres e é armazenado em maiúsculas.

### Busca de payload dinâmico

`FetchDynamicPayload` entende:

- payload EMV bruto (`text/*`).
- JSON com campos como `pixCopyPaste`, `pix`, `payload`, `pixCopiaECola` etc.
- Campos de expiração (`expiresAt`, `expiration`, `expiry`) em formatos RFC3339.

Para testes locais, URLs HTTP são aceitas quando o host é `localhost` ou `127.*`; para produção exige HTTPS.

## Build & Docker

- `make build`, `make test`, `make fmt`
- `make cli` compila o binário em `bin/pixgen`
- `make run ARGS="serve --addr :8080"` executa o servidor REST local
- `make docker-build && make docker-run` constroem e sobem a imagem (porta 8080)

## Roadmap

- [x] Gerar pagamentos a partir de parâmetros
  - [x] Estático
  - [x] Dinâmico
- [x] Parse e validação de códigos EMV
- [x] Exportar pagamento para imagem
- [x] Exportar pagamento para código EMV
- [x] Buscar, parsear e validar payloads dinâmicos remotos
  - [x] Verificar se já expirou
- [x] Melhorar testes
- [ ] Documentação completa (métodos, parâmetros, exemplos)
- [x] Testes para pagamentos dinâmicos
- [x] Ferramentas CLI para inspeção de payload
- [ ] Endurecimento da API REST e opções de autenticação

## Contribuindo

Siga o [GitHub Flow](https://guides.github.com/introduction/flow): crie uma branch, faça commits e abra um [pull request](https://github.com/thiagozs/go-genpix/compare).

## Versionamento e licença

Utilizamos [SemVer](http://semver.org/). Confira as versões em [tags](https://github.com/thiagozs/go-pixgen/tags). Licença em [LICENSE](LICENSE).

**2022**, Thiago Zilli Sarmento :heart:
