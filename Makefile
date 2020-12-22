os ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')

.PHONY: build
build: dist/prego

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
dist/prego: $(SRC)
	go build -o dist/prego