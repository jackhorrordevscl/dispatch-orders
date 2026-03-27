# ── Variables ─────────────────────────────────────────────────────────────────
APP_NAME   = dispatch-orders
CMD_PATH   = ./cmd/api
BINARY     = bin/$(APP_NAME)
DOCKER_DB  = dispatch-orders-db

# ── Build ─────────────────────────────────────────────────────────────────────
.PHONY: build
build:
	@echo "→ Building $(APP_NAME)..."
	@mkdir -p bin
	go build -o $(BINARY) $(CMD_PATH)
	@echo "✓ Binary ready at $(BINARY)"

.PHONY: run
run:
	go run $(CMD_PATH)/main.go

.PHONY: clean
clean:
	@rm -rf bin/
	@echo "✓ Cleaned"

# ── Database ──────────────────────────────────────────────────────────────────
.PHONY: db-up
db-up:
	docker compose up postgres -d
	@echo "✓ PostgreSQL running on port 5433"

.PHONY: db-down
db-down:
	docker compose down

.PHONY: db-logs
db-logs:
	docker compose logs -f postgres

.PHONY: db-shell
db-shell:
	docker exec -it $(DOCKER_DB) psql -U postgres -d dispatch_orders

# ── Dev ───────────────────────────────────────────────────────────────────────
.PHONY: dev
dev: db-up
	@echo "→ Starting server..."
	go run $(CMD_PATH)/main.go

.PHONY: tidy
tidy:
	go mod tidy

# ── Tests ─────────────────────────────────────────────────────────────────────
.PHONY: test
test:
	go test ./... -v -count=1

.PHONY: test-coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Report at coverage.html"

# ── Lint ──────────────────────────────────────────────────────────────────────
.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# ── Help ──────────────────────────────────────────────────────────────────────
.PHONY: help
help:
	@echo ""
	@echo "  $(APP_NAME) — available commands:"
	@echo ""
	@echo "  make build          Build the binary"
	@echo "  make run            Run without building"
	@echo "  make dev            Start DB + server"
	@echo "  make db-up          Start PostgreSQL only"
	@echo "  make db-down        Stop all containers"
	@echo "  make db-shell       Open psql shell"
	@echo "  make test           Run all tests"
	@echo "  make test-coverage  Tests + HTML coverage report"
	@echo "  make lint           Run linter"
	@echo "  make tidy           go mod tidy"
	@echo "  make clean          Remove binaries"
	@echo ""
