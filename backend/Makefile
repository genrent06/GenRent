.PHONY: run build tidy docker-up docker-down seed-admin migrate-up migrate-down migrate-status migrate-reset

# Run the server (run from backend/ directory, frontend served from ../frontend)
run:
	go run ./cmd/main.go

# Build binary
build:
	go build -o bin/genrent ./cmd/main.go

# Build migrate command
build-migrate:
	go build -o bin/migrate ./cmd/migrate/main.go

# Tidy dependencies
tidy:
	go mod tidy

# Start PostgreSQL via Docker
docker-up:
	docker-compose up -d postgres
	@echo "PostgreSQL started on port 5432"
	@echo "Database: genrent | User: postgres | Password: postgres"

# Start all services
docker-all:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# ── Database Migrations (using Goose) ────────────────────────────────────────

# Build and run migrations up
migrate-up: build-migrate
	./bin/migrate up

# Rollback one migration
migrate-down: build-migrate
	./bin/migrate down

# Check migration status
migrate-status: build-migrate
	./bin/migrate status

# Rollback all migrations
migrate-reset: build-migrate
	./bin/migrate reset

# Run migrations in Docker container
migrate-docker-up:
	docker-compose exec app go run ./cmd/migrate/main.go up

migrate-docker-down:
	docker-compose exec app go run ./cmd/migrate/main.go down

migrate-docker-status:
	docker-compose exec app go run ./cmd/migrate/main.go status

# Create admin user (run after server is running)
seed-admin:
	@echo "Creating admin user..."
	curl -s -X POST http://localhost:8080/api/auth/register \
		-H "Content-Type: application/json" \
		-d '{"name":"Admin","email":"admin@genrent.in","phone":"9999999999","password":"admin123","role":"admin"}' \
		| python3 -m json.tool
	@echo "\nAdmin created: admin@genrent.in / admin123"

# Format code
fmt:
	gofmt -w .

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# ── Production ───────────────────────────────────────────────────────────────

# Deploy: pull latest, rebuild, restart (zero-downtime via graceful shutdown)
deploy:
	git pull
	docker compose --env-file .env.production build --no-cache
	docker compose --env-file .env.production up -d
	@echo "Deployed. Logs: make logs"

# View live app logs
logs:
	docker compose logs -f app

# View DB logs
logs-db:
	docker compose logs -f postgres

# Run DB migrations in production
migrate-prod:
	docker compose --env-file .env.production exec app ./migrate up

# Manual DB backup (auto backup runs via cron)
backup:
	bash scripts/backup.sh

# View Caddy (HTTPS proxy) logs
logs-caddy:
	docker compose logs -f caddy

# Reload Caddy config without restart (after editing Caddyfile)
caddy-reload:
	docker compose exec caddy caddy reload --config /etc/caddy/Caddyfile
