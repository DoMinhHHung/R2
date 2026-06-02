.PHONY: run build test migrate-up migrate-down generate-keys swagger

-include .env

APP_NAME=auth-service
MIGRATION_DIR=./migrations

run:
	go run ./cmd/server/main.go

# Build
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/server/main.go

migrate-up:
	migrate -path $(MIGRATION_DIR) -database "$(DB_DSN)" up

migrate-down:
	migrate -path $(MIGRATION_DIR) -database "$(DB_DSN)" down

migrate-to:
	migrate -path $(MIGRATION_DIR) -database "$(DB_DSN)" goto $(V)

generate-keys:
	@powershell -NoProfile -Command "New-Item -ItemType Directory -Force keys | Out-Null"
	go run ./scripts/gen_keys.go
	@echo "✅ RS256 key pair generated in ./keys/"

swagger:
	swag init -g cmd/server/main.go -o docs

test:
	go test ./... -v -race -coverprofile=coverage.out

test-coverage:
	go tool cover -html=coverage.out

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest