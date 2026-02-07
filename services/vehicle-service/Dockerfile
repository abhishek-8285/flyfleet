FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vehicle-service ./cmd/

# Use distroless for minimal image
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary
COPY --from=builder /app/vehicle-service .

# Expose port
EXPOSE 8081

# Run the binary
USER nonroot:nonroot
ENTRYPOINT ["./vehicle-service"]