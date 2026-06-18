# AGENTS.md

## Project: tmux-manager

TUI session manager for tmux. Go + Bubble Tea.

### Stack

- **Language**: Go 1.24+
- **TUI**: Bubble Tea (charmbracelet/bubbletea)
- **Theme**: Catppuccin Mocha
- **CI**: GitHub Actions (lint + test + GoReleaser)
- **Install**: `curl -fsSL ... | bash` one-liner

### Key Paths

| Path | Purpose |
|------|---------|
| `cmd/tmux-manager/main.go` | Entry point |
| `internal/cli/` | CLI commands (`tm`, `tm restore`, `tm setup`, `tm update`) |
| `internal/config/` | YAML config parsing & validation |
| `internal/tmux/` | tmux client, hooks, types |
| `internal/tui/` | Bubble Tea TUI model & styles |
| `internal/update/` | Self-update mechanism |
| `internal/setup/` | Shell integration & tmux.conf setup |
| `install.sh` | One-command install script |

### Development

```bash
go build -o tm ./cmd/tmux-manager
go test -race ./...
golangci-lint run
```

### GitHub Workflow

Follow [github-flow](/.agents/skills/github-flow/SKILL.md). Branch → `--no-ff` merge → push. No PRs.