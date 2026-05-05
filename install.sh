#!/usr/bin/env bash
set -euo pipefail

ATMUX_REPO_URL="${ATMUX_REPO_URL:-https://github.com/gabewillen/atmux.git}"
# Remote clone ref: ATMUX_VERSION is a release tag (becomes v0.2.0), not the semver file at repo root.
# Install identity for this machine is recorded separately in config/install/stamp (date + git short hash).
if [[ -n "${ATMUX_VERSION:-}" ]]; then
  ATMUX_REPO_REF="v${ATMUX_VERSION}"
else
  ATMUX_REPO_REF="${ATMUX_REPO_REF:-main}"
fi
INSTALL_SLASH_COMMANDS=1
INSTALL_SCOPE="${ATMUX_INSTALL_SCOPE:-}"
PROJECT_ROOT="${ATMUX_PROJECT_ROOT:-}"
ATMUX_HOME="${ATMUX_HOME:-}"
ATMUX_BIN_DIR=""
ATMUX_SRC_DIR=""

say() {
  printf '%s\n' "$*"
}

ensure_dirs() {
  mkdir -p "$ATMUX_HOME" "$ATMUX_BIN_DIR" "$ATMUX_BIN_DIR/scripts" "$ATMUX_HOME/agents" "$ATMUX_HOME/adapters" "$ATMUX_HOME/shims"
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

  local commands_root="$HOME"
  if [[ "$INSTALL_SCOPE" == "project" ]]; then
    commands_root="$PROJECT_ROOT"
  fi

  local claude_root="$commands_root/.claude/skills"
  local gemini_root="$commands_root/.gemini/commands"
  local codex_root="$commands_root/.codex/prompts"

  mkdir -p "$claude_root" "$gemini_root" "$codex_root"

  rm -rf "$claude_root/atmux-send" "$claude_root/atmux-assign" "$claude_root/atmux-capture"
  cp -R "$templates_root/claude-code"/. "$claude_root"/

  write_file "$gemini_root/atmux-send.toml" "$(cat "$templates_root/gemini/atmux-send.toml")"
  write_file "$gemini_root/atmux-assign.toml" "$(cat "$templates_root/gemini/atmux-assign.toml")"
  write_file "$gemini_root/atmux-capture.toml" "$(cat "$templates_root/gemini/atmux-capture.toml")"

  write_file "$codex_root/atmux-send.md" "$(cat "$templates_root/codex/atmux-send.md")"
  write_file "$codex_root/atmux-assign.md" "$(cat "$templates_root/codex/atmux-assign.md")"
  write_file "$codex_root/atmux-capture.md" "$(cat "$templates_root/codex/atmux-capture.md")"

  say "Installed slash/custom commands for Claude Code, Gemini CLI, and Codex under $commands_root."
}

install_from_local() {
  local src_root="$1"
  rm -rf "$ATMUX_SRC_DIR"
  mkdir -p "$(dirname "$ATMUX_SRC_DIR")"

  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete \
      --exclude '.git' \
      --exclude '/.atmux' \
      --exclude '/.claude' \
      --exclude '/.codex' \
      --exclude '/.cursor' \
      --exclude '/.gemini' \
      --exclude '/tests' \
      "$src_root"/ "$ATMUX_SRC_DIR"/
  else
    mkdir -p "$ATMUX_SRC_DIR"
    cp -R "$src_root"/. "$ATMUX_SRC_DIR"/
    rm -rf "$ATMUX_SRC_DIR/.git"
    rm -rf "$ATMUX_SRC_DIR/.atmux"
    rm -rf "$ATMUX_SRC_DIR/.claude"
    rm -rf "$ATMUX_SRC_DIR/.codex"
    rm -rf "$ATMUX_SRC_DIR/.cursor"
    rm -rf "$ATMUX_SRC_DIR/.gemini"
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
  if [[ "$INSTALL_SCOPE" == "project" ]]; then
    local quoted_project_root
    printf -v quoted_project_root '%q' "$PROJECT_ROOT"
    cat > "$ATMUX_BIN_DIR/atmux" <<LAUNCHER
#!/usr/bin/env bash
set -euo pipefail
ATMUX_PROJECT_ROOT=$quoted_project_root
ATMUX_HOME="\$ATMUX_PROJECT_ROOT/.atmux"
ATMUX_SOURCE_ROOT="\$ATMUX_HOME/src/atmux"
export ATMUX_HOME ATMUX_SOURCE_ROOT
exec "\$ATMUX_SOURCE_ROOT/bin/atmux" "\$@"
LAUNCHER
  else
    cat > "$ATMUX_BIN_DIR/atmux" <<'LAUNCHER'
#!/usr/bin/env bash
set -euo pipefail
ATMUX_HOME="${ATMUX_HOME:-$HOME/.atmux}"
exec "$ATMUX_HOME/src/atmux/bin/atmux" "$@"
LAUNCHER
  fi
  chmod +x "$ATMUX_BIN_DIR/atmux"
}

