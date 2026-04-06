VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
DEV_PORT := 4443

.PHONY: build release frontend frontend-test clean test unit fmt lint dev

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

frontend-test:
	@echo "Running frontend tests..."
	@cd frontend && npm ci --silent && npm run test

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf build/
	@rm -rf cmd/barb/dist/

test: frontend-test
	@echo "Running all tests (including integration)..."
	@go test -race -count=1 -timeout=120s ./...

unit:
	@echo "Running unit tests only..."
	@go test -race -tags unit -count=1 -timeout=60s ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@go run golang.org/x/tools/cmd/goimports@v0.38.0 -w .

lint:
	@echo "Linting code..."
	@docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.11 golangci-lint run
	@echo "Type checking frontend..."
	@cd frontend && npx vue-tsc --noEmit

dev: frontend
	@docker run -d --name mailpit --rm -p 1025:1025 -p 8025:8025 \
		-e MP_SMTP_AUTH_ACCEPT_ANY=1 -e MP_SMTP_AUTH_ALLOW_INSECURE=1 \
		axllent/mailpit:latest
	@bash -c 'trap "docker stop mailpit; kill 0" EXIT; \
		(cd frontend && VITE_API_PORT=$(DEV_PORT) npm run dev) & \
		go run ./cmd/barb --addr :$(DEV_PORT) --debug'
