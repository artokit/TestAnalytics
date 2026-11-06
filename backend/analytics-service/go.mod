module analytics-service

go 1.27

require (
	github.com/gorilla/mux v1.8.1
	github.com/redis/go-redis/v9 v9.23.0
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.6
	golang.org/x/sync v0.24.0
	shared v1.0.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace shared => ../shared
