module github.com/jeckersberger/rentflow/services/auth-service

go 1.22

require (
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
	github.com/lib/pq v1.10.9
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
