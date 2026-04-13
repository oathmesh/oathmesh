.PHONY: help dev test test-all race lint build demo clean docker-up docker-down pkl-gen test-node test-python audit-prep

help: ## Show this help
	@echo "OathMesh Makefile"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ── Development ──────────────────────────────────────────────────────────────

dev: ## Run issuer in dev mode (port 4000)
	OATHMESH_PRIVATE_KEY_FILE=./private.pem \
	OATHMESH_ISSUER=http://localhost:4000 \
	go run ./cmd/oathmesh serve --port 4000

build: ## Build the oathmesh CLI binary
	CGO_ENABLED=0 go build -o bin/oathmesh ./cmd/oathmesh

# ── Testing ──────────────────────────────────────────────────────────────────

test: ## Run Go tests
	go test -v ./...

race: ## Run Go tests with race detector
	go test -race ./...

test-node: ## Run Node SDK tests (vitest)
	cd sdk/node && npx vitest run

test-python: ## Run Python SDK tests (pytest)
	cd sdk/python && python -m pytest tests/ -v

test-all: race test-node test-python ## Run ALL tests (Go + Node + Python)

# ── Quality ──────────────────────────────────────────────────────────────────

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed — run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# ── Docker ───────────────────────────────────────────────────────────────────

docker-up: ## Start all services via docker compose
	docker compose up -d --build

docker-down: ## Stop all services
	docker compose down

# ── Demo ─────────────────────────────────────────────────────────────────────

demo: ## Run the golden-path end-to-end demo
	bash demo.sh

# ── Conformance ───────────────────────────────────────────────────────────────

conformance: ## Run cross-SDK conformance tests (requires services running)
	@echo "Running cross-SDK conformance suite..."
	@echo "Ensure issuer, chi-api, express-api, and fastapi-api are running"
	@echo "via: docker compose up -d issuer chi-api express-api fastapi-api"
	@bash conformance/run_all.sh

# ── Benchmarks ─────────────────────────────────────────────────────────────────

bench: ## Run Go benchmarks and output to bench.txt
	@echo "Running Go benchmarks..."
	@echo "# OathMesh Benchmarks - $$(date)" > bench.txt
	@echo "" >> bench.txt
	go test -bench=. -benchmem -run=^$$ ./internal/verify/... ./internal/sign/... 2>&1 | tee -a bench.txt
	@echo ""
	@echo "Results saved to bench.txt"
	@echo "Update ARCHITECTURE.md with real measured numbers using: make update-arch"

# ── Codegen ──────────────────────────────────────────────────────────────────

pkl-gen: ## Regenerate Go code from Pkl policy schema
	pkl-gen-go --schema policy/schema.pkl --output internal/policy/generated.go

# ── Audit ───────────────────────────────────────────────────────────────────────

audit-prep: ## Run pre-audit checks (tests, lint, vulnerability scan)
	@echo "Running pre-audit checks for OathMesh..."
	@echo ""
	@echo "1. Running Go tests..."
	@go test ./... || { echo "TESTS FAILED"; exit 1; }
	@echo ""
	@echo "2. Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "Warning: golangci-lint not installed"
	@echo ""
	@echo "3. Checking for vulnerabilities..."
	@command -v govulncheck >/dev/null 2>&1 && govulncheck ./... || echo "Warning: govulncheck not installed - install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
	@echo ""
	@echo "Pre-audit checks complete."

# ── Cleanup ──────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf bin/ dist/
	rm -rf sdk/node/dist/
