build:
	@echo "Building..."
	@go build -o main.exe cmd/main.go

run:
	@go run cmd/main.go

docker-run:
	@docker compose up --build

docker-down:
	@docker compose down

watch:
	@powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
		air; \
	} else { \
		go install github.com/air-verse/air@latest; \
		air; \
	}"

.PHONY: build run docker-run docker-down watch
