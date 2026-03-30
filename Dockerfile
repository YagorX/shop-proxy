FROM golang:1.25.0 AS builder
WORKDIR /src/shop-proxy

COPY shop-proxy/go.mod shop-proxy/go.sum ./
COPY shop-contracts /src/shop-contracts
RUN go mod download

COPY shop-proxy ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/proxy-service ./cmd/proxy-service

FROM alpine:3.20
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/proxy-service /app/proxy-service
COPY shop-proxy/config/config.docker.yaml /app/config/config.docker.yaml
COPY shop-proxy/certs ./certs

USER app

EXPOSE 8085 9095

ENV CONFIG_PATH=/app/config/config.docker.yaml

ENTRYPOINT ["/app/proxy-service"]
