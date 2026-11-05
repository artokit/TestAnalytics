
FROM golang:1.27-alpine AS build

WORKDIR /src

# shared подключается через replace ../shared, поэтому кладём его рядом
COPY shared/ ./shared/

COPY users-service/go.mod users-service/go.sum ./users-service/

WORKDIR /src/users-service

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY users-service/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/users-service \
      ./cmd/app


FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build --chown=app:app /out/users-service /app/users-service
COPY --chown=app:app users-service/.config/prod.json /app/.config/prod.json
COPY --chown=app:app users-service/migrations /app/migrations

USER app

ENV CONFIG_FILE=/app/.config/prod.json

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:80/healthz >/dev/null || exit 1

ENTRYPOINT ["./users-service"]
