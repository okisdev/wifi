VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS  = -s -w -X github.com/okisdev/wifi/cmd.version=$(VERSION) \
           -X github.com/okisdev/wifi/cmd.commit=$(COMMIT) \
           -X github.com/okisdev/wifi/cmd.date=$(DATE)

.PHONY: build build-darwin build-linux build-windows clean lint

build:
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o bin/wifi .

build-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o bin/wifi-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o bin/wifi-darwin-arm64 .

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/wifi-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/wifi-linux-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/wifi-windows-amd64.exe .

clean:
	rm -rf bin/

lint:
	go vet ./...
