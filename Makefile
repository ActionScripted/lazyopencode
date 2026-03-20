BINARY  := build/lazyoc
LINK    := $(HOME)/.local/bin/lazyoc
SRC     := $(wildcard *.go)

.PHONY: build link install clean fmt vet lint test check

build: $(SRC)
	@mkdir -p build
	go build -o $(BINARY) .

link:
	@mkdir -p $(dir $(LINK))
	ln -sf $(CURDIR)/$(BINARY) $(LINK)

install: build link

clean:
	rm -f $(BINARY) lazyoc
	rm -f $(LINK)

fmt:
	gofmt -w .
	@which goimports > /dev/null && goimports -w . || echo "goimports not installed: go install golang.org/x/tools/cmd/goimports@latest"

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed: https://golangci-lint.run/usage/install/"; exit 1)
	golangci-lint run

test:
	go test ./...

check: fmt vet lint test
