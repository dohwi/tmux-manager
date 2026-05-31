# tmux-manager

TUI tmux session manager built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Browse** all tmux sessions in a terminal interface
- **Create** new detached sessions (defaults to current directory name)
- **Rename** existing sessions inline
- **Delete** sessions with confirmation
- **Restore** sessions from YAML config files (`~/.config/tmux-manager/sessions/`)

## Installation

```bash
go install github.com/dohwi/tmux-manager@latest
```

Or clone and build:

```bash
git clone https://github.com/dohwi/tmux-manager.git
cd tmux-manager
go build -o tmux-manager
```

## Usage

```
tmux-manager           # Launch TUI session browser
tmux-manager restore   # Restore all sessions from config
```

### Keybindings

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate sessions |
| `Enter` | Attach to selected session |
| `Ctrl+N` | Create new session |
| `Ctrl+R` | Rename selected session |
| `Ctrl+D` | Delete selected session |
| `Ctrl+C` | Quit |

### Session Config (YAML)

Place YAML files in `~/.config/tmux-manager/sessions/`:

```yaml
# ~/.config/tmux-manager/sessions/dev.yaml
directory: ~/projects/myapp
windows:
  - name: editor
    command: nvim
  - name: server
    command: npm run dev
  - name: logs
    command: tail -f /var/log/app.log
```

Run `tmux-manager restore` to recreate all configured sessions.
