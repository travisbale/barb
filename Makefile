VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

.PHONY: build release frontend clean test unit fmt lint dev

frontend:
	@echo "Building frontend..."
	@cd frontend && npm ci && npm run build
	@echo "  frontend/dist/"

build: frontend
	@mkdir -p build
	@go build $(LDFLAGS) -o build/barb ./cmd/barb
	@echo "  build/barb"

release: frontend
	@mkdir -p build
	@GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o build/barb-linux-amd64       ./cmd/barb
	@GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o build/barb-linux-arm64       ./cmd/barb
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o build/barb-windows-amd64.exe ./cmd/barb

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf build/
	@rm -rf cmd/barb/dist/

test:
	@echo "Running all tests..."
	@go test -count=1 -timeout=120s ./...

unit:
	@echo "Running unit tests..."
	@go test ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@go run golang.org/x/tools/cmd/goimports@v0.38.0 -w .

lint:
	@echo "Linting code..."
	@docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.11 golangci-lint run

dev-backend:
	@go run ./cmd/barb --debug

dev-frontend:
	@cd frontend && npm install --silent && npm run dev
