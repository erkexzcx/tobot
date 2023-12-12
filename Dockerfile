FROM --platform=$BUILDPLATFORM golang:1.21-alpine as builder
RUN apk add --no-cache tesseract-ocr-dev gcc g++ git leptonica-dev ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG version
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} go build -a -ldflags "-w -s -X main.version=$version -extldflags '-static'" -o tobot ./cmd/tobot/main.go

FROM alpine
RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-lit ca-certificates
COPY --from=builder /app/tobot /tobot
ENTRYPOINT ["/tobot"]
