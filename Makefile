BINARY  := build/lazyoc
LINK    := $(HOME)/.local/bin/lazyoc
SRC     := $(wildcard *.go)

.PHONY: build link install clean

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
