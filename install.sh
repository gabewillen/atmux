#!/usr/bin/env bash
set -euo pipefail

AMUX_HOME="${AMUX_HOME:-$HOME/.amux}"
AMUX_BIN_DIR="$AMUX_HOME/bin"
AMUX_SRC_DIR="$AMUX_HOME/src/amux"
AMUX_REPO_URL="${AMUX_REPO_URL:-https://github.com/gabrielwillen/amux.git}"
AMUX_REPO_REF="${AMUX_REPO_REF:-main}"

say() {
  printf '%s\n' "$*"
}

ensure_dirs() {
  mkdir -p "$AMUX_HOME" "$AMUX_BIN_DIR" "$AMUX_HOME/agents" "$AMUX_HOME/adapters"
}

install_from_local() {
  local src_root="$1"
  rm -rf "$AMUX_SRC_DIR"
  mkdir -p "$(dirname "$AMUX_SRC_DIR")"

  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete --exclude '.git' "$src_root"/ "$AMUX_SRC_DIR"/
  else
    mkdir -p "$AMUX_SRC_DIR"
    cp -R "$src_root"/. "$AMUX_SRC_DIR"/
    rm -rf "$AMUX_SRC_DIR/.git"
  fi
}

install_from_remote() {
  rm -rf "$AMUX_SRC_DIR"
  mkdir -p "$(dirname "$AMUX_SRC_DIR")"
  git clone --depth 1 --branch "$AMUX_REPO_REF" "$AMUX_REPO_URL" "$AMUX_SRC_DIR" >/dev/null
}

write_launcher() {
  cat > "$AMUX_BIN_DIR/amux" <<'LAUNCHER'
#!/usr/bin/env bash
set -euo pipefail
AMUX_HOME="${AMUX_HOME:-$HOME/.amux}"
exec "$AMUX_HOME/src/amux/bin/amux" "$@"
LAUNCHER
  chmod +x "$AMUX_BIN_DIR/amux"
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
  ensure_dirs

  local script_dir src_root
  script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
  src_root="$script_dir"

  if [[ -x "$src_root/bin/amux" && -d "$src_root/bin/scripts" ]]; then
    say "Installing amux from local checkout: $src_root"
    install_from_local "$src_root"
  else
    say "Installing amux from remote: $AMUX_REPO_URL ($AMUX_REPO_REF)"
    install_from_remote
  fi

  write_launcher
  add_path_hint

  say ""
  say "amux installed"
  say "  home: $AMUX_HOME"
  say "  launcher: $AMUX_BIN_DIR/amux"
  say ""
  say "Try: amux --help"
}

main "$@"
