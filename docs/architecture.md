# atmux Architecture

## Overview

`atmux` is a bash-based tmux orchestration layer for AI agents. It creates repo-scoped tmux sessions, starts adapter CLIs inside those sessions, routes notifications through panes, and stores coordination state on disk under `ATMUX_HOME`.

The default install is project-local:

```text
<project>/.atmux/
```

System installs use:

```text
~/.atmux/
```

## Major Components

### CLI Router

`bin/atmux` is the top-level command. Subcommands are implemented as executable scripts under `bin/(atmux)/`.

Examples:

- `bin/(atmux)/session`
- `bin/(atmux)/send`
- `bin/(atmux)/notify`
- `bin/(atmux)/exec`
- `bin/(atmux)/watch`
- `bin/(atmux)/kill`

Installed project or system launchers execute the installed source copy under `<ATMUX_HOME>/src/atmux`.

### Session And Worktree Layer

Agent sessions use this naming convention:

```text
atmux-<repo>-<agent>
```

Default worktrees live at:

```text
<ATMUX_HOME>/agents/<repo>-<agent>
```

When a worktree is created, submodules are initialized recursively with:

```sh
git submodule update --init --recursive
```

### Adapter Layer

Adapters live under `adapter/<name>/` and provide a `cmd` entrypoint plus scripts for:

- `start`
- `status`
- `model`
- `reasoning-level`
- `control-file`

Built-in adapters:

- `auto`
- `claude-code`
- `codex`
- `cursor-agent`
- `gemini`

Each adapter owns its CLI-specific startup arguments, status parsing, model validation, and input submit keys.

### Intelligence Mapping

The user-facing `--intelligence 0-100` scale is translated by each adapter's `manifest.json`.

For Gemini, low intelligence intentionally selects the lite model:

| Intelligence | Gemini Model | Reasoning |
|--------------|--------------|-----------|
| 0-39 | `gemini-3.1-flash-lite-preview` | `low` |
| 40-74 | `gemini-3-flash-preview` | `medium` |
| 75-89 | `gemini-3.1-pro-preview` | `medium` |
| 90-100 | `gemini-3.1-pro-preview` | `high` |

### Notification Queue

Notifications are queued per target pane under:

```text
<ATMUX_HOME>/notify-queue/<pane-key>/
```

Delivery is serialized by a background worker per pane. The worker:

- Watches a FIFO wake file for new payloads.
- Preserves existing prompt input where possible.
- Uses adapter prompt/status hints to avoid submitting into busy panes.
- Handles large pasted content by waiting for prompt input to settle before submitting.

### Exec And Watch State

`atmux exec` stores process metadata under:

```text
<ATMUX_HOME>/exec/<repo>/<pid>/
```

`atmux watch --pid` registers watcher panes under that exec directory. `atmux kill --pid` signals the tracked process, waits for executor notifications and watcher fan-out, then removes the metadata directory.

### Issues And Messages

Filesystem-backed coordination state lives under:

```text
<ATMUX_HOME>/issues/<repo>/
<ATMUX_HOME>/messages/<repo>/
```

`atmux assign`, `atmux comment`, and `atmux send` write state there and notify the relevant panes.

## Release Flow

Releases are tag-driven. A release tag must match `VERSION` exactly:

```text
VERSION=<version>
tag=v<version>
```

The GitHub Release workflow validates the match, archives the tagged source, and uploads `atmux-<version>.tar.gz`.
