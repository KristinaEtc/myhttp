BINARY_NAME="myhttp"

all : lint test build

build:
	go build -o $(BINARY_NAME) cmd/main.go

run: ## Run service locally
	go run -race cmd/main.go

lint: ## Run linter
	@echo 'running linter...'
	@golangci-lint run ./...

test: ## Run tests
	@echo 'running unit-tests...'
	@go test -race ./...


