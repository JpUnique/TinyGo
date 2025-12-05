APP_NAME=TinyGo
DOCKER_IMAGE=$(APP_NAME):latest

.PHONY: run build docker-up docker-down migrate-up migrate-down

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down -v

migrate-up:
	docker-compose run --rm migrate up

migrate-down:
	docker-compose run --rm migrate down

fmt:
	go fmt ./...
	go run golang.org/x/tools/cmd/goimports@v0.29.0 -w .

build:
	go build

test:
	go test -v ./... -short

start: build
	./TinyGo --log-level info

clean:
	go clean
	go clean -cache

tidy:
	go mod tidy

run:
	go run main.go