BINARY  := build/lazyopencode
LINK    := $(HOME)/.local/bin/lazyopencode
SRC     := $(wildcard *.go)

.PHONY: build link install clean fmt vet lint test cover check release

build: $(SRC)
	@mkdir -p build
	go build -o $(BINARY) .

link:
	@mkdir -p $(dir $(LINK))
	ln -sf $(CURDIR)/$(BINARY) $(LINK)

install: build link

clean:
	rm -f $(BINARY)
	rm -f $(LINK)

fmt:
	gofmt -w .
	@which goimports > /dev/null && goimports -w . || (echo "goimports not installed: go install golang.org/x/tools/cmd/goimports@latest"; exit 1)

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed: https://golangci-lint.run/usage/install/"; exit 1)
	golangci-lint run

test:
	go test -race -count=1 ./...

cover:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

check: fmt vet lint test

release: check
	@latest=$$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | head -1); \
	suggested=$$(echo $$latest | awk -F. '{$$NF=$$NF+1; print $$0}' OFS=.); \
	printf "Latest tag: $$latest\nNew tag [$$suggested]: "; \
	read input; \
	new=$${input:-$$suggested}; \
	printf "Tag $$new — push? [y/N]: "; \
	read confirm; \
	case $$confirm in \
		[yY]) git tag $$new && git push origin $$new ;; \
		*) echo "Aborted." ;; \
	esac
