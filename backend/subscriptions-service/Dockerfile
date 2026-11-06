
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY shared/ ./shared/

COPY subscriptions-service/go.mod subscriptions-service/go.sum ./subscriptions-service/

WORKDIR /src/subscriptions-service

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY subscriptions-service/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/subscriptions-service \
      ./cmd/app


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build --chown=app:app /out/subscriptions-service /app/subscriptions-service
COPY --chown=app:app subscriptions-service/migrations /app/migrations

USER app

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:80/healthz >/dev/null || exit 1

ENTRYPOINT ["./subscriptions-service"]
