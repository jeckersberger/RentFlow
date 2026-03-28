FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy Go workspace files first (for layer caching).
COPY go.work ./
COPY go.work.sum* ./

# Copy common package (shared library).
COPY pkg/common/ ./pkg/common/

# Copy all peer service go.mod files referenced by go.work.
COPY services/auth/go.mod ./services/auth/go.mod
COPY services/project/go.mod ./services/project/go.mod
COPY services/scanner/go.mod ./services/scanner/go.mod
COPY services/warehouse/go.mod ./services/warehouse/go.mod
COPY services/invoice/go.mod ./services/invoice/go.mod
COPY services/document/go.mod ./services/document/go.mod
COPY services/crew/go.mod ./services/crew/go.mod
COPY services/federation/go.mod ./services/federation/go.mod
COPY services/maintenance/go.mod ./services/maintenance/go.mod
COPY services/transport/go.mod ./services/transport/go.mod
COPY services/insurance/go.mod ./services/insurance/go.mod
COPY services/workflow/go.mod ./services/workflow/go.mod
COPY services/ai/go.mod ./services/ai/go.mod
COPY services/notification/go.mod ./services/notification/go.mod
COPY services/reporting/go.mod ./services/reporting/go.mod
COPY services/audit/go.mod ./services/audit/go.mod
COPY services/expense/go.mod ./services/expense/go.mod
COPY services/customer/go.mod ./services/customer/go.mod

# Copy inventory service source.
COPY services/inventory/ ./services/inventory/

# Download dependencies and build.
WORKDIR /build/services/inventory
RUN go mod download
RUN go build -o /app/app ./cmd/server

# --- Runtime stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/app .
COPY services/inventory/migrations/ ./migrations/

EXPOSE 8002

CMD ["/app/app"]
