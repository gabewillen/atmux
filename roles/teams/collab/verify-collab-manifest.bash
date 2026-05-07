#!/usr/bin/env bash
# Quick guard: default collab MEMBERS must begin with the unified facilitator.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export ATMUX_TEAM="verify-collab-manifest"
export ATMUX_REPO="verify"
# shellcheck disable=SC1090
source "$here/manifest"
[[ "${#MEMBERS[@]}" -ge 2 ]] || {
  printf 'expected >=2 MEMBERS entries, got %d\n' "${#MEMBERS[@]}" >&2
  exit 1
}
first_member="${MEMBERS[0]%%[[:space:]]*}"
[[ -n "$first_member" ]] || {
  printf 'first MEMBER entry has empty first token\n' >&2
  exit 1
}
if [[ "$first_member" != "${ATMUX_TEAM}-facilitator" && "$first_member" != *-facilitator ]]; then
  printf 'first MEMBER token invalid: first_member=%q MEMBERS[0]=%q\n' "$first_member" "${MEMBERS[0]}" >&2
  exit 1
fi
printf 'OK: %d MEMBERS, facilitator first\n' "${#MEMBERS[@]}"
