# Define the Docker Compose command
DCOMPOSE = docker-compose

# Start all services in detached mode
up:
	@echo "Starting all services..."
	$(DCOMPOSE) up -d

# Stop all running containers
down:
	@echo "Stopping all services..."
	$(DCOMPOSE) down

# Restart all services
restart: down up

# Show running containers
ps:
	@echo "Listing running containers..."
	$(DCOMPOSE) ps

# View logs of all services
logs:
	@echo "Displaying logs..."
	$(DCOMPOSE) logs -f

# Rebuild images without using cache
build:
	@echo "Building images..."
	$(DCOMPOSE) build --no-cache

# Remove all volumes (use with caution, this deletes all stored data)
clean:
	@echo "Removing all volumes..."
	$(DCOMPOSE) down -v

# Execute a shell inside the PostgreSQL container
postgres-shell:
	@echo "Connecting to PostgreSQL shell..."
	docker exec -it postgres_db psql -U user -d mydb

# Execute Redis CLI inside the Redis container
redis-cli:
	@echo "Connecting to Redis CLI..."
	docker exec -it redis_cache redis-cli

# Execute a shell inside the Memcached container
memcached-cli:
	@echo "Connecting to Memcached (via telnet)..."
	telnet localhost 11211

# Build and run the Go application
app-build:
	@echo "Building the Go application..."
	docker build --no-cache -t multi-tier-caching-example .

# Run the Go application in a container
app-run: app-build
	@echo "Running the Go application..."
	docker run --rm --name multi-tier-caching-example --network host my-go-app
	#docker run --rm --name multi-tier-caching-example --network host --entrypoint /bin/sh -it my-go-app

test:
	@echo "Running cache_benchmark_test.go..."
	export GOROOT=/usr/local/go
	export PATH=$PATH:$GOROOT/bin
	go test -v
