SHELL := /bin/bash
BASEDIR = $(shell pwd)

all: 
	gotool
	@go build -v .
proto:
	protoc --proto_path=. --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/*.proto
clean:
	rm -f go-tag-service
	find . -name "[._]*.s[a-w][a-z]" | xargs -i rm -f {}
gotool:
	gofmt -w .
	go vet . |& grep -v vendor;true
help:
	@echo "make - compile the source code"
	@echo "make clean - remove binary file and vim swp files"
	@echo "make gotool - run go tool 'fmt' and 'vet'"

.PHONY: proto clean gotool help