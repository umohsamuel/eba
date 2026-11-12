docker-run:
	@docker compose -f docker-compose.yaml up -d

docker-down:
	@docker compose -f docker-compose.yaml down

docker-ps:
	@docker compose -f docker-compose.yaml ps

docker-logs:
	@docker compose -f docker-compose.yaml logs -f --tail=100

.PHONY: docker-run docker-down docker-ps docker-logs
