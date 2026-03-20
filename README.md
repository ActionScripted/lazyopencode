# lazyopencode

lazygit for opencode — browse, search, and preview opencode sessions; group by workspace; yank paths or IDs; open a shell in any session's directory; delete sessions individually or in bulk.

## Requirements

- [opencode](https://opencode.ai)

## Install

### Homebrew (recommended)

    brew install actionscripted/lazyopencode/lazyopencode

To upgrade:

    brew upgrade lazyopencode

**Other methods:**

- download a binary from [GitHub Releases](https://github.com/ActionScripted/lazyopencode/releases/latest) and move it onto your `$PATH`
- run `go install github.com/actionscripted/lazyopencode@latest` (requires Go 1.25+)
- build from source with `make install` (requires Go 1.25+, symlinks to `~/.local/bin/lazyopencode`).

## Usage

```sh
lazyopencode
```

## Keybinds

We're going for Vim-style binds for speed. We have a normal mode with sessions, search and a workspace mode for extra laziness.

### Modes

| Key | Mode       | Notes                                                    |
| --- | ---------- | -------------------------------------------------------- |
| `/` | Search     | Type to filter; `esc` to exit                            |
| `w` | Workspaces | `d` deletes all sessions in workspace; `w`/`esc` to exit |

### Movement

| Key         | Action                |
| ----------- | --------------------- |
| `j` / `↓`   | Down                  |
| `k` / `↑`   | Up                    |
| `enter`     | Open selected session |
| `q` / `esc` | Quit / back           |

### Chords

| Keys | Action                               |
| ---- | ------------------------------------ |
| `dd` | Delete selected session              |
| `gs` | Open `$SHELL` in session's directory |
| `gw` | Jump to session's workspace          |
| `yd` | Yank session directory to clipboard  |
| `ys` | Yank session ID to clipboard         |

## Development

Install dev tools:

```sh
brew install golangci-lint
go install golang.org/x/tools/cmd/goimports@latest
```

| Command      | What it does                                    |
| ------------ | ----------------------------------------------- |
| `make check` | fmt + vet + lint + test (run before committing) |
| `make fmt`   | Format with gofmt and goimports                 |
| `make vet`   | Run go vet                                      |
| `make lint`  | Run golangci-lint                               |
| `make test`  | Run go test ./...                               |

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
