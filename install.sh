#!/usr/bin/env bash
set -euo pipefail

AMUX_HOME="${AMUX_HOME:-$HOME/.amux}"
AMUX_BIN_DIR="$AMUX_HOME/bin"
AMUX_SRC_DIR="$AMUX_HOME/src/amux"
AMUX_REPO_URL="${AMUX_REPO_URL:-https://github.com/gabrielwillen/amux.git}"
AMUX_REPO_REF="${AMUX_REPO_REF:-main}"
INSTALL_SLASH_COMMANDS=1

say() {
  printf '%s\n' "$*"
}

ensure_dirs() {
  mkdir -p "$AMUX_HOME" "$AMUX_BIN_DIR" "$AMUX_BIN_DIR/scripts" "$AMUX_HOME/agents" "$AMUX_HOME/adapters"
}

write_file() {
  local target="$1"
  local content="$2"
  mkdir -p "$(dirname "$target")"
  printf '%s' "$content" > "$target"
}

install_slash_commands() {
  local templates_root="$AMUX_SRC_DIR/templates/slash-commands"
  [[ -d "$templates_root" ]] || return 0

  local claude_root="$HOME/.claude/skills"
  local gemini_root="$HOME/.gemini/commands"
  local codex_root="$HOME/.codex/prompts"

  mkdir -p "$claude_root" "$gemini_root" "$codex_root"

  rm -rf "$claude_root/amux-send" "$claude_root/amux-assign" "$claude_root/amux-capture"
  cp -R "$templates_root/claude-code"/. "$claude_root"/

  write_file "$gemini_root/amux-send.toml" "$(cat "$templates_root/gemini/amux-send.toml")"
  write_file "$gemini_root/amux-assign.toml" "$(cat "$templates_root/gemini/amux-assign.toml")"
  write_file "$gemini_root/amux-capture.toml" "$(cat "$templates_root/gemini/amux-capture.toml")"

  write_file "$codex_root/amux-send.md" "$(cat "$templates_root/codex/amux-send.md")"
  write_file "$codex_root/amux-assign.md" "$(cat "$templates_root/codex/amux-assign.md")"
  write_file "$codex_root/amux-capture.md" "$(cat "$templates_root/codex/amux-capture.md")"

  say "Installed slash/custom commands for Claude Code, Gemini CLI, and Codex."
}

install_from_local() {
  local src_root="$1"
  rm -rf "$AMUX_SRC_DIR"
  mkdir -p "$(dirname "$AMUX_SRC_DIR")"

  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete --exclude '.git' --exclude '/tests' "$src_root"/ "$AMUX_SRC_DIR"/
  else
    mkdir -p "$AMUX_SRC_DIR"
    cp -R "$src_root"/. "$AMUX_SRC_DIR"/
    rm -rf "$AMUX_SRC_DIR/.git"
    rm -rf "$AMUX_SRC_DIR/tests"
  fi
}

install_from_remote() {
  rm -rf "$AMUX_SRC_DIR"
  mkdir -p "$(dirname "$AMUX_SRC_DIR")"
  git clone --depth 1 --branch "$AMUX_REPO_REF" "$AMUX_REPO_URL" "$AMUX_SRC_DIR" >/dev/null
  rm -rf "$AMUX_SRC_DIR/tests"
}

write_launcher() {
  cat > "$AMUX_BIN_DIR/amux" <<'LAUNCHER'
#!/usr/bin/env bash
set -euo pipefail
AMUX_HOME="${AMUX_HOME:-$HOME/.amux}"
exec "$AMUX_HOME/src/amux/bin/amux.sh" "$@"
LAUNCHER
  chmod +x "$AMUX_BIN_DIR/amux"
}

install_subcommands() {
  local src_scripts="$AMUX_SRC_DIR/bin/amux"
  local dst_scripts="$AMUX_BIN_DIR/scripts"

  mkdir -p "$dst_scripts"
  rm -f "$dst_scripts"/*

  if [[ -d "$src_scripts" ]]; then
    cp -R "$src_scripts"/. "$dst_scripts"/
    find "$dst_scripts" -type f -exec chmod +x {} \;
  fi
}

add_path_hint() {
  local export_line='export PATH="$HOME/.amux/bin:$PATH"'
  local changed=0
  local file

  for file in "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
    [[ -f "$file" ]] || continue
    if ! grep -Fq "$HOME/.amux/bin" "$file"; then
      printf '\n# amux\n%s\n' "$export_line" >> "$file"
      changed=1
    fi
  done

  if [[ "$changed" -eq 1 ]]; then
    say "Updated shell profile(s) with ~/.amux/bin PATH entry."
  fi

  case ":$PATH:" in
    *":$HOME/.amux/bin:"*) ;;
    *)
      say "Current shell PATH does not include ~/.amux/bin."
      say "Run: export PATH=\"$HOME/.amux/bin:\$PATH\""
      ;;
  esac
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --no-slash-commands)
        INSTALL_SLASH_COMMANDS=0
        shift
        ;;
      -h|--help|help)
        cat <<'USAGE'
Usage:
  ./install.sh [--no-slash-commands]

Options:
  --no-slash-commands  Skip installing Claude/Gemini/Codex slash commands.
USAGE
        exit 0
        ;;
      *)
        echo "unknown flag: $1" >&2
        exit 2
        ;;
    esac
  done

  ensure_dirs

  local script_dir src_root
  script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
  src_root="$script_dir"

  if [[ -x "$src_root/bin/amux.sh" && -d "$src_root/bin/amux" ]]; then
    say "Installing amux from local checkout: $src_root"
    install_from_local "$src_root"
  else
    say "Installing amux from remote: $AMUX_REPO_URL ($AMUX_REPO_REF)"
    install_from_remote
  fi

  write_launcher
  install_subcommands
  if [[ "$INSTALL_SLASH_COMMANDS" -eq 1 ]]; then
    install_slash_commands
  fi
  add_path_hint

  say ""
  say "amux installed"
  say "  home: $AMUX_HOME"
  say "  launcher: $AMUX_BIN_DIR/amux"
  say ""
  say "Try: amux --help"
}

main "$@"
