module github.com/jeckersberger/rentflow/services/insurance-service

go 1.22

require github.com/jeckersberger/rentflow/pkg/common v0.0.0

require github.com/lib/pq v1.12.0 // indirect

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
