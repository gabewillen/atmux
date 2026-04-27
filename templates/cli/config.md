```sh
atmux config get   <key>          [--global]
atmux config set   <key> <value>  [--global]
atmux config unset <key>          [--global]
atmux config list                 [--global]
```

One file per key under `<repo>/.atmux/config/<key>` (local) or `~/.atmux/config/<key>` (with `--global`). Key paths are hierarchical with `/`, e.g. `update/auto`. `get` without `--global` resolves local first, then global.
