FROM golang:1.26-alpine AS builder

ARG SERVICE_NAME
WORKDIR /app

# Copy only mod/work files first to cache dependencies for the entire workspace
COPY go.work go.work.sum ./
COPY packages/common/go/go.mod packages/common/go/go.sum ./packages/common/go/
COPY proto/go.mod proto/go.sum ./proto/
# Copy ALL service mod/sum files (even if some sums are missing, use wildcard to be safe)
COPY services/*/go.mod ./services/
# Need to copy individually if wildcard doesn't preserve structure well in some docker versions
COPY services/microservices/admin-support-service/go.mod services/microservices/admin-support-service/go.sum ./services/microservices/admin-support-service/
COPY services/microservices/alert-service/go.mod services/microservices/alert-service/go.sum ./services/microservices/alert-service/
COPY services/microservices/analytics-service/go.mod services/microservices/analytics-service/go.sum ./services/microservices/analytics-service/
COPY services/microservices/auth-service/go.mod services/microservices/auth-service/go.sum ./services/microservices/auth-service/
COPY services/microservices/billing-service/go.mod services/microservices/billing-service/go.sum ./services/microservices/billing-service/
COPY services/microservices/driver-service/go.mod services/microservices/driver-service/go.sum ./services/microservices/driver-service/
COPY services/fleet-core-service-legacy/go.mod services/fleet-core-service-legacy/go.sum ./services/fleet-core-service-legacy/
COPY services/microservices/fleet-core-service-new/go.mod services/microservices/fleet-core-service-new/go.sum ./services/microservices/fleet-core-service-new/
COPY services/microservices/fleet-ml-service/go.mod ./services/microservices/fleet-ml-service/
COPY services/microservices/governance-service/go.mod services/microservices/governance-service/go.sum ./services/microservices/governance-service/
COPY services/microservices/india-integration-service/go.mod services/microservices/india-integration-service/go.sum ./services/microservices/india-integration-service/
COPY services/microservices/ingestion-service/go.mod services/microservices/ingestion-service/go.sum ./services/microservices/ingestion-service/
COPY services/microservices/integrations-service/go.mod services/microservices/integrations-service/go.sum ./services/microservices/integrations-service/
COPY services/microservices/maintenance-service/go.mod services/microservices/maintenance-service/go.sum ./services/microservices/maintenance-service/
COPY services/microservices/optimization-service/go.mod ./services/microservices/optimization-service/
COPY services/microservices/reports-service/go.mod services/microservices/reports-service/go.sum ./services/microservices/reports-service/
COPY services/microservices/settings-service/go.mod services/microservices/settings-service/go.sum ./services/microservices/settings-service/
COPY services/microservices/tracking-service/go.mod services/microservices/tracking-service/go.sum ./services/microservices/tracking-service/
COPY services/microservices/trip-service/go.mod services/microservices/trip-service/go.sum ./services/microservices/trip-service/
COPY services/microservices/vehicle-service/go.mod services/microservices/vehicle-service/go.sum ./services/microservices/vehicle-service/
COPY services/microservices/web-bff/go.mod services/microservices/web-bff/go.sum ./services/microservices/web-bff/
COPY services/microservices/worker-service/go.mod services/microservices/worker-service/go.sum ./services/microservices/worker-service/

# Pre-download dependencies (cached until ANY mod file changes)
RUN go mod download

# Copy packages and services for the workspace
COPY packages ./packages
COPY services ./services
COPY proto ./proto

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin ./services/${SERVICE_NAME}/cmd/

# Create uploads directory and set permissions for nonroot (UID 65532)
RUN mkdir -p /app/uploads && chown 65532:65532 /app/uploads

# Final stage
FROM gcr.io/distroless/static:nonroot AS distroless

# We need a stage with a shell and pg_dump for DR service
FROM alpine:3.21 AS dr-base
RUN apk add --no-cache postgresql16-client bash

FROM dr-base
WORKDIR /
COPY --from=builder /app/bin .
COPY --from=builder --chown=65532:65532 /app/uploads ./uploads

# Expose port (8080 is standard for internal services, but compose handles mapping)
EXPOSE 8080

# Run the binary
USER 65532:65532
ENTRYPOINT ["./bin"]
