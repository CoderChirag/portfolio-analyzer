.Phony: build

build:
	CGO_ENABLED=0 go build -o folio-tracker ./cmd

start: build
	DB_HOST=localhost DB_PORT=3306 DB_ROOT_USER=root DB_ROOT_PASSWORD=root DB_FOLIO_USER=folio DB_FOLIO_PASSWORD=folio DB_NAME=folio DB_POOL_SIZE=5 MIGRATIONS_DIR=lib/db/migrations ./folio-tracker

test:
	go test -v ./...