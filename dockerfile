FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o tkpst_parser

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/tkpst_parser /app/tkpst_parser

WORKDIR /app

ENTRYPOINT ["/app/tkpst_parser"]
