package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/thiagozs/go-pixgen/pix"
)

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pixgen",
		Short: "Pix payload generator CLI",
		Run: func(cmd *cobra.Command, args []string) {

			cmd.Usage()
		},
	}

	cmd.AddCommand(newGenerateCmd(), newServeCmd())
	return cmd
}

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Pix payload and QRCode locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := collectPixParams(cmd)
			if err != nil {
				return err
			}

			payload, qr, asciiQR, parsed, err := buildPix(params)
			if err != nil {
				return err
			}

			fmt.Printf("Copy and Paste: %s\n", payload)
			if parsed != nil {
				fmt.Printf("Kind: %s\n", parsed.Kind())
				fmt.Printf("Merchant: %s (%s)\n", parsed.MerchantName, parsed.MerchantCity)
				fmt.Printf("TxID: %s\n", parsed.AdditionalDataField.TxID)
			}
			fmt.Printf("QR Code (base64): %s\n", base64.StdEncoding.EncodeToString(qr))
			if asciiQR != "" {
				fmt.Println("QR Code (ASCII):")
				fmt.Println(asciiQR)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.String("kind", "static", "Pix kind: static or dynamic")
	flags.String("key", "", "Pix key (required for static)")
	flags.String("url", "", "Dynamic Pix URL (required for dynamic)")
	flags.String("merchant-name", "", "Merchant name")
	flags.String("merchant-city", "", "Merchant city")
	flags.String("amount", "", "Transaction amount (optional)")
	flags.String("description", "", "Transaction description (optional)")
	flags.String("additional-info", "", "Additional info (static only)")
	flags.String("txid", "", "Transaction identifier (optional)")
	flags.Int("qr-size", 0, "PNG QR code size in pixels (default 256 when omitted)")
	flags.Int("ascii-scale", 1, "Scale factor (>=1) for ASCII QR output")
	flags.Bool("ascii-quiet", false, "Include quiet zone border in ASCII QR output")
	flags.String("ascii-black", "", "Character(s) used for dark modules in ASCII QR output")
	flags.String("ascii-white", "", "Character(s) used for light modules in ASCII QR output")

	return cmd
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run HTTP service that generates Pix payloads",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			addr := flags.Lookup("addr").Value.String()

			handler := http.NewServeMux()
			handler.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})
			handler.HandleFunc("/pix", pixHandler)

			server := &http.Server{
				Addr:         addr,
				Handler:      handler,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 15 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			log.Printf("pixgen HTTP server listening on %s", addr)
			return server.ListenAndServe()
		},
	}

	cmd.Flags().String("addr", ":8080", "HTTP listen address")

	return cmd
}

type pixRequest struct {
	Kind           string `json:"kind"`
	PixKey         string `json:"pixKey"`
	URL            string `json:"url"`
	MerchantName   string `json:"merchantName"`
	MerchantCity   string `json:"merchantCity"`
	Amount         string `json:"amount"`
	Description    string `json:"description"`
	AdditionalInfo string `json:"additionalInfo"`
	TxID           string `json:"txid"`
}

type pixResponse struct {
	Payload string `json:"payload"`
	QRCode  string `json:"qrCode"`
	Kind    string `json:"kind"`
	TxID    string `json:"txid"`
}

func pixHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req pixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}
	log.Printf("pix request received: kind=%s merchant=%s city=%s amount=%s txid=%s",
		req.Kind, req.MerchantName, req.MerchantCity, req.Amount, req.TxID)

	params, err := requestToParams(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload, qr, _, parsed, err := buildPix(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := pixResponse{
		Payload: payload,
		QRCode:  base64.StdEncoding.EncodeToString(qr),
		Kind:    parsed.Kind().String(),
		TxID:    parsed.AdditionalDataField.TxID,
	}
	log.Printf("pix response sent: kind=%s txid=%s payload_len=%d qr_len=%d",
		resp.Kind, resp.TxID, len(resp.Payload), len(resp.QRCode))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
	}
}

type pixParams struct {
	Kind           pix.PixKind
	PixKey         string
	URL            string
	MerchantName   string
	MerchantCity   string
	Amount         string
	Description    string
	AdditionalInfo string
	TxID           string
	QRCodeSize     int
	ASCII          asciiParams
}

type asciiParams struct {
	Scale     int
	Quiet     bool
	QuietSet  bool
	BlackChar string
	WhiteChar string
}

