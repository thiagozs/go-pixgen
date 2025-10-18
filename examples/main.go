package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/thiagozs/go-pixgen/pix"
)

func main() {
	if err := runStaticExample(); err != nil {
		fmt.Printf("static example error: %v\n", err)
	}

	fmt.Println()

	if err := runDynamicExample(); err != nil {
		fmt.Printf("dynamic example error: %v\n", err)
	}
}

func runStaticExample() error {
	fmt.Println("== Static Pix Example ==")

	opts := []pix.Options{
		pix.OptPixKey("+5511999999999"),
		pix.OptMerchantName("Thiago Zilli Sarmento"),
		pix.OptMerchantCity("ARARANGUA"),
		pix.OptKind(pix.STATIC),
		pix.OptAmount("125.50"),
		pix.OptDescription("Pedido #123"),
		pix.OptAdditionalInfo("Gerado com go-pixgen"),
		pix.OptTxId("PEDIDO-123"),
	}

	p, err := pix.New(opts...)
	if err != nil {
		return fmt.Errorf("create pix: %w", err)
	}

	payload := p.GenPayload()
	fmt.Printf("Copy & Paste payload: %s\n", payload)

	qrBytes, err := p.GenQRCode()
	if err != nil {
		return fmt.Errorf("generate qrcode: %w", err)
	}
	fmt.Printf("QRCode bytes generated: %d\n", len(qrBytes))

	parsed, err := pix.ParsePayload(payload)
	if err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	fmt.Printf("Parsed kind: %s\n", parsed.Kind())
	fmt.Printf("Parsed TxID: %s\n", parsed.AdditionalDataField.TxID)

	return nil
}

func runDynamicExample() error {
	fmt.Println("== Dynamic Pix Example ==")

	server := newMockDynamicServer()
	defer server.Close()

	opts := []pix.Options{
		pix.OptKind(pix.DYNAMIC),
		pix.OptUrl(server.URL),
		pix.OptMerchantName("Fulano de Tal"),
		pix.OptMerchantCity("CURITIBA"),
	}

	p, err := pix.New(opts...)
	if err != nil {
		return fmt.Errorf("create dynamic pix: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dynamicPayload, err := p.FetchDynamicPayload(ctx, server.Client())
	if err != nil {
		return fmt.Errorf("fetch dynamic payload: %w", err)
	}

	fmt.Printf("Remote payload (first 40 chars): %.40s...\n", dynamicPayload.Raw)
	if dynamicPayload.ExpiresAt != nil {
		fmt.Printf("Expires at: %s\n", dynamicPayload.ExpiresAt.Format(time.RFC3339))
	}
	if dynamicPayload.Parsed != nil {
		fmt.Printf("Remote kind inferred: %s\n", dynamicPayload.Parsed.Kind())
		fmt.Printf("Remote merchant: %s - %s\n", dynamicPayload.Parsed.MerchantName, dynamicPayload.Parsed.MerchantCity)
	}

	return nil
}

func newMockDynamicServer() *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		opts := []pix.Options{
			pix.OptKind(pix.DYNAMIC),
			pix.OptUrl("https://provider.example/api/pix/ABC123"),
			pix.OptMerchantName("Fulano de Tal"),
			pix.OptMerchantCity("CURITIBA"),
			pix.OptAmount("50.00"),
			pix.OptTxId("ABC123"),
		}

		payloadPix, err := pix.New(opts...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		payload := payloadPix.GenPayload()
		resp := map[string]string{
			"pixCopyPaste": payload,
			"expiresAt":    time.Now().Add(5 * time.Minute).Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}
