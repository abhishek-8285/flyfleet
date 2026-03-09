FROM golang:1.26-alpine AS builder

ARG SERVICE_NAME
WORKDIR /app

# Copy only mod/work files first to cache dependencies for the entire workspace
COPY go.work go.work.sum ./
COPY packages/common/go/go.mod packages/common/go/go.sum ./packages/common/go/
COPY proto/go.mod proto/go.sum ./proto/
# Copy service sources and resolve dependencies for the selected service only.
COPY services ./services

# Pre-download dependencies for target service.
RUN cd ./services/${SERVICE_NAME} && go mod download

# Copy packages and proto for local replace modules.
COPY packages ./packages
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
