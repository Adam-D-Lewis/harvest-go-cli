# Harvest CLI

A command-line tool for [Harvest](https://www.getharvest.com/) time tracking.

## Installation

```bash
go build -o harvest
sudo cp harvest /usr/local/bin/
```

## Configuration

Create `~/.config/harvest/.env`:

```bash
HARVEST_TOKEN=your_token_here
HARVEST_ACCOUNT_ID=your_account_id
```

Get your token from: https://id.getharvest.com/developers

## Usage

```bash
# Start a timer (interactive)
harvest start

# Start with fuzzy matching
harvest start "project" "task"
harvest start "project" "task" "notes"

# Stop timer
harvest stop

# Log time entry
harvest log
harvest log "project" "task" 2.5 "notes"

# View entries
harvest view today
harvest view week

# List projects/tasks
harvest ls projects
harvest ls tasks -p <project_id>

# Status (for tmux)
harvest status
```

## Shell Completion

```bash
# Zsh
mkdir -p ~/.zsh/completions
harvest completion zsh > ~/.zsh/completions/_harvest

# Add to ~/.zshrc:
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit

# Bash
harvest completion bash > /etc/bash_completion.d/harvest
```

## tmux Integration

Add to `~/.tmux.conf`:

```bash
set -g status-right "#(harvest status) | %H:%M"
set -g status-interval 60
```
