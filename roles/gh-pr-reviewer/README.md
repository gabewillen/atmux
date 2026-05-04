# GitHub PR Reviewer

The `gh-pr-reviewer` role reviews GitHub pull requests, looks for
concrete risks in the diff, and posts structured review feedback with
`gh`.

## Defaults

- Kind: `agent`
- Default adapters: `claude-code`, `codex`
- Default intelligence: `75`

## Usage

```sh
atmux agent create reviewer --role gh-pr-reviewer
atmux send --to reviewer "Review PR <number>"
```

The role expects the GitHub CLI (`gh`) to be authenticated and available
in the agent environment. It is designed for review work, not for making
code changes directly.

See [role.md](./role.md) for the role prompt.
