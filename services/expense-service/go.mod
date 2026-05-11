module github.com/jeckersberger/rentflow/services/expense-service

go 1.22

require (
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
	github.com/lib/pq v1.12.0
)

require (
	github.com/EventStore/EventStore-Client-Go/v4 v4.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
