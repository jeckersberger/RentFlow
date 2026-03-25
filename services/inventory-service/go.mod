module github.com/jeckersberger/rentflow/services/inventory-service

go 1.24.0

require (
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
	github.com/lib/pq v1.12.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

require (
	github.com/EventStore/EventStore-Client-Go/v4 v4.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.3 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
