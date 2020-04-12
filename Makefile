os ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')

.PHONY: build
build: prego_$(os)

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
prego_%: $(SRC)
	GOOS=$* go build -o $(@F)