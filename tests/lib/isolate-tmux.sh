#!/usr/bin/env bash
# Source-only library. Guarantees that any tmux command run by this test
# (or any child of this test) targets a private socket — never the
# developer's default tmux server.
#
# Source this near the top of every test script:
#
#   source "$(dirname "${BASH_SOURCE[0]}")/lib/isolate-tmux.sh"
#
# Idempotent: if the runner has already isolated the socket (signalled
# via ATMUX_TESTS_TMUX_TMPDIR), this is a no-op.
#
# A test that bypasses this lib and runs `tmux` directly against the
# default socket is a bug — see feedback memory feedback_tmux_isolation.

# Refuse to run as a script — must be sourced so the trap and exports
# affect the test process itself.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  echo "isolate-tmux.sh must be sourced, not executed" >&2
  exit 2
fi

__atmux_isolate_repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." 2>/dev/null && pwd -P || true)"
if [[ -n "$__atmux_isolate_repo_root" ]]; then
  export ATMUX_SOURCE_ROOT="$__atmux_isolate_repo_root"
fi

# Tests must not inherit the invoking agent's identity. Individual tests set
# these explicitly when they need agent-mode behavior.
unset ATMUX_AGENT_NAME ATMUX_SESSION_NAME ATMUX_TEAM ATMUX_ROLE

__atmux_isolate_real_tmux() {
  local repo_root="" shim_path="" candidate resolved
  repo_root="$__atmux_isolate_repo_root"
  if [[ -n "$repo_root" && -x "$repo_root/bin/tmux" ]]; then
    shim_path="$repo_root/bin/tmux"
  fi

  while IFS= read -r candidate; do
    [[ -n "$candidate" && -x "$candidate" ]] || continue
    resolved="$(cd -- "$(dirname -- "$candidate")" 2>/dev/null && pwd -P || true)"
    [[ -n "$resolved" ]] || continue
    resolved="$resolved/$(basename -- "$candidate")"
    [[ -n "$shim_path" && "$resolved" == "$shim_path" ]] && continue
    [[ -x "$(dirname -- "$resolved")/atmux" && -d "$(dirname -- "$resolved")/(atmux)" ]] && continue
    printf '%s\n' "$resolved"
    return 0
  done < <(
    type -P tmux 2>/dev/null || true
    printf '%s\n' /opt/homebrew/bin/tmux /usr/local/bin/tmux /usr/bin/tmux /bin/tmux
  )

  return 1
}

if [[ -z "${ATMUX_TESTS_TMUX_TMPDIR:-}" ]]; then
  ATMUX_TESTS_TMUX_TMPDIR="$(mktemp -d "${TMPDIR:-/tmp}/atmux-test-tmux.XXXXXX")" \
    || { echo "isolate-tmux: failed to mktemp -d" >&2; exit 99; }
  export ATMUX_TESTS_TMUX_TMPDIR
  __atmux_isolate_owner=1
else
  __atmux_isolate_owner=0
fi

ATMUX_REAL_TMUX="$(__atmux_isolate_real_tmux || true)"
if [[ -z "$ATMUX_REAL_TMUX" || ! -x "$ATMUX_REAL_TMUX" ]]; then
  echo "isolate-tmux: real tmux binary is unavailable; refusing to run" >&2
  exit 99
fi
export ATMUX_REAL_TMUX
export ATMUX_ALLOW_REAL_TMUX=1

export TMUX_TMPDIR="$ATMUX_TESTS_TMUX_TMPDIR"
export ATMUX_TMUX_SOCKET="$ATMUX_TESTS_TMUX_TMPDIR/tmux-$(id -u)/default"
# Never let an inherited live tmux client socket override the private test
# socket. This is the critical guard for running atmux tests from inside atmux.
export TMUX=""
mkdir -p "$(dirname -- "$ATMUX_TMUX_SOCKET")"
chmod 700 "$(dirname -- "$ATMUX_TMUX_SOCKET")"

# Hard-fail if TMUX_TMPDIR is somehow empty after our setup. This should
# never happen, but a refusal is better than silently polluting the
# default socket.
if [[ -z "${TMUX_TMPDIR:-}" ]]; then
  echo "isolate-tmux: TMUX_TMPDIR is empty after isolation; refusing to run" >&2
  exit 99
fi
if [[ -z "${ATMUX_TMUX_SOCKET:-}" || "$ATMUX_TMUX_SOCKET" != "$ATMUX_TESTS_TMUX_TMPDIR"/* ]]; then
  echo "isolate-tmux: ATMUX_TMUX_SOCKET is not under the isolated tmpdir; refusing to run" >&2
  exit 99
fi

# Only the test that *created* the isolated tmpdir is responsible for
# tearing it down. If the runner created it, the runner process owns cleanup.
#
# Tests routinely set their own `trap ... EXIT` after sourcing this lib,
# which would silently replace any trap we register here. To survive
# that, spawn a detached watcher tied to the test's PID: it sleeps until
# the test process disappears, then kills the isolated tmux server and
# removes the tmpdir. This guarantees cleanup independent of the test's
# trap discipline.
if [[ "${__atmux_isolate_owner:-0}" == "1" ]]; then
  __atmux_isolate_parent=$$
  (
    # shellcheck disable=SC2034
    parent="$__atmux_isolate_parent"
    while kill -0 "$parent" 2>/dev/null; do
      sleep 1
    done
    if [[ -n "${ATMUX_TESTS_TMUX_TMPDIR:-}" && -d "$ATMUX_TESTS_TMUX_TMPDIR" ]]; then
      TMUX= TMUX_TMPDIR="$ATMUX_TESTS_TMUX_TMPDIR" ATMUX_TMUX_SOCKET="$ATMUX_TMUX_SOCKET" ATMUX_REAL_TMUX="$ATMUX_REAL_TMUX" ATMUX_ALLOW_REAL_TMUX=1 \
        tmux kill-server >/dev/null 2>&1 || true
      rm -rf "$ATMUX_TESTS_TMUX_TMPDIR"
    fi
  ) </dev/null >/dev/null 2>&1 &
  disown 2>/dev/null || true
  unset __atmux_isolate_parent
fi

unset __atmux_isolate_owner
unset -f __atmux_isolate_real_tmux
