package pix

// EMV schema tags conforme BACEN Pix Manual v2.9.0
const (
	// Payload Format Indicator
	TAG_INIT = "00" // Identifica o formato do payload Pix (valor fixo "01")

	// Point of Initiation Method
	TAG_INIT_METHOD = "01" // 11 = Estático | 12 = Dinâmico

	// Merchant Account Information (subtags 00-51)
	TAG_MAI          = "26" // Grupo de dados do recebedor (Merchant Account Information)
	TAG_MAI_GUI      = "00" // Identificador do domínio do arranjo Pix (obrigatório: "br.gov.bcb.pix")
	TAG_MAI_PIXKEY   = "01" // Chave Pix (telefone, e-mail, CPF/CNPJ ou EVP)
	TAG_MAI_INFO_ADD = "02" // Informação adicional exibida ao pagador (até 72 caracteres)
	TAG_MAI_FSS      = "03" // ISPB do facilitador de saque (opcional, novo na v2.9.0)
	TAG_MAI_URL      = "25" // URL para QR Dinâmico (sem protocolo, máx. 77 caracteres)

	// Merchant Category Code (opcional)
	TAG_MCC = "52" // MCC padrão "0000" se não houver código de categoria

	// Transaction Currency
	TAG_TRANSACTION_CURRENCY = "53" // Código da moeda ISO4217 (BRL = "986")

	// Transaction Amount
	TAG_TRANSACTION_AMOUNT = "54" // Valor monetário, formato: d{1,10}.d{2}

	// Country Code
	TAG_COUNTRY_CODE = "58" // Código ISO do país (BR = Brasil)

	// Merchant Name e City
	TAG_MERCHANT_NAME = "59" // Nome do recebedor (até 25 caracteres, sem acentuação)
	TAG_MERCHANT_CITY = "60" // Cidade do recebedor (até 15 caracteres, sem acentuação)

	// Additional Data Field Template
	TAG_ADDITIONAL_DATA = "62" // Grupo de dados adicionais (TxID, informações extras)

	// Additional Data Subtags
	TAG_TXID = "05" // Identificador da transação (estático: 1-25 alfanumérico ou "***"; dinâmico usa "***")

	// CRC16 Checksum
	TAG_CRC = "63" // Checksum de 4 dígitos hexadecimais (CRC-CCITT XModem)

	// Domínio oficial do BACEN para QR Pix
	BC_GUI = "br.gov.bcb.pix"
)
