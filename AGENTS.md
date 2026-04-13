# lazyopencode

lazyopencode is a terminal UI (TUI) for managing [opencode](https://opencode.ai) sessions — think lazygit but for opencode. It reads opencode's SQLite database directly (read-only) and lets you browse, search, preview, and open sessions without leaving the terminal.

## Tech stack

- **Language**: Go
- **TUI framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style model/update/view)
- **UI components**: [Bubbles](https://github.com/charmbracelet/bubbles) (text input)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Database**: opencode's SQLite DB via `modernc.org/sqlite` (pure Go, no CGO)

## Project structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point; resolves the DB path and starts the Bubble Tea program |
| `model.go` | App state (`model` struct), `Init`/`Update` logic, all command constructors (`loadMessagesForCursor`, `loadSessionsCmd`, `openSessionCmd`, etc.), message types, `resolveHome` |
| `view.go` | Layout constants, `View()`, `renderList`, `renderPreview`, `renderHint`, `renderTopBar`, `formatSessionRow`, `overlayModal`, all modals, `truncate` |
| `view_workspaces.go` | `renderWorkspacesView`, `renderWorkspaceList`, `formatWorkspaceRow`, `renderWorkspaceSessions`, `formatWorkspaceSessionRow` |
| `view_stats.go` | `renderStatsView`, `renderStats`, `renderSectionRule`, `renderTableRule`, `renderFieldset`, `buildSummaryKV`, `buildModelRows`, `renderModelHeader`, `renderModelRows`, `buildProjectRows`, `renderProjectHeader`, `renderProjectRows`, `padRight` |
| `view_format.go` | `formatDuration`, `formatDurationMS`, `fmtCommas`, `formatTokens`, `fmtCost`, `modelCost`, `modelPricing`, `knownModelPricing`, `renderHintSegments` |
| `update.go` | Key handler helpers (`updateNormal`, `updateSearch`, `updateWorkspaces`, `updateConfirmDelete`, `updateConfirmDeleteWorkspace`, `updateGoto`, `updateStats`, `updateError`, `updateYank`); `clamp` |
| `keys.go` | Key bindings (`KeyMap`) and `Mode` enum |
| `session.go` | `Session`, `Message`, `SessionStats`, `ModelStat`, `ProjectStat`, `GlobalStats` types; `homeToTilde`, `baseName` helpers; `filterSessions`, `buildWorkspaces`, `removeSessionByID` |
| `db.go` | SQLite queries — `openReadOnlyDB`, `loadSessions`, `loadMessages`, `loadStats`, `loadGlobalStats`; populates `Session.DisplayDir` and `Session.ShortDir` at load time |
| `demo.go` | `demoSessions`, `demoMessages`, `demoStats`, `demoGlobalStats`, `demoFeaturedMessages` — hardcoded fake data for `--demo` mode |
| `styles.go` | Lip Gloss color vars, style definitions, and panel-background style variants; all package-level `var`s |
| `Makefile` | `build`, `install`, `fmt`, `vet`, `lint`, `test`, `check` targets |
| `.golangci.yml` | golangci-lint configuration |
| `.editorconfig` | Editor conventions (tabs, UTF-8, trailing newlines) |

## Data flow

1. On startup, `loadSessions` queries `~/.local/share/opencode/opencode.db` (read-only).
2. `Session.DisplayDir`, `Session.ShortDir`, `Session.CreatedAt`, and the `Summary*` fields are computed/populated once at load time in `db.go` — not on every render call.
3. Sessions are displayed in a list; filtering happens in-memory via `filterSessions`.
4. When a session is selected, `loadMessagesForCursor` fires a `tea.Batch` of two commands: `loadMessagesCmd` and `loadStatsCmd`. Both run concurrently and arrive as `messagesLoadedMsg` / `statsLoadedMsg`.
5. Pressing `enter` on a session launches `opencode --session <id>` in the current directory.

## Architecture

The Bubble Tea Elm pattern is non-negotiable: **all state lives in `model`, all mutations happen in `Update`, `View` is pure.** Side effects (DB queries, process launches) happen exclusively inside `tea.Cmd` closures returned from `Update`. If you find yourself putting logic in `View` or mutating state outside of `Update`, you're doing it wrong.

New async results use typed message structs (e.g. `type fooLoadedMsg struct { ... }`). All message types live in `model.go`. This keeps the full event surface of the app visible in one place.

All methods on `model` use **value receivers**. Helper functions that need to mutate state return `(model, tea.Cmd)` — never use pointer receivers to smuggle mutations through. `loadMessagesForCursor() (model, tea.Cmd)` is the canonical example.

Styles are declared as package-level `var`s in `styles.go` and nowhere else. This means a theme change is a single-file edit. Don't declare styles inline or inside render functions.

DB functions open their own connection, run their query, and close. SQLite read-only connections are cheap — a shared pool adds complexity with no real benefit here. Don't optimize this.

Session deletion shells out to `opencode session delete <id>` rather than writing directly to the SQLite DB. This is intentional — it keeps lazyopencode read-only at the DB layer and delegates mutations to the owning process.

## Where to make changes

| If you want to… | Touch these files |
|---|---|
| Add a key binding | `keys.go` (add to `KeyMap` + `DefaultKeyMap`), `update.go` (handle with `key.Matches`) |
| Add a new style or color | `styles.go` only |
| Add a DB query | `db.go` only |
| Add a new message type | `model.go` only |
| Add a new display mode | `keys.go`, `model.go`, `update.go`, `view.go` |
| Add a session or message field | `session.go` + `db.go` |
| Add a pure session/workspace helper | `session.go` only |
| Add a cost/pricing helper | `view_format.go` only (`modelCost`, `fmtCost`) |
| Add a workspace render function | `view_workspaces.go` only |
| Add a stats render function | `view_stats.go` only |
| Add or edit demo/fake data | `demo.go` only |
| Change lint rules | `.golangci.yml` only |
| Change editor conventions | `.editorconfig` only |

## Key conventions

- All DB access is read-only (`?mode=ro`). A missing DB file (ENOENT) is treated as empty state; any other open failure (permissions, corrupt file) surfaces as an error modal.
- No CGO — the sqlite driver is `modernc.org/sqlite`.
- All key bindings go through `key.Binding` in `KeyMap`. Never match keys with raw `msg.String() ==` comparisons.
- `Session.DisplayDir`, `Session.ShortDir`, `Session.CreatedAt`, and `Session.Summary*` fields are pre-computed at load time. Do not call `os.UserHomeDir` in render paths. Use `resolveHome()` (defined in `model.go`) wherever a home directory string is needed outside the render path.
- `SessionStats` (message count, output tokens, context window size) is loaded asynchronously via `loadStatsCmd` in parallel with `loadMessagesCmd` whenever the cursor moves. The model field `stats *SessionStats` is `nil` while loading.
- `renderPreview` computes `headerLines` dynamically from the rendered header string — do not use a hardcoded constant. `renderWorkspaceSessions` follows the same pattern.
- Update this file when adding new files, modes, or architectural patterns.

## Tooling

Run `make check` before committing — it runs `fmt`, `vet`, `lint`, and `test` in sequence.

| Command | What it does |
|---------|-------------|
| `make fmt` | Formats all Go files with `gofmt` and `goimports` |
| `make vet` | Runs `go vet ./...` |
| `make lint` | Runs `golangci-lint run` (requires `golangci-lint` installed) |
| `make test` | Runs `go test ./...` |
| `make check` | Runs all of the above in order |

Lint rules live in `.golangci.yml`. The active linters are `errcheck`, `govet`, `misspell`, `nolintlint`, `revive`, `staticcheck`, and `unused`. Formatters (`gofmt`, `goimports`) are configured separately under the `formatters` key (golangci-lint v2 format). `revive` is configured without the exported-symbol doc-comment rule. `nolintlint` requires every `//nolint` directive to name the specific linter and include a justification comment. To adjust linters or tweak rule severity, edit `.golangci.yml` only.

Install required dev tools:

```sh
brew install golangci-lint
go install golang.org/x/tools/cmd/goimports@latest
```

## CI / CD

Two GitHub Actions workflows live in `.github/workflows/`:

| Workflow | File | Trigger | What it does |
|----------|------|---------|-------------|
| CI | `ci.yml` | Push to `main`, all PRs | Format check, `go vet`, golangci-lint, `go test` |
| Release | `release.yml` | Push a `v*` tag | Vets + tests, cross-compiles for both macOS targets, publishes GitHub Release with binaries and auto-generated notes, and auto-updates the Homebrew tap |

**Cutting a release:**

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds `lazyopencode-<os>-<arch>` binaries for `darwin/amd64` and `darwin/arm64` (pure-Go cross-compilation, `CGO_ENABLED=0`). The `main.version` variable is stamped with the tag name at build time via `-ldflags`. After publishing the GitHub Release, the workflow also commits updated version and sha256 values directly into the formula at `actionscripted/homebrew-lazyopencode` via the GitHub API, using the `TAP_GITHUB_TOKEN` secret.
