
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY shared/ ./shared/

COPY payments-service/go.mod payments-service/go.sum ./payments-service/

WORKDIR /src/payments-service

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY payments-service/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/payments-service \
      ./cmd/app


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build --chown=app:app /out/payments-service /app/payments-service
COPY --chown=app:app payments-service/.config/prod.json /app/.config/prod.json
COPY --chown=app:app payments-service/migrations /app/migrations

USER app

ENV ENV=prod

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:80/healthz >/dev/null || exit 1

ENTRYPOINT ["./payments-service"]
