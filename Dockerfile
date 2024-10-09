ARG ALPINETAG="3.20"
ARG GOVERSION="1.23"

FROM golang:${GOVERSION}-alpine${ALPINETAG} AS builder
RUN apk add --no-cache tesseract-ocr-dev leptonica-dev gcc g++
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG version
RUN go build -a -ldflags "-w -s -X main.version=$version" -o tobot ./cmd/tobot/main.go

FROM alpine:${ALPINETAG}
RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-lit leptonica ca-certificates
COPY --from=builder /app/tobot /tobot
ENTRYPOINT ["/tobot"]
