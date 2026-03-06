DC=docker compose

.PHONY: up down logs deploy
auto:
	@echo "Use make up|down|logs|deploy"

up:
	$(DC) up -d --build

down:
	$(DC) down -v

logs:
	$(DC) logs -f

deploy:
	$(DC) -f docker-compose.yml -f docker-compose.deploy.yml up -d --build
