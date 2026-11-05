module users-service

go 1.27

require (
	github.com/gorilla/mux v1.8.1
	github.com/jmoiron/sqlx v1.4.0
	shared v1.0.0
)

require (
	github.com/lib/pq v1.12.3 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pressly/goose/v3 v3.29.0 // indirect
	github.com/sethvargo/go-retry v0.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
)

replace shared => ../shared
