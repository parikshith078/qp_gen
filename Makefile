# Default variables
DOCKER_COMPOSE = docker compose
DEV_COMPOSE_FILES = -f docker-compose.yaml -f docker-compose.dev.yaml

# Colors for terminal output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

.PHONY: up down build dev clean restart logs help

# Default target when just running 'make'
help:
		@echo "$(COLOR_GREEN)Available commands:$(COLOR_RESET)"
		@echo "$(COLOR_YELLOW)make up$(COLOR_RESET)		   - Start production containers"
		@echo "$(COLOR_YELLOW)make down$(COLOR_RESET)	   - Stop and remove containers"
		@echo "$(COLOR_YELLOW)make build$(COLOR_RESET)	   - Build production containers"
		@echo "$(COLOR_YELLOW)make dev$(COLOR_RESET)	   - Start development environment with hot-reload"
		@echo "$(COLOR_YELLOW)make clean$(COLOR_RESET)	   - Clean up all containers and volumes"
		@echo "$(COLOR_YELLOW)make restart$(COLOR_RESET)   - Restart all containers"
		@echo "$(COLOR_YELLOW)make logs$(COLOR_RESET)	   - View container logs"

# Start production containers
up:
		@echo "$(COLOR_GREEN)Starting production containers...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) up -d

# Start development environment
dev:
		@echo "$(COLOR_GREEN)Starting development environment...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) $(DEV_COMPOSE_FILES) up --build

# Stop containers
down:
		@echo "$(COLOR_GREEN)Stopping containers...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) down

# Build containers
build:
		@echo "$(COLOR_GREEN)Building containers...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) build

# Clean up containers, volumes, and build cache
clean:
		@echo "$(COLOR_GREEN)Cleaning up...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) down -v
		@docker system prune -f
		@rm -rf broker-service/tmp

# Restart containers
restart:
	@echo "$(COLOR_GREEN)Restarting containers...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) restart

# View logs
logs:
	@echo "$(COLOR_GREEN)Viewing logs...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) logs -f

