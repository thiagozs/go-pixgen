module github.com/thiagozs/go-pixgen

go 1.17

require (
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/snksoft/crc v1.1.0
	github.com/spf13/cobra v1.7.0
)

replace github.com/spf13/cobra => ./internal/cobra
