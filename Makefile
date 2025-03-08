# Default variables
DOCKER_COMPOSE = docker compose
DEV_COMPOSE_FILES = -f docker-compose.yaml -f docker-compose.dev.yaml
DOCKER_REGISTRY = pariksh1th
IMAGE_NAME = broker-service

# testing
# Version management
VERSION := $(shell cat VERSION)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_SHA := $(shell git rev-parse --short HEAD)

# Colors for terminal output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

.PHONY: up down build dev clean restart logs help version bump-patch bump-minor bump-major tag

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
		@echo "$(COLOR_YELLOW)make version$(COLOR_RESET)   - Show current version"
		@echo "$(COLOR_YELLOW)make bump-patch$(COLOR_RESET) - Bump patch version (1.0.0 -> 1.0.1)"
		@echo "$(COLOR_YELLOW)make bump-minor$(COLOR_RESET) - Bump minor version (1.0.0 -> 1.1.0)"
		@echo "$(COLOR_YELLOW)make bump-major$(COLOR_RESET) - Bump major version (1.0.0 -> 2.0.0)"
		@echo "$(COLOR_YELLOW)make tag$(COLOR_RESET)	      - Build and tag the current version"

# Start production containers
up:
		@echo "$(COLOR_GREEN)Starting production containers version $(VERSION)...$(COLOR_RESET)"
		@VERSION=$(VERSION) BUILD_DATE=$(BUILD_DATE) COMMIT_SHA=$(COMMIT_SHA) $(DOCKER_COMPOSE) up --build -d

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
		@echo "$(COLOR_GREEN)Building containers version $(VERSION)...$(COLOR_RESET)"
		@VERSION=$(VERSION) BUILD_DATE=$(BUILD_DATE) COMMIT_SHA=$(COMMIT_SHA) $(DOCKER_COMPOSE) build

# Clean up containers, volumes, and build cache
clean:
		@echo "$(COLOR_GREEN)Cleaning up...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) down -v
		@docker system prune -f
		@rm -rf broker-service/tmp
		@rm -rf tmp

# Restart containers
restart:
	@echo "$(COLOR_GREEN)Restarting containers...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) restart

# View logs
logs:
	@echo "$(COLOR_GREEN)Viewing logs...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) logs -f

# Version management commands
version:
	@echo "Current version: $(VERSION)"

bump-patch:
	@echo "$(VERSION)" | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g' > VERSION
	@echo "Bumped patch version to $$(cat VERSION)"

bump-minor:
	@echo "$(VERSION)" | awk -F. '{$$2 = $$2 + 1; $$3 = 0;} 1' | sed 's/ /./g' > VERSION
	@echo "Bumped minor version to $$(cat VERSION)"

bump-major:
	@echo "$(VERSION)" | awk -F. '{$$1 = $$1 + 1; $$2 = 0; $$3 = 0;} 1' | sed 's/ /./g' > VERSION
	@echo "Bumped major version to $$(cat VERSION)"

tag: build
	@echo "$(COLOR_GREEN)Tagging version $(VERSION)...$(COLOR_RESET)"
	@git tag -a v$(VERSION) -m "Version $(VERSION)"
	@git push --follow-tags
	@echo "$(COLOR_GREEN)Pushing Docker images version $(VERSION)...$(COLOR_RESET)"
	@docker tag $(IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	@docker tag $(IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest
	@docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest
	@echo "$(COLOR_GREEN)Successfully pushed version $(VERSION) to Docker registry$(COLOR_RESET)"

