.Phony: build

build:
	CGO_ENABLED=0 go build -o folio-tracker ./cmd

start: build
	./folio-tracker