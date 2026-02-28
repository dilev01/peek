VERSION := 0.1.0
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build run test clean install lint

build:
	go build $(LDFLAGS) -o bin/peek ./cmd/peek/

run:
	go run $(LDFLAGS) ./cmd/peek/ $(ARGS)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

install: build
	cp bin/peek /usr/local/bin/peek

uninstall:
	rm -f /usr/local/bin/peek
