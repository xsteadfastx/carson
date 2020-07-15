.PHONY: generate build clean test lint dep-update

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
            echo v0)

generate:
	GOFLAGS=-mod=vendor go generate ./...

build:
	\
		CGO_ENABLED=0 \
		gox \
		-osarch="linux/amd64" \
		-osarch="linux/386" \
		-osarch="linux/arm" \
		-osarch="linux/arm64" \
		-mod vendor \
		-ldflags '-extldflags "-static" -X "main.version=${VERSION}"' \
		.

clean:
	rm -f carson
	rm -f carson_linux_*

test:
	GOFLAGS=-mod=vendor go test -race -cover -v ./...

lint:
	golangci-lint run --enable-all --disable gomnd --disable godox

dep-update:
	go get -u ./...
	go test ./...
	go mod tidy
	go mod vendor
