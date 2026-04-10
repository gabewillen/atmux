#!/usr/bin/env bash
set -euo pipefail

ATMUX_HOME="${ATMUX_HOME:-$HOME/.atmux}"
ATMUX_BIN_DIR="$ATMUX_HOME/bin"
ATMUX_SRC_DIR="$ATMUX_HOME/src/atmux"
ATMUX_REPO_URL="${ATMUX_REPO_URL:-https://github.com/gabrielwillen/atmux.git}"
# ATMUX_VERSION takes precedence over ATMUX_REPO_REF when set (e.g. ATMUX_VERSION=0.2.0)
if [[ -n "${ATMUX_VERSION:-}" ]]; then
  ATMUX_REPO_REF="v${ATMUX_VERSION}"
else
  ATMUX_REPO_REF="${ATMUX_REPO_REF:-main}"
fi
INSTALL_SLASH_COMMANDS=1

say() {
  printf '%s\n' "$*"
}

ensure_dirs() {
  mkdir -p "$ATMUX_HOME" "$ATMUX_BIN_DIR" "$ATMUX_BIN_DIR/scripts" "$ATMUX_HOME/agents" "$ATMUX_HOME/adapters"
}

write_file() {
  local target="$1"
  local content="$2"
  mkdir -p "$(dirname "$target")"
  printf '%s' "$content" > "$target"
}

install_slash_commands() {
  local templates_root="$ATMUX_SRC_DIR/templates/slash-commands"
  [[ -d "$templates_root" ]] || return 0

  local claude_root="$HOME/.claude/skills"
  local gemini_root="$HOME/.gemini/commands"
  local codex_root="$HOME/.codex/prompts"

  mkdir -p "$claude_root" "$gemini_root" "$codex_root"

  rm -rf "$claude_root/atmux-send" "$claude_root/atmux-assign" "$claude_root/atmux-capture"
  cp -R "$templates_root/claude-code"/. "$claude_root"/

  write_file "$gemini_root/atmux-send.toml" "$(cat "$templates_root/gemini/atmux-send.toml")"
  write_file "$gemini_root/atmux-assign.toml" "$(cat "$templates_root/gemini/atmux-assign.toml")"
  write_file "$gemini_root/atmux-capture.toml" "$(cat "$templates_root/gemini/atmux-capture.toml")"

  write_file "$codex_root/atmux-send.md" "$(cat "$templates_root/codex/atmux-send.md")"
  write_file "$codex_root/atmux-assign.md" "$(cat "$templates_root/codex/atmux-assign.md")"
  write_file "$codex_root/atmux-capture.md" "$(cat "$templates_root/codex/atmux-capture.md")"

  say "Installed slash/custom commands for Claude Code, Gemini CLI, and Codex."
}

install_from_local() {
  local src_root="$1"
  rm -rf "$ATMUX_SRC_DIR"
  mkdir -p "$(dirname "$ATMUX_SRC_DIR")"

  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete --exclude '.git' --exclude '/tests' "$src_root"/ "$ATMUX_SRC_DIR"/
  else
    mkdir -p "$ATMUX_SRC_DIR"
    cp -R "$src_root"/. "$ATMUX_SRC_DIR"/
    rm -rf "$ATMUX_SRC_DIR/.git"
    rm -rf "$ATMUX_SRC_DIR/tests"
  fi
}

install_from_remote() {
  rm -rf "$ATMUX_SRC_DIR"
  mkdir -p "$(dirname "$ATMUX_SRC_DIR")"
  git clone --depth 1 --branch "$ATMUX_REPO_REF" "$ATMUX_REPO_URL" "$ATMUX_SRC_DIR" >/dev/null
  rm -rf "$ATMUX_SRC_DIR/tests"
}

write_launcher() {
  cat > "$ATMUX_BIN_DIR/atmux" <<'LAUNCHER'
#!/usr/bin/env bash
set -euo pipefail
ATMUX_HOME="${ATMUX_HOME:-$HOME/.atmux}"
exec "$ATMUX_HOME/src/atmux/bin/atmux" "$@"
LAUNCHER
  chmod +x "$ATMUX_BIN_DIR/atmux"
}

install_subcommands() {
  local src_scripts="$ATMUX_SRC_DIR/bin/(atmux)"
  local dst_scripts="$ATMUX_BIN_DIR/scripts"

  mkdir -p "$dst_scripts"
  rm -f "$dst_scripts"/*

  if [[ -d "$src_scripts" ]]; then
    cp -R "$src_scripts"/. "$dst_scripts"/
    find "$dst_scripts" -type f -exec chmod +x {} \;
  fi
}

add_path_hint() {
  local export_line='export PATH="$HOME/.atmux/bin:$PATH"'
  local changed=0
  local file

  for file in "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
    [[ -f "$file" ]] || continue
    if ! grep -Fq "$HOME/.atmux/bin" "$file"; then
      printf '\n# atmux\n%s\n' "$export_line" >> "$file"
      changed=1
    fi
  done

  if [[ "$changed" -eq 1 ]]; then
    say "Updated shell profile(s) with ~/.atmux/bin PATH entry."
  fi

  case ":$PATH:" in
    *":$HOME/.atmux/bin:"*) ;;
    *)
      say "Current shell PATH does not include ~/.atmux/bin."
      say "Run: export PATH=\"$HOME/.atmux/bin:\$PATH\""
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

  if [[ -x "$src_root/bin/atmux" && -d "$src_root/bin/(atmux)" ]]; then
    say "Installing atmux from local checkout: $src_root"
    install_from_local "$src_root"
  else
    say "Installing atmux from remote: $ATMUX_REPO_URL ($ATMUX_REPO_REF)"
    install_from_remote
  fi

  write_launcher
  install_subcommands
  if [[ "$INSTALL_SLASH_COMMANDS" -eq 1 ]]; then
    install_slash_commands
  fi
  add_path_hint

  local installed_version=""
  if [[ -f "$ATMUX_SRC_DIR/VERSION" ]]; then
    installed_version="$(tr -d '[:space:]' < "$ATMUX_SRC_DIR/VERSION")"
  fi

  say ""
  say "atmux installed"
  say "  home:     $ATMUX_HOME"
  say "  launcher: $ATMUX_BIN_DIR/atmux"
  [[ -n "$installed_version" ]] && say "  version:  $installed_version"
  say ""
  say "Try: atmux --help"
}

main "$@"
