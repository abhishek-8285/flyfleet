FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dr-service ./cmd/

# Use distroless for minimal image
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary
COPY --from=builder /app/dr-service .

# Expose port
EXPOSE 8084

# Run the binary
USER nonroot:nonroot
ENTRYPOINT ["./dr-service"]
