DC=docker compose

.PHONY: up down logs deploy
up:
	$(DC) -f docker-compose.yml -f docker-compose.local.yml up -d --build

down:
	$(DC) -f docker-compose.yml -f docker-compose.local.yml down -v

logs:
	$(DC) logs -f

deploy:
	$(DC) -f docker-compose.yml -f docker-compose.deploy.yml up -d --build
