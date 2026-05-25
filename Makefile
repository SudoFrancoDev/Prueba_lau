.PHONY: run build test test-unit migrate import-csv docker-up docker-down lint

## Build the binary
build:
	go build -o bin/bia-energy ./cmd/api

## Run locally (requires XAMPP MySQL running on port 3306)
run:
	go run ./cmd/api

## Run all unit tests (no DB required)
test:
	go test ./internal/... -v -count=1

## Run only service + handler unit tests
test-unit:
	go test ./internal/service/... ./internal/handler/... -v -count=1

## Apply DB migrations via XAMPP MySQL
migrate:
	mysql -h $${DB_HOST:-127.0.0.1} -P $${DB_PORT:-3306} \
	      -u $${DB_USER:-root} \
	      < migrations/001_create_consumptions.sql

## Import CSV (usage: make import-csv  or  make import-csv CSV=path/to/file.csv)
import-csv:
	./scripts/import_csv.sh $${CSV:-data/consumptions.csv}

## Start Docker services (MySQL + API)
docker-up:
	docker-compose up -d --build

## Stop Docker services
docker-down:
	docker-compose down

## Run linter
lint:
	golangci-lint run ./...
