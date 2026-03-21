module github.com/jeckersberger/rentflow/services/inventory-service

go 1.22

require (
	github.com/jeckersberger/rentflow/pkg/common v0.0.0
	github.com/lib/pq v1.12.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

replace github.com/jeckersberger/rentflow/pkg/common => ../../pkg/common
