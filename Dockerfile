FROM golang:1.22-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/pixgen ./cmd/pixgen

FROM gcr.io/distroless/base-debian12

COPY --from=builder /out/pixgen /usr/local/bin/pixgen

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/pixgen"]
CMD ["serve", "--addr", ":8080"]
