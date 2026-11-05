FROM golang:1.27-alpine AS build

WORKDIR /src

COPY shared/ ./shared/

COPY analytics-service/go.mod analytics-service/go.sum ./analytics-service/

WORKDIR /src/analytics-service

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY analytics-service/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/analytics-service \
      ./cmd/app


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build --chown=app:app /out/analytics-service /app/analytics-service
COPY --chown=app:app analytics-service/.config/prod.json /app/.config/prod.json

USER app

ENV ENV=prod
ENV SYNC_INTERVAL=${SYNC_INTERVAL:-10m}
ENV MOCK_DELAY=${MOCK_DELAY:-0}
ENV MOCK_DELAY_JITTER=${MOCK_DELAY_JITTER:-0}
ENV MOCK_ERROR_RATE_500=${MOCK_ERROR_RATE_500:-0}
ENV MOCK_ERROR_RATE_429=${MOCK_ERROR_RATE_429:-0}

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
    CMD wget -qO- http://127.0.0.1:80/healthz >/dev/null || exit 1

ENTRYPOINT ["./analytics-service"]
