module github.com/jeckersberger/rentflow/services/insurance-service

go 1.24

require (
	github.com/google/uuid v1.6.0
	github.com/jeckersberger/rentflow/pkg/common v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.12.0
	github.com/rs/zerolog v1.34.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
