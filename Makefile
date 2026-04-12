.PHONY: help pkl-gen build test vet lint clean

help:
	@echo "OathMesh Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  pkl-gen   - Generate Go code from Pkl policy schema"
	@echo "  build    - Build the oathmesh binary"
	@echo "  test     - Run all tests"
	@echo "  vet      - Run go vet"
	@echo "  lint     - Run golangci-lint (if installed)"
	@echo "  clean    - Clean build artifacts"

pkl-gen:
	@echo "Generating Go code from Pkl schema..."
	pkl-gen-go --schema policy/schema.pkl --output internal/policy/generated.go

build:
	go build -o bin/oathmesh ./cmd/oathmesh

test:
	go test -v ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed"

clean:
	rm -rf bin/
