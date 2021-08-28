
# Makefile for broker-go

VERSION ?= 0.0.0
BINARY_NAME ?= ticker-broker

build:

test:

archive:

release:

build-prerequisites:
	mkdir -p bin dist

release-prerequisites:

test-prerequisites:

install-tools:

### BUILD ###################################################################

build-broker-go: build-prerequisites
	go build -ldflags "-X main.version=${VERSION} -X main.commit=$$(git rev-parse --short HEAD 2>/dev/null || echo \"none\")" -o bin/$(OUTPUT_DIR)$(BINARY_NAME) cli/main.go
build-broker-go-linux_amd64: build-prerequisites
	$(MAKE) GOOS=linux GOARCH=amd64 OUTPUT_DIR=linux_amd64/ build
build-broker-go-darwin_amd64: build-prerequisites
	$(MAKE) GOOS=darwin GOARCH=amd64 OUTPUT_DIR=darwin_amd64/ build
build-broker-go-windows_amd64: build-prerequisites
	$(MAKE) GOOS=windows GOARCH=amd64 OUTPUT_DIR=windows_amd64/ build

build-linux_amd64: build-broker-go-linux_amd64
build-darwin_amd64: build-broker-go-darwin_amd64
build-windows_amd64: build-broker-go-windows_amd64

build: build-broker-go
build-all: build-linux_amd64 build-darwin_amd64 build-windows_amd64

### ARCHIVE #################################################################

archive-broker-go-linux_amd64: build-broker-go-linux_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-linux_amd64.tar.gz -C bin/linux_amd64/ .
archive-broker-go-darwin_amd64: build-broker-go-darwin_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-darwin_amd64.tar.gz -C bin/darwin_amd64/ .
archive-broker-go-windows_amd64: build-broker-go-windows_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-windows_amd64.tar.gz -C bin/windows_amd64/ .

archive-linux_amd64: archive-broker-go-linux_amd64
archive-darwin_amd64: archive-broker-go-darwin_amd64
archive-windows_amd64: archive-broker-go-windows_amd64

archive: archive-linux_amd64 archive-darwin_amd64 archive-windows_amd64

release: archive
	sha1sum dist/*.tar.gz > dist/$(BINARY_NAME)-${VERSION}.shasums

### TEST ####################################################################

mock-certs:
	certstrap --depot-path tls init -o ticker -ou infosec -c DE --st HESSEN --passphrase '' --cn ticker-ca
	certstrap --depot-path tls request-cert --passphrase '' -o ticker -c DE --st HESSEN --ou infosec --cn ticker-broker
	certstrap --depot-path tls sign --passphrase '' --CA ticker-ca --csr tls/ticker-broker.csr 127.0.0.1
	certstrap --depot-path tls request-cert --passphrase '' -o ticker -c DE --st HESSEN --ou infosec --cn ticker-client

test-broker-go:
	ginkgo
test-broker-go-watch:
	ginkgo watch
test: test-broker-go
.PHONY: test-broker-go
.PHONY: test

clean:
	rm -r bin/* dist/*

### DATABASE ################################################################

db-up:
	psql < db/up.sql

db-down:
	psql < db/down.sql

db-seed:
	psql < db/seed.sql

