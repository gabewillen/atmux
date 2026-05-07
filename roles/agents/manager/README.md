# Manager

The `manager` role is a delegation-only engineering lead for a repository.
It subscribes to issue and pull-request feeds, triages incoming work,
creates agents or teams to implement and review changes, and keeps model
usage aligned with task complexity.

## Defaults

- Kind: `agent`
- Default adapters: `codex`, `claude-code`, `cursor-agent`, `opencode`, `gemini`
- Default intelligence: `90`

## Usage

```sh
atmux agent create manager --role manager
```

The role expects the GitHub CLI (`gh`) to be authenticated and available
when issue/PR feed subscription is desired. If `gh` is unavailable, the
manager still runs with prompt-only behavior.

The manager should not edit code directly. It delegates implementation,
test, review, and research work to other agents or teams.
