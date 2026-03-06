DC=docker compose

.PHONY: up down logs test deploy
up:
	$(DC) -f docker-compose.yml -f docker-compose.local.yml up -d --build

down:
	$(DC) -f docker-compose.yml -f docker-compose.local.yml down -v

logs:
	$(DC) logs -f

test:
	docker run --rm -v "$(CURDIR)/backend:/src" -w /src golang:1.23-alpine sh -lc '/usr/local/go/bin/go test ./...'

deploy:
	$(DC) -f docker-compose.yml -f docker-compose.deploy.yml up -d --build
