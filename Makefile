.PHONY: run build migrate-up migrate-down

build:
	@go build -o ./bin/main ./cmd/main
	
run: build
	@./bin/main

migrate-up:
	@go run ./cmd/migration up

migrate-down:
	@go run ./cmd/migration down