# --- build ---
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY shared/ ./shared/

COPY seeder-service/go.mod seeder-service/go.sum ./seeder-service/

WORKDIR /src/seeder-service

RUN go mod download

COPY seeder-service/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /bin/seeder \
      ./cmd/seeder

# --- runtime ---
FROM scratch
COPY --from=build /bin/seeder /usr/bin/seeder
ENTRYPOINT ["/usr/bin/seeder"]
