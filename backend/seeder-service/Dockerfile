# --- build ---
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/seeder ./cmd/seeder

# --- runtime ---
FROM scratch
COPY --from=build /bin/seeder /usr/bin/seeder
ENTRYPOINT ["/usr/bin/seeder"]
