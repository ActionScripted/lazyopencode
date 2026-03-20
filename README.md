# lazyopencode

lazygit for opencode — browse, search, and preview opencode sessions; group by workspace; yank paths or IDs; open a shell in any session's directory; delete sessions individually or in bulk.

## Requirements

- [opencode](https://opencode.ai)
- [Go](https://golang.org/dl/) 1.25+

## Install

```sh
make install
```

Builds to `build/lazyopencode` and symlinks to `~/.local/bin/lazyopencode`.

## Usage

```sh
lazyopencode
```

## Keybinds

| Key | Context | Action |
|-----|---------|--------|
| `j` / `↓` | Normal, Workspaces | Move down |
| `k` / `↑` | Normal, Workspaces | Move up |
| `enter` | Normal | Open selected session in opencode |
| `/` | Normal | Search / filter |
| `w` | Normal | Workspaces view |
| `y` | Normal | Yank sub-menu |
| `g` | Normal | Goto sub-menu |
| `d` | Normal | Delete selected session |
| `q` / `esc` | Normal | Quit |
| `esc` | Search | Return to normal mode |
| `d` | Workspaces | Delete all sessions in selected workspace |
| `w` / `esc` | Workspaces | Return to normal mode |
| `d` | Yank | Copy session directory to clipboard |
| `s` | Yank | Copy session ID to clipboard |
| `esc` / `q` | Yank | Cancel |
| `s` | Goto | Open `$SHELL` in the session's directory |
| `w` | Goto | Jump to that session's workspace |
| `esc` / `q` | Goto | Cancel |
| `y` / `d` | Delete confirm | Confirm |
| `n` / `esc` | Delete confirm | Cancel |

## Development

Install dev tools:

```sh
brew install golangci-lint
go install golang.org/x/tools/cmd/goimports@latest
```

| Command | What it does |
|---------|-------------|
| `make check` | fmt + vet + lint + test (run before committing) |
| `make fmt` | Format with gofmt and goimports |
| `make vet` | Run go vet |
| `make lint` | Run golangci-lint |
| `make test` | Run go test ./... |

## Releases

Releases are cut by pushing a semver tag. The GitHub Actions release workflow
builds binaries for all supported platforms and publishes them to GitHub
Releases with auto-generated release notes.

```sh
git tag v0.1.0
git push origin v0.1.0
```

Supported platforms: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`.

CI runs on every push to `main` and on all pull requests (`vet` + `lint` + `test`).

## Contributing

Issues and PRs are welcome. Run `make check` before submitting — it covers formatting, vetting, linting, and tests. No formal contributing process is defined yet.

## License

[MIT](LICENSE)
