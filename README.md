# lazyoc

lazygit for opencode — a terminal UI for managing opencode sessions.

## Requirements

- [opencode](https://opencode.ai)
- [Go](https://golang.org/dl/) 1.25+
- [golangci-lint](https://golangci-lint.run/usage/install/) — required for `make lint` / `make check`
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) — required for `make fmt` / `make check`

```sh
# golangci-lint
brew install golangci-lint

# goimports
go install golang.org/x/tools/cmd/goimports@latest
```

## Install

```sh
make install
```

Builds to `build/lazyoc` and symlinks to `~/.local/bin/lazyoc`.

## Usage

```sh
lazyoc
```

## Keybinds

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate down / up |
| `enter` | Open selected session in opencode |
| `/` | Search / filter |
| `tab` | Toggle sessions / paths view |
| `ctrl-s` | Toggle subagent sessions |
| `p` | Toggle preview pane |
| `ctrl-d` | Delete selected session or path |
| `ctrl-x` | Delete all sessions |
| `q` / `esc` | Quit |

## Development

```sh
make check   # fmt + vet + lint + test (run before committing)
make fmt     # format with gofmt and goimports
make vet     # run go vet
make lint    # run golangci-lint
make test    # run go test ./...
```
