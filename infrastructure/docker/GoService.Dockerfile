FROM golang:1.25-alpine AS builder

ARG SERVICE_NAME
WORKDIR /app

# Copy go.work and shared mod files
COPY go.work go.work.sum ./

# Copy packages and services for the workspace
COPY packages ./packages
COPY services ./services

# Pre-download dependencies
RUN go mod download

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
