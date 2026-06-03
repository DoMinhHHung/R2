.PHONY: swagger run build

swagger:
	swag init -g cmd/main.go -o docs --parseDependency

run:
	go run ./cmd/main.go

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/gateway ./cmd/main.go