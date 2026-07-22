FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


ARG BUILD_TIME
ARG VERSION

RUN go build \
    -ldflags="-s -w \
    -X main.buildTime=${BUILD_TIME} \
    -X main.version=${VERSION}" \
    -o /greenlight-api \
    ./cmd/api



FROM alpine:latest

WORKDIR /app

COPY --from=builder /greenlight-api .

EXPOSE 4000

ENTRYPOINT ["./greenlight-api"]