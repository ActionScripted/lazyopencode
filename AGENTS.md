# lazyoc

lazyoc is a terminal UI (TUI) for managing [opencode](https://opencode.ai) sessions — think lazygit but for opencode. It reads opencode's SQLite database directly (read-only) and lets you browse, search, preview, and open sessions without leaving the terminal.

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
| `model.go` | App state (`model` struct), `Init`/`Update` logic, session filtering |
| `view.go` | `View` function — renders the full TUI layout |
| `update.go` | Key handler helpers (`updateNormal`, `updateSearch`) |
| `keys.go` | Key bindings and `Mode` enum (`ModeNormal`, `ModeSearch`) |
| `session.go` | `Session` and `Message` types; display helpers |
| `db.go` | SQLite queries — `loadSessions` and `loadMessages` |
| `styles.go` | Lip Gloss style definitions |
| `Makefile` | `make install` builds and symlinks to `~/.local/bin/lazyoc` |

## Data flow

1. On startup, `loadSessions` queries `~/.local/share/opencode/opencode.db` (read-only).
2. Sessions are displayed in a list; filtering happens in-memory via `filterSessions`.
3. When a session is selected, `loadMessages` fetches its chat history for the preview pane.
4. Pressing `enter` on a session launches `opencode --session <id>` in the current directory.

## Key conventions

- All DB access is read-only (`?mode=ro`). A missing DB is treated as an empty state, not an error.
- The Bubble Tea pattern is strict: state lives only in `model`, mutations happen in `Update`, rendering is pure in `View`.
- No CGO — the sqlite driver is `modernc.org/sqlite`.
