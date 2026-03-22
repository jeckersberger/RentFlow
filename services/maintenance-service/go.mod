module github.com/jeckersberger/rentflow/services/maintenance-service

go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
)

require (
	github.com/lib/pq v1.12.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
