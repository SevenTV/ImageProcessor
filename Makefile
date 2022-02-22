all: linux

BUILDER := "unknown"
VERSION := "unknown"

ifeq ($(origin IMAGES_BUILDER),undefined)
	BUILDER = $(shell git config --get user.name);
else
	BUILDER = ${IMAGES_BUILDER};
endif

ifeq ($(origin IMAGES_VERSION),undefined)
	VERSION = $(shell git rev-parse HEAD);
else
	VERSION = ${IMAGES_VERSION};
endif

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o bin/images .

lint:
	staticcheck ./...
	go vet ./...
	golangci-lint run

deps:
	CGO_ENABLED=0 go mod download
	CGO_ENABLED=0 go install honnef.co/go/tools/cmd/staticcheck@latest
	CGO_ENABLED=0 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

test:
	go test -count=1 -cover ./...
