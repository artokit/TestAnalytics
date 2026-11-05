# --- build ---
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY shared/ ./shared/

COPY seeder-service/go.mod seeder-service/go.sum ./seeder-service/

WORKDIR /src/seeder-service

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY seeder-service/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /bin/seeder \
      ./cmd/seeder

# --- runtime ---
FROM scratch
COPY --from=build /bin/seeder /usr/bin/seeder
ENTRYPOINT ["/usr/bin/seeder"]
