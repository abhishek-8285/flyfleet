BINARY_NAME=billing-service

build:
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

