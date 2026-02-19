FROM golang:1.25-alpine AS builder

ARG SERVICE_NAME
WORKDIR /app

# Copy only mod/work files first to cache dependencies for the entire workspace
COPY go.work go.work.sum ./
COPY packages/common/go/go.mod packages/common/go/go.sum ./packages/common/go/
# Copy all service mod files to satisfy go.work requirements
COPY services/admin-support-service/go.mod services/admin-support-service/go.sum ./services/admin-support-service/
COPY services/alert-service/go.mod services/alert-service/go.sum ./services/alert-service/
COPY services/analytics-service/go.mod services/analytics-service/go.sum ./services/analytics-service/
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/
COPY services/billing-service/go.mod services/billing-service/go.sum ./services/billing-service/
COPY services/driver-service/go.mod services/driver-service/go.sum ./services/driver-service/
COPY services/fleet-core-service/go.mod services/fleet-core-service/go.sum ./services/fleet-core-service/
COPY services/governance-service/go.mod services/governance-service/go.sum ./services/governance-service/
COPY services/india-integration-service/go.mod services/india-integration-service/go.sum ./services/india-integration-service/
COPY services/ingestion-service/go.mod services/ingestion-service/go.sum ./services/ingestion-service/
COPY services/integrations-service/go.mod services/integrations-service/go.sum ./services/integrations-service/
COPY services/maintenance-service/go.mod services/maintenance-service/go.sum ./services/maintenance-service/
COPY services/reports-service/go.mod services/reports-service/go.sum ./services/reports-service/
COPY services/settings-service/go.mod services/settings-service/go.sum ./services/settings-service/
COPY services/tracking-service/go.mod services/tracking-service/go.sum ./services/tracking-service/
COPY services/trip-service/go.mod services/trip-service/go.sum ./services/trip-service/
COPY services/vehicle-service/go.mod services/vehicle-service/go.sum ./services/vehicle-service/
COPY services/web-bff/go.mod services/web-bff/go.sum ./services/web-bff/
COPY services/worker-service/go.mod services/worker-service/go.sum ./services/worker-service/

# Pre-download dependencies (cached until ANY mod file changes)
RUN go mod download

# Copy packages and services for the workspace
COPY packages ./packages
COPY services ./services

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin ./services/${SERVICE_NAME}/cmd/

# Create uploads directory and set permissions for nonroot (UID 65532)
RUN mkdir -p /app/uploads && chown 65532:65532 /app/uploads

# Final stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/bin .
COPY --from=builder --chown=nonroot:nonroot /app/uploads ./uploads

# Expose port (8080 is standard for internal services, but compose handles mapping)
EXPOSE 8080

# Run the binary
USER nonroot:nonroot
ENTRYPOINT ["./bin"]