install_subcommands() {
  local src_scripts="$ATMUX_SRC_DIR/bin/(atmux)"
  local dst_scripts="$ATMUX_BIN_DIR/scripts"

  rm -rf "$dst_scripts"
  mkdir -p "$dst_scripts"

  if [[ -d "$src_scripts" ]]; then
    cp -R "$src_scripts"/. "$dst_scripts"/
    find "$dst_scripts" -type f -exec chmod +x {} \;
  fi
}

install_shipped_shims() {
  local src_shims="$ATMUX_SRC_DIR/shims"
  local dst_shims="$ATMUX_HOME/shims"
  local src name dst
  [[ -d "$src_shims" ]] || return 0
  mkdir -p "$dst_shims"
  while IFS= read -r src; do
    [[ -d "$src" ]] || continue
    name="$(basename "$src")"
    dst="$dst_shims/$name"
    if [[ -d "$dst/.git" ]]; then
      say "Leaving third-party shim in place: $name"
      continue
    fi
    rm -rf "$dst"
    cp -R "$src" "$dst"
    find "$dst" -type f -perm -u+x -exec chmod +x {} \;
  done < <(find "$src_shims" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)
}

add_path_hint() {
  local export_line='export PATH="$HOME/.atmux/bin:$PATH"'
  local changed=0
  local file

  if [[ "$INSTALL_SCOPE" == "project" ]]; then
    case ":$PATH:" in
      *":$ATMUX_BIN_DIR:"*) ;;
      *)
        say "Project install does not modify shell profiles."
        say "Run from this project with: export PATH=\"$ATMUX_BIN_DIR:\$PATH\""
        ;;
    esac
    return 0
  fi

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

default_project_root() {
  if [[ -n "$PROJECT_ROOT" ]]; then
    (cd "$PROJECT_ROOT" && pwd)
    return 0
  fi

  if git_root="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    printf '%s\n' "$git_root"
  else
    pwd
  fi
}

choose_install_scope() {
  case "$INSTALL_SCOPE" in
    "" ) ;;
    project|system) return 0 ;;
    *)
      echo "invalid install scope: $INSTALL_SCOPE" >&2
      echo "expected: project or system" >&2
      exit 2
      ;;
  esac

  if [[ -t 0 && -t 1 ]]; then
    local answer=""
    printf 'Install atmux for this project or system-wide? [project/system] (project): '
    IFS= read -r answer || true
    case "$answer" in
      ""|p|P|project|PROJECT) INSTALL_SCOPE="project" ;;
      s|S|system|SYSTEM) INSTALL_SCOPE="system" ;;
      *)
        echo "invalid install scope: $answer" >&2
        exit 2
        ;;
    esac
  else
    INSTALL_SCOPE="project"
  fi
}

configure_paths() {
  choose_install_scope
  local requested_home="$ATMUX_HOME"

  if [[ "$INSTALL_SCOPE" == "project" ]]; then
    PROJECT_ROOT="$(default_project_root)"
    ATMUX_HOME="$PROJECT_ROOT/.atmux"
    if [[ -n "$requested_home" && "$requested_home" != "$ATMUX_HOME" ]]; then
      say "Ignoring ATMUX_HOME=$requested_home for project install; using $ATMUX_HOME."
    fi
  else
    if [[ -n "$requested_home" && -f "$requested_home/config/install/scope" && "$(tr -d '[:space:]' < "$requested_home/config/install/scope")" == "project" ]]; then
      say "Ignoring project ATMUX_HOME=$requested_home for system install; using $HOME/.atmux."
      ATMUX_HOME="$HOME/.atmux"
    else
      ATMUX_HOME="${requested_home:-$HOME/.atmux}"
    fi
  fi

  ATMUX_BIN_DIR="$ATMUX_HOME/bin"
  ATMUX_SRC_DIR="$ATMUX_HOME/src/atmux"
}

