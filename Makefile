.PHONY: all build run stop clean dev backend frontend docker-build docker-up docker-down

# Colors
GREEN = \033[0;32m
YELLOW = \033[1;33m
NC = \033[0m

all: build

# Build everything
build: backend frontend
	@echo "$(GREEN)Build complete!$(NC)"

# Build backend
backend:
	@echo "$(YELLOW)Building backend...$(NC)"
	cd backend && go build -o vsentry .
	@echo "$(GREEN)Backend built: ./backend/vsentry$(NC)"

# Build frontend
frontend:
	@echo "$(YELLOW)Building frontend...$(NC)"
	cd frontend && npm install && npm run build
	@echo "$(GREEN)Frontend built: ./frontend/dist$(NC)"

# Run in development mode
dev:
	@echo "$(YELLOW)Starting development servers...$(NC)"
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:5173"
	cd backend && go run main.go &
	cd frontend && npm run dev &

# Run production
run: build
	@echo "$(YELLOW)Starting VSentry...$(NC)"
	cd backend && ./vsentry &

# Stop all services
stop:
	@pkill -f vsentry || true
	@echo "$(GREEN)Services stopped$(NC)"

# Clean build artifacts
clean:
	rm -f backend/vsentry
	rm -rf frontend/dist
	rm -f backend/*.log
	rm -f backend/vsentry.db
	rm -rf backend/badger_data
	@echo "$(GREEN)Clean complete$(NC)"

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d
	@echo "$(GREEN)VSentry running at http://localhost:8088$(NC)"

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v

# Install dependencies
deps-backend:
	cd backend && go mod download

deps-frontend:
	cd frontend && npm install

# Default target
help:
	@echo "VSentry Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build backend and frontend"
	@echo "  make backend        - Build backend only"
	@echo "  make frontend       - Build frontend only"
	@echo "  make dev            - Run in development mode"
	@echo "  make run            - Run production"
	@echo "  make stop           - Stop all services"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker images"
	@echo "  make docker-up      - Start with Docker"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make deps-backend   - Install backend dependencies"
	@echo "  make deps-frontend  - Install frontend dependencies"