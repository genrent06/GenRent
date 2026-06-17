.PHONY: run build build-migrate tidy docker-up docker-all docker-down \
       migrate-up migrate-down migrate-status migrate-reset \
       migrate-docker-up migrate-docker-down migrate-docker-status \
       seed-admin fmt dev deploy logs logs-db migrate-prod backup \
       logs-caddy caddy-reload

run build build-migrate tidy docker-up docker-all docker-down \
migrate-up migrate-down migrate-status migrate-reset \
migrate-docker-up migrate-docker-down migrate-docker-status \
seed-admin fmt dev deploy logs logs-db migrate-prod backup \
logs-caddy caddy-reload:
	$(MAKE) -C backend $@
