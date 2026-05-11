module github.com/jeckersberger/rentflow/services/auth-service

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
	github.com/lib/pq v1.12.0
	golang.org/x/crypto v0.50.0
)

require (
	github.com/EventStore/EventStore-Client-Go/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/redis/go-redis/v9 v9.0.0 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/grpc v1.61.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
