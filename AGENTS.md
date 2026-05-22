## Cursor Cloud specific instructions

This is a pure-bash project with no build step. The only system dependencies are `bash`, `git`, `tmux`, and `shellcheck` (for linting).

### Running the CLI

Add the repo `bin/` directory to PATH:

```sh
export PATH="$PWD/bin:$PATH"
export ATMUX_ALLOW_OUTSIDE_TMUX=1
```

The `ATMUX_ALLOW_OUTSIDE_TMUX=1` env var is required when running atmux commands from outside a tmux session (e.g., in CI or from a cloud agent shell).

### Lint

```sh
find bin install.sh -type f -not -path '*/.git/*' | xargs shellcheck --severity=error -e SC1091
```

### Tests

```sh
ATMUX_ALLOW_OUTSIDE_TMUX=1 ./tests/runner
```

Tests use an isolated tmux server (via `TMUX_TMPDIR`) so they never touch real tmux sessions. Sharding is available via `--shard N/TOTAL` for parallel execution. Live adapter tests (suffix `_live`) are skipped unless `ATMUX_RUN_LIVE_ADAPTER_TESTS=1` is set.

### Key gotchas

- The test runner uses `pkill -f` internally to clean up between tests; do not run tests while other atmux processes are active in the same user context.
- Tests create and destroy tmux sessions rapidly; if a test hangs, check for orphan tmux sessions with `tmux list-sessions` (scoped to the isolated TMUX_TMPDIR).
- The `.tool-versions` file declares `golang 1.25.6` but no Go source exists; it can be ignored for development purposes.
