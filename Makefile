.PHONY: run build migrate-up migrate-down docker-build docker-run compose-up compose-down

build:
	@mkdir -p ./bin
	@go build -o ./bin/main ./cmd/main
	
run: build
	@./bin/main

migrate-up:
	@go run ./cmd/migration up

migrate-down:
	@go run ./cmd/migration down

docker-build:
	@mkdir -p ./bin
	@CGO_ENABLED=0 GOOS=linux go build -o ./bin/main ./cmd/main

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down