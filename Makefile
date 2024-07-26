COMPOSE_FILE = docker-compose.yml

all: build up

# Build the Docker images
build:
	@echo "Building Docker images..."
	docker-compose -f $(COMPOSE_FILE) build

# Start the containers
up:
	@echo "Starting containers..."
	docker-compose -f $(COMPOSE_FILE) up -d

# Stop the containers
down:
	@echo "Stopping containers..."
	docker-compose -f $(COMPOSE_FILE) down

# Restart the containers
restart: down up

# Remove stopped containers and unused images
clean:
	@echo "Cleaning up unused Docker resources..."
	docker-compose -f $(COMPOSE_FILE) down -v
	docker system prune -f

# Display container logs
logs:
	@echo "Displaying container logs..."
	docker-compose -f $(COMPOSE_FILE) logs -f

# Tail logs for a specific service
logs-%:
	@echo "Tailing logs for service $*..."
	docker-compose -f $(COMPOSE_FILE) logs -f $*

# Help message
help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all       Build and start the application (default)"
	@echo "  build     Build Docker images"
	@echo "  up        Start the containers"
	@echo "  down      Stop the containers"
	@echo "  restart   Restart the containers"
	@echo "  clean     Clean up unused Docker resources"
	@echo "  logs      Display container logs"
	@echo "  logs-<service> Tail logs for a specific service (e.g., logs-prometheus)"
	@echo "  help      Display this help message"

.PHONY: all build up down restart clean logs help logs-%
