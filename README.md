# shop-proxy

`shop-proxy` — отдельный proxy-service для задания со звездочкой.

Сервис встраивается между `shop-gateway` и `shop-catalog-service`, прозрачно проксирует gRPC/TCP трафик, считает байты в обе стороны и позволяет управлять искусственной задержкой через admin HTTP API.

## Что умеет сервис

1. Принимать TCP/gRPC трафик на `:9095`.
2. Пересылать его в upstream `catalog-service:9091`.
3. Считать:
   - количество соединений;
   - количество активных соединений;
   - байты `client -> upstream`;
   - байты `upstream -> client`;
   - ошибки copy loop.
4. Отдавать admin HTTP endpoints на `:8085`.
5. Защищать `/admin/*` через middleware с `Authorization`, `X-App-Id`, `ValidateToken` и `IsAdmin`.
6. Позволять включать fault injection через `delay`.

## Архитектура

```text
shop-gateway
  -> gRPC/TCP -> shop-proxy:9095
shop-proxy:9095
  -> gRPC/TCP -> shop-catalog-service:9091

operator / Prometheus
  -> HTTP -> shop-proxy:8085
```

## Основные endpoints

Public:

1. `GET /health`
2. `GET /ready`
3. `GET /metrics`

Admin-only:

1. `GET /admin/state`
2. `POST /admin/faults/delay?ms=500`
3. `POST /admin/reset`

## Конфигурация

Файлы:

1. `config/config.local.yaml`
2. `config/config.docker.yaml`

Ключевые поля:

1. `http.addr` — admin HTTP server, по умолчанию `:8085`
2. `proxy.listen_addr` — входящий TCP proxy порт, по умолчанию `:9095`
3. `proxy.upstream_addr` — upstream catalog gRPC адрес
4. `auth_grpc.addr` — gRPC адрес `auth-service`
5. `auth_tls.*` — mTLS настройки для admin auth middleware
6. `faults.delay` — стартовая задержка

## Локальный запуск

Из директории `shop-proxy`:

```bash
go run ./cmd/proxy-service --config config/config.local.yaml
```

## Docker запуск

```bash
docker build -t shop-proxy:local .
docker run --rm -p 8085:8085 -p 9095:9095 --name proxy-service shop-proxy:local
```

## Проверки

Health:

```bash
curl http://localhost:8085/health
curl http://localhost:8085/ready
curl http://localhost:8085/metrics
```

Admin state:

```bash
curl -H "Authorization: Bearer <access_token>" -H "X-App-Id: 1" http://localhost:8085/admin/state
```

Установить delay:

```bash
curl -X POST -H "Authorization: Bearer <access_token>" -H "X-App-Id: 1" "http://localhost:8085/admin/faults/delay?ms=500"
```

Сбросить delay:

```bash
curl -X POST -H "Authorization: Bearer <access_token>" -H "X-App-Id: 1" http://localhost:8085/admin/reset
```

## Что осталось улучшить

1. Добавить больше видов fault injection, например bandwidth limit.
2. Добавить больше operational сценариев, например jitter или bandwidth limit.
3. При желании добавить отдельный dashboard в Grafana под `proxy_*` метрики.