func collectPixParams(cmd *cobra.Command) (pixParams, error) {
	flags := cmd.Flags()
	params, err := parseParams(
		flags.Lookup("kind").Value.String(),
		flags.Lookup("key").Value.String(),
		flags.Lookup("url").Value.String(),
		flags.Lookup("merchant-name").Value.String(),
		flags.Lookup("merchant-city").Value.String(),
		flags.Lookup("amount").Value.String(),
		flags.Lookup("description").Value.String(),
		flags.Lookup("additional-info").Value.String(),
		flags.Lookup("txid").Value.String(),
	)
	if err != nil {
		return pixParams{}, err
	}

	if fl := flags.Lookup("qr-size"); fl != nil && fl.Value.String() != fl.DefValue {
		size, err := strconv.Atoi(fl.Value.String())
		if err != nil {
			return pixParams{}, fmt.Errorf("invalid qr-size: %w", err)
		}
		if size <= 0 {
			return pixParams{}, fmt.Errorf("qr-size must be greater than zero")
		}
		params.QRCodeSize = size
	}

	if fl := flags.Lookup("ascii-scale"); fl != nil && fl.Value.String() != fl.DefValue {
		scale, err := strconv.Atoi(fl.Value.String())
		if err != nil {
			return pixParams{}, fmt.Errorf("invalid ascii-scale: %w", err)
		}
		if scale < 1 {
			return pixParams{}, fmt.Errorf("ascii-scale must be at least 1")
		}
		params.ASCII.Scale = scale
	}

	if fl := flags.Lookup("ascii-quiet"); fl != nil && fl.Value.String() != fl.DefValue {
		quiet, err := strconv.ParseBool(fl.Value.String())
		if err != nil {
			return pixParams{}, fmt.Errorf("invalid ascii-quiet: %w", err)
		}
		params.ASCII.Quiet = quiet
		params.ASCII.QuietSet = true
	}

	if fl := flags.Lookup("ascii-black"); fl != nil && fl.Value.String() != fl.DefValue {
		params.ASCII.BlackChar = fl.Value.String()
	}

	if fl := flags.Lookup("ascii-white"); fl != nil && fl.Value.String() != fl.DefValue {
		params.ASCII.WhiteChar = fl.Value.String()
	}

	return params, nil
}

func requestToParams(req pixRequest) (pixParams, error) {
	return parseParams(
		req.Kind,
		req.PixKey,
		req.URL,
		req.MerchantName,
		req.MerchantCity,
		req.Amount,
		req.Description,
		req.AdditionalInfo,
		req.TxID,
	)
}

func parseParams(
	kindStr, key, url, merchantName, merchantCity, amount, description, additional, txid string,
) (pixParams, error) {
	kind, err := parseKind(kindStr)
	if err != nil {
		return pixParams{}, err
	}

	if merchantName == "" {
		return pixParams{}, errors.New("merchant-name is required")
	}
	if merchantCity == "" {
		return pixParams{}, errors.New("merchant-city is required")
	}

	if kind == pix.STATIC && key == "" {
		return pixParams{}, errors.New("pix key is required for static payloads")
	}
	if kind == pix.DYNAMIC && url == "" {
		return pixParams{}, errors.New("url is required for dynamic payloads")
	}

	return pixParams{
		Kind:           kind,
		PixKey:         key,
		URL:            url,
		MerchantName:   merchantName,
		MerchantCity:   merchantCity,
		Amount:         amount,
		Description:    description,
		AdditionalInfo: additional,
		TxID:           txid,
	}, nil
}

func parseKind(kind string) (pix.PixKind, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "static":
		return pix.STATIC, nil
	case "dynamic":
		return pix.DYNAMIC, nil
	default:
		return pix.STATIC, fmt.Errorf("invalid kind %q (expected static or dynamic)", kind)
	}
}

func buildPix(params pixParams) (string, []byte, string, *pix.ParsedPayload, error) {
	opts := []pix.Options{
		pix.OptKind(params.Kind),
		pix.OptMerchantName(params.MerchantName),
		pix.OptMerchantCity(params.MerchantCity),
	}

	if params.PixKey != "" {
		opts = append(opts, pix.OptPixKey(params.PixKey))
	}
	if params.Amount != "" {
		opts = append(opts, pix.OptAmount(params.Amount))
	}
	if params.Description != "" {
		opts = append(opts, pix.OptDescription(params.Description))
	}
	if params.AdditionalInfo != "" {
		opts = append(opts, pix.OptAdditionalInfo(params.AdditionalInfo))
	}
	if params.TxID != "" {
		opts = append(opts, pix.OptTxId(params.TxID))
	}
	if params.URL != "" {
		opts = append(opts, pix.OptUrl(params.URL))
	}
	if params.QRCodeSize > 0 {
		opts = append(opts, pix.OptQRCodeSize(params.QRCodeSize))
	}
	if params.ASCII.Scale > 0 {
		opts = append(opts, pix.OptQRCodeScale(params.ASCII.Scale))
	}
	if params.ASCII.BlackChar != "" || params.ASCII.WhiteChar != "" {
		opts = append(opts, pix.OptASCIICharset(params.ASCII.BlackChar, params.ASCII.WhiteChar))
	}
	if params.ASCII.QuietSet {
		opts = append(opts, pix.OptASCIIQuietZone(params.ASCII.Quiet))
	}

	p, err := pix.New(opts...)
	if err != nil {
		return "", nil, "", nil, err
	}

	payload, err := p.GenPayload()
	if err != nil {
		return "", nil, "", nil, err
	}
	qr, err := p.GenQRCode()
	if err != nil {
		return "", nil, "", nil, err
	}

	asciiQR, err := p.GenQRCodeASCII()
	if err != nil {
		return "", nil, "", nil, err
	}

	parsed, err := pix.ParsePayload(payload)
	if err != nil {
		return "", nil, "", nil, err
	}

	return payload, qr, asciiQR, parsed, nil
}
