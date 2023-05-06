## Build stage
FROM golang:1.20-alpine AS build-env
RUN apk add --no-cache \
    tesseract-ocr-dev \
    gcc \
    g++ \
    git \
    leptonica-dev \
    ca-certificates
ADD . /build
WORKDIR /build
RUN go build -a -ldflags '-s -w' -o tobot ./cmd/tobot/main.go

# TODO - anti cheat always fail, no idea why

## Create image
FROM alpine:3.17
RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-lit \
    ca-certificates
COPY --from=build-env /build/tobot /tobot
WORKDIR /
ENTRYPOINT ["/tobot"]
