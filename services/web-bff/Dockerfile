FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
# Note: go.sum will be created after go mod tidy or download if dependencies are added
# COPY go.sum ./ 

# Download dependencies
# RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o web-bff ./cmd/main.go

# Use distroless for minimal image
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary
COPY --from=builder /app/web-bff .

# Expose port
EXPOSE 8081

# Run the binary
USER nonroot:nonroot
ENTRYPOINT ["./web-bff"]
