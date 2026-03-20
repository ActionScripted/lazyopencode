# lazyoc

lazygit for opencode — a terminal UI for managing opencode sessions.

## Requirements

- [opencode](https://opencode.ai)
- Go 1.21+

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
