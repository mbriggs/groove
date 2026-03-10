# groove (gv)

Git worktree manager with integrated tmux sessions and port management.

## Install

```sh
go install github.com/mbriggs/groove/cmd/gv@latest
```

Requires [tmux](https://github.com/tmux/tmux) to be installed and in your PATH.

## Setup

Create a `.groove.yml` at your repo root:

```yaml
ports:
  web: ~        # ~ = assign a free port
  db: 5432      # fixed port
  redis: ~
hooks:
  open: ./scripts/groove_open.sh
  archive: ./scripts/groove_archive.sh
```

Optional global config at `~/.config/groove/config.yml`:

```yaml
worktree_root: ~/groove        # default root for all worktrees
default_remote: origin
shell: zsh                     # shell for tmux sessions
branch_prefix: mbriggs         # prefix for new branches (e.g. mbriggs/my-feature)
```

## Commands

| Command | Description |
|---------|-------------|
| `gv open <branch>` | Create worktree from new branch |
| `gv checkout <branch>` | Create worktree from existing remote branch |
| `gv attach` | Fuzzy-find and switch to a worktree session |
| `gv archive` | Archive current worktree |
| `gv prune` | Clean up worktrees with deleted remote branches |
| `gv gc` | Reconcile state against reality |
| `gv update` | Rebase onto default branch |
| `gv list` | List all worktrees |
| `gv status` | List worktrees for current project |
| `gv sessions up [--all]` | Create missing tmux sessions |
| `gv sessions down [--all]` | Kill tmux sessions |

## Environment Variables

Inside a groove session, these env vars are available:

```
GROOVE_PROJECT        - project name
GROOVE_WORKTREE_ID    - worktree identifier
GROOVE_WORKTREE_PATH  - path to worktree
GROOVE_PROJECT_ROOT   - path to main repo
GROOVE_BRANCH         - branch name
GROOVE_DEFAULT_BRANCH - default branch (e.g. main)
GROOVE_PORT_<NAME>    - assigned port for each configured port
GROOVE_ENV_FILE       - path to .groove-env file
```

Env vars are set on the tmux session via `set-environment`, so new panes and windows inherit them automatically.

For shells that don't pick up tmux environment on their own, add to your `.zshrc` or `.bashrc`:

```sh
[ -f "$GROOVE_ENV_FILE" ] && source "$GROOVE_ENV_FILE"
```

## State

State is stored at `~/.local/share/groove/state.json` (respects `$XDG_DATA_HOME`).
