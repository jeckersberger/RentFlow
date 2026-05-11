module github.com/jeckersberger/rentflow/services/audit-service

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
	github.com/lib/pq v1.12.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
