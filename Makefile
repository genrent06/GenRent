.PHONY: run build tidy docker-up docker-all docker-down \
       seed-admin fmt dev deploy logs logs-db \
       backup logs-caddy caddy-reload

run build tidy docker-up docker-all docker-down \
seed-admin fmt dev deploy logs logs-db \
backup logs-caddy caddy-reload:
	$(MAKE) -C backend $@
