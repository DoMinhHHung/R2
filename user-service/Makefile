.PHONY: run build test migrate-up migrate-down swagger

-include .env

APP_NAME=user-service
MIGRATION_DIR=./migrations

run:
	go run ./cmd/server/main.go

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" \
		-o bin/$(APP_NAME) ./cmd/server/main.go

migrate-up:
	migrate -path $(MIGRATION_DIR) -database "$(DB_DSN)" up

migrate-down:
	migrate -path $(MIGRATION_DIR) -database "$(DB_DSN)" down

swagger:
	swag init -g cmd/server/main.go -o docs --parseDependency

test:
	go test ./... -v -race -coverprofile=coverage.out

test-coverage:
	go tool cover -html=coverage.out

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest