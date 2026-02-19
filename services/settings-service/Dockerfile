FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o settings-service ./cmd

FROM alpine:3.19

WORKDIR /root/

COPY --from=builder /app/settings-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8085

CMD ["./settings-service"]
