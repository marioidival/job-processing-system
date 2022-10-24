BIN := $(shell pwd)/bin
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_CMD := docker-compose -p job-system -f $(DOCKER_COMPOSE_FILE)
DB_CONN = postgres://jobsystem:jobsystem@localhost:5432/jobsystem?sslmode=disable
SOURCE_MIGRATION = file://internal/db/schema/migrations
SCHEMA_DB_NAME := schema-$(shell date +"%s")
SCHEMA_DB_URL := "postgres://postgres:postgres@localhost:5432/$(SCHEMA_DB_NAME)?sslmode=disable"
SCHEMA_FILE_PATH := ./internal/db/schema/schema.sql

generate/schema:
	$(DOCKER_COMPOSE_CMD) up -d postgres
	for i in 1 2 3 4 5; do pg_isready -h localhost -p 5432 -t 3 -U postgres && break || sleep 3; done
	$(DOCKER_COMPOSE_CMD) exec postgres createdb $(SCHEMA_DB_NAME)
	$(BIN)/migrate -source $(SOURCE_MIGRATION) -database $(SCHEMA_DB_URL) up 1
	pg_dump --schema-only --no-owner -d $(SCHEMA_DB_URL) > $(SCHEMA_FILE_PATH)
	$(DOCKER_COMPOSE_CMD) exec postgres dropdb $(SCHEMA_DB_NAME)


$(BIN)/migrate: go.mod go.sum
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate



.PHONY: db/migrate
db/migrate: $(BIN)/migrate
	sleep 2
	$(BIN)/migrate -source $(SOURCE_MIGRATION) -database $(DB_CONN) up 1


$(BIN)/sqlc: go.mod go.sum
	$(GO) install github.com/kyleconroy/sqlc/cmd/sqlc

generate/queries: $(BIN)/sqlc $(SCHEMA_FILE_PATH) ## Generate queries.
	$(BIN)/sqlc -f ./internal/db/sqlc.yml compile
	$(BIN)/sqlc -f ./internal/db/sqlc.yml generate

$(BIN)/worker:
	$(GO) install ./cmd/worker

$(BIN)/api:
	$(GO) install ./cmd/api

start/worker: $(BIN)/worker
	DATABASE_URL=$(DB_CONN) sh -c '$(BIN)/worker'

start/api: up db/migrate $(BIN)/api
	DATABASE_URL=$(DB_CONN) sh -c '$(BIN)/api'

.PHONE: up
up:
	$(DOCKER_COMPOSE_CMD) up -d postgres