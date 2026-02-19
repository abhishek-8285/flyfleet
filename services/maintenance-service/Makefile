BINARY_NAME=maintenance-service

build:
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

migrate-up:
	# Add your migration tool command here (static placeholder)
	echo "Running migrations up..."

migrate-down:
	echo "Running migrations down..."

docker-build:
	docker build -t $(BINARY_NAME) .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/
	go clean
soul
