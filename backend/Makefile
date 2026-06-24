GOPATH_BIN := $(HOME)/go/bin
GO := $(GOPATH_BIN)/go
DB_URL ?= $(shell grep DATABASE_URL .env | cut -d= -f2-)

.PHONY: run migrate-up migrate-down generate tidy

run:
	$(GO) run ./cmd/api

tidy:
	$(GO) mod tidy

migrate-up:
	$(GOPATH_BIN)/migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	$(GOPATH_BIN)/migrate -path migrations -database "$(DB_URL)" down 1

generate:
	$(GOPATH_BIN)/sqlc generate
