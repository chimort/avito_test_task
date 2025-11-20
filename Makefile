COMPOSE=docker-compose

.PHONY: all build up down test

all: build up

build:
	$(COMPOSE) build

up:
	$(COMPOSE) up

down:
	$(COMPOSE) down -v

test:
	go test -count=1 ./tests/handlers ./tests/service ./tests/repository