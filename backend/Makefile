.PHONY: run build tidy docker-up docker-down seed-admin \
	dev deploy logs logs-db migrate-prod backup \
	logs-caddy caddy-reload

# Run the server (migrations run automatically on startup)
run:
	go run ./cmd/main.go

# Build binary
build:
	go build -o bin/genrent ./cmd/main.go

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

# Create admin user (run after server is running)
seed-admin:
	@echo "Creating admin user..."
	curl -s -X POST http://localhost:8080/api/v1/auth/register \
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

# Manual DB backup (auto backup runs via cron)
backup:
	bash scripts/backup.sh

# View Caddy (HTTPS proxy) logs
logs-caddy:
	docker compose logs -f caddy

# Reload Caddy config without restart (after editing Caddyfile)
caddy-reload:
	docker compose exec caddy caddy reload --config /etc/caddy/Caddyfile
