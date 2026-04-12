.PHONY: help dev test race lint build demo clean docker-up pkl-gen

help: ## Show this help
	@echo "OathMesh Makefile"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Run issuer in dev mode
	OATHMESH_PRIVATE_KEY_FILE=./private.pem \
	OATHMESH_ISSUER=http://localhost:4000 \
	go run ./cmd/oathmesh serve --port 4000

test: ## Run all tests
	go test -v ./...

race: ## Run all tests with race detector
	go test -race ./...

lint: ## Run golangci-lint (if installed)
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed — run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

build: ## Build the oathmesh binary
	CGO_ENABLED=0 go build -o bin/oathmesh ./cmd/oathmesh

demo: ## Run the golden path end-to-end demo
	bash demo.sh

clean: ## Remove build artifacts
	rm -rf bin/ dist/

docker-up: ## Start all services via docker-compose
	docker-compose up -d

pkl-gen: ## Regenerate Go code from Pkl policy schema
	pkl-gen-go --schema policy/schema.pkl --output internal/policy/generated.go
