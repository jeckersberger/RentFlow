# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build

# Copy workspace and common package
COPY go.work go.work
COPY pkg/common/ pkg/common/
COPY services/workflow-service/ services/workflow-service/

WORKDIR /build/services/workflow-service
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -ldflags="-s -w" -o /app ./cmd/server

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app .
COPY --from=builder /build/services/workflow-service/migrations ./migrations/

EXPOSE 8013
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -qO- http://localhost:8013/health || exit 1

ENTRYPOINT ["/app/app"]
