.PHONY: help build run stop clean migrate test lint

help:
	@echo "Available commands:"
	@echo "  make build    - Build all containers"
	@echo "  make run      - Start all services"
	@echo "  make stop     - Stop all services"
	@echo "  make clean    - Clean up containers and volumes"
	@echo "  make migrate  - Run database migrations"
	@echo "  make test     - Run tests"
	@echo "  make lint     - Run linters"

build:
	docker-compose build

run:
	docker-compose up -d

stop:
	docker-compose down

clean:
	docker-compose down -v

migrate:
	docker-compose exec backend go run ./cmd/orca migrate up

test:
	cd backend && go test ./...
	cd frontend && npm test

lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

logs:
	docker-compose logs -f

backend-logs:
	docker-compose logs -f backend

frontend-logs:
	docker-compose logs -f frontend

db-logs:
	docker-compose logs -f postgres