write_install_metadata() {
  mkdir -p "$ATMUX_HOME/config/install"
  printf '%s\n' "$INSTALL_SCOPE" > "$ATMUX_HOME/config/install/scope"
  if [[ "$INSTALL_SCOPE" == "project" ]]; then
    printf '%s\n' "$PROJECT_ROOT" > "$ATMUX_HOME/config/install/project-root"
  else
    rm -f "$ATMUX_HOME/config/install/project-root"
  fi
}

# Records when/from what commit this tree was installed (UTC date + git short SHA).
# Override with ATMUX_INSTALL_STAMP for reproducible or CI installs.
write_install_stamp_from_source() {
  local src_root="${1:-}"
  mkdir -p "$ATMUX_HOME/config/install"
  local stamp short=""
  if [[ -n "${ATMUX_INSTALL_STAMP:-}" ]]; then
    stamp="$ATMUX_INSTALL_STAMP"
  else
    if [[ -n "$src_root" && -d "$src_root/.git" ]]; then
      short="$(git -C "$src_root" rev-parse --short HEAD 2>/dev/null || true)"
    fi
    if [[ -z "$short" && -d "${ATMUX_SRC_DIR:-}" && -d "$ATMUX_SRC_DIR/.git" ]]; then
      short="$(git -C "$ATMUX_SRC_DIR" rev-parse --short HEAD 2>/dev/null || true)"
    fi
    if [[ -n "$short" ]]; then
      stamp="$(date -u +%Y.%m.%d)-$short"
    else
      stamp="$(date -u +%Y.%m.%d)-unknown"
    fi
  fi
  printf '%s\n' "$stamp" > "$ATMUX_HOME/config/install/stamp"
}

write_project_gitignore() {
  [[ "$INSTALL_SCOPE" == "project" ]] || return 0

  cat > "$ATMUX_HOME/.gitignore" <<'GITIGNORE'
*
!.gitignore
!bin/
!bin/**
!src/
!src/**
!adapters/
!adapters/**
!shims/
!shims/**
!config/
!config/**
config/install/project-root
GITIGNORE
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --no-slash-commands)
        INSTALL_SLASH_COMMANDS=0
        shift
        ;;
      --project)
        INSTALL_SCOPE="project"
        shift
        ;;
      --system)
        INSTALL_SCOPE="system"
        shift
        ;;
      --scope)
        [[ $# -ge 2 ]] || { echo "--scope requires project or system" >&2; exit 2; }
        INSTALL_SCOPE="$2"
        shift 2
        ;;
      --scope=*)
        INSTALL_SCOPE="${1#--scope=}"
        shift
        ;;
      --project-root)
        [[ $# -ge 2 ]] || { echo "--project-root requires a path" >&2; exit 2; }
        PROJECT_ROOT="$2"
        shift 2
        ;;
      --project-root=*)
        PROJECT_ROOT="${1#--project-root=}"
        shift
        ;;
      -h|--help|help)
        cat <<'USAGE'
Usage:
  ./install.sh [--project|--system] [--project-root <dir>] [--no-slash-commands]

Options:
  --project            Install into <project>/.atmux and project-local CLI command dirs.
  --system             Install into ~/.atmux and user-level CLI command dirs.
  --scope <scope>      Same as --project or --system. Values: project, system.
  --project-root <dir> Project directory for --project (default: git root or cwd).
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

  configure_paths
  ensure_dirs
  write_install_metadata
  write_project_gitignore

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

  write_install_stamp_from_source "$src_root"

  write_launcher
  install_subcommands
  install_shipped_shims
  if [[ "$INSTALL_SLASH_COMMANDS" -eq 1 ]]; then
    install_slash_commands
  fi
  add_path_hint

  local installed_version="" install_stamp=""
  if [[ -f "$ATMUX_SRC_DIR/VERSION" ]]; then
    installed_version="$(tr -d '[:space:]' < "$ATMUX_SRC_DIR/VERSION")"
  fi
  if [[ -f "$ATMUX_HOME/config/install/stamp" ]]; then
    install_stamp="$(tr -d '[:space:]' < "$ATMUX_HOME/config/install/stamp")"
  fi

  say ""
  say "atmux installed"
  say "  scope:    $INSTALL_SCOPE"
  [[ "$INSTALL_SCOPE" == "project" ]] && say "  project:  $PROJECT_ROOT"
  say "  home:     $ATMUX_HOME"
  say "  launcher: $ATMUX_BIN_DIR/atmux"
  [[ -n "$installed_version" ]] && say "  version:  $installed_version"
  [[ -n "$install_stamp" ]] && say "  install:  $install_stamp"
  say ""
  say "Try: atmux --help"
}

main "$@"
