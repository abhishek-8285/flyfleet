FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=dev" -o /bin/mlplatform ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/mlplatform /bin/mlplatform
COPY config.yaml /app/config.yaml

ENTRYPOINT ["mlplatform"]
