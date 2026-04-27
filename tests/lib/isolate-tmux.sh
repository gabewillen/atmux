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

if [[ -z "${ATMUX_TESTS_TMUX_TMPDIR:-}" ]]; then
  ATMUX_TESTS_TMUX_TMPDIR="$(mktemp -d "${TMPDIR:-/tmp}/atmux-test-tmux.XXXXXX")" \
    || { echo "isolate-tmux: failed to mktemp -d" >&2; exit 99; }
  export ATMUX_TESTS_TMUX_TMPDIR
  __atmux_isolate_owner=1
else
  __atmux_isolate_owner=0
fi

export TMUX_TMPDIR="$ATMUX_TESTS_TMUX_TMPDIR"

# Hard-fail if TMUX_TMPDIR is somehow empty after our setup. This should
# never happen, but a refusal is better than silently polluting the
# default socket.
if [[ -z "${TMUX_TMPDIR:-}" ]]; then
  echo "isolate-tmux: TMUX_TMPDIR is empty after isolation; refusing to run" >&2
  exit 99
fi

# Only the test that *created* the isolated tmpdir is responsible for
# tearing it down. If the runner created it, the runner cleans up.
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
      TMUX_TMPDIR="$ATMUX_TESTS_TMUX_TMPDIR" tmux kill-server >/dev/null 2>&1 || true
      rm -rf "$ATMUX_TESTS_TMUX_TMPDIR"
    fi
  ) </dev/null >/dev/null 2>&1 &
  disown 2>/dev/null || true
  unset __atmux_isolate_parent
fi

unset __atmux_isolate_owner
