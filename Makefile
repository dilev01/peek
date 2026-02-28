VERSION := 0.1.0
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build run test clean install cross-compile

build:
	go build $(LDFLAGS) -o bin/peek ./cmd/peek/

run:
	go run $(LDFLAGS) ./cmd/peek/ $(ARGS)

test:
	go test ./...

clean:
	rm -rf bin/

install: build
	cp bin/peek /usr/local/bin/peek

cross-compile:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/peek-darwin-arm64 ./cmd/peek/
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/peek-darwin-amd64 ./cmd/peek/
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/peek-linux-amd64 ./cmd/peek/
