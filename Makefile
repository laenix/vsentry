.PHONY: all build run stop clean dev backend frontend agent-linux agent-windows docker-build docker-up docker-down docker-clean deps-backend deps-frontend help

# Colors
GREEN = \033[0;32m
YELLOW = \033[1;33m
NC = \033[0m

all: build

# Build everything for local production run
build: backend frontend
	@echo "$(GREEN)Build complete!$(NC)"

# Build backend
backend:
	@echo "$(YELLOW)Building backend...$(NC)"
	cd backend && CGO_ENABLED=0 go build -trimpath -o vsentry .
	@echo "$(GREEN)Backend built: ./backend/vsentry$(NC)"

# Build frontend
frontend:
	@echo "$(YELLOW)Building frontend...$(NC)"
	cd frontend && npm install && npm run build
	@echo "$(GREEN)Frontend built: ./frontend/dist$(NC)"

# Build Agent locally for testing/debugging OCSF rules (Not used in Docker prod)
agent-linux:
	@echo "$(YELLOW)Building Linux Agent for local test...$(NC)"
	cd backend/cmd/collectors && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o vsentry-agent .
	@echo "$(GREEN)Agent built: ./backend/cmd/collectors/vsentry-agent$(NC)"

agent-windows:
	@echo "$(YELLOW)Building Windows Agent for local test...$(NC)"
	cd backend/cmd/collectors && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o vsentry-agent.exe .
	@echo "$(GREEN)Agent built: ./backend/cmd/collectors/vsentry-agent.exe$(NC)"

# Run in development mode (Smartly starts VictoriaLogs dependency first)
dev:
	@echo "$(YELLOW)Starting VictoriaLogs dependency via Docker...$(NC)"
	docker-compose up -d victorialogs
	@echo "$(YELLOW)Starting development servers...$(NC)"
	@echo "Backend: http://localhost:8088"
	@echo "Frontend: http://localhost:5173"
	cd backend && VICTORIALOGS_URL=http://localhost:9428 EXTERNAL_URL=http://localhost:8088 go run . &
	cd frontend && npm run dev &

# Run local production
run: build
	@echo "$(YELLOW)Starting VSentry...$(NC)"
	cd backend && VICTORIALOGS_URL=http://localhost:9428 ./vsentry &

# Stop all local services
stop:
	@pkill -f vsentry || true
	@pkill -f "npm run dev" || true
	@echo "$(GREEN)Local services stopped$(NC)"

# Clean build artifacts and local agent caches
clean:
	rm -f backend/vsentry
	rm -f backend/cmd/collectors/vsentry-agent*
	rm -rf frontend/dist
	rm -f backend/*.log
	rm -f backend/vsentry.db
	rm -rf /tmp/.vsentry*_cache  # Clean up local agent DLQ and Bookmarks
	@echo "$(GREEN)Clean complete$(NC)"

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d --build
	@echo "$(GREEN)VSentry running at http://localhost:8088$(NC)"

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v
	@echo "$(GREEN)Docker volumes and containers removed$(NC)"

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
	@echo "  make build         - Build backend and frontend"
	@echo "  make backend       - Build backend only"
	@echo "  make frontend      - Build frontend only"
	@echo "  make agent-linux   - Build Linux agent locally for debugging"
	@echo "  make agent-windows - Build Windows agent locally for debugging"
	@echo "  make dev           - Run in dev mode (Auto-starts VictoriaLogs)"
	@echo "  make run           - Run local production"
	@echo "  make stop          - Stop all local services"
	@echo "  make clean         - Clean build artifacts & local cache"
	@echo "  make docker-build  - Build Docker images"
	@echo "  make docker-up     - Start the full stack with Docker"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make docker-clean  - Stop containers and WIPE ALL DATA"
	@echo "  make deps-backend  - Install backend dependencies"
	@echo "  make deps-frontend - Install frontend dependencies"