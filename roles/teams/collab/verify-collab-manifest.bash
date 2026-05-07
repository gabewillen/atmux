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
[[ "${MEMBERS[0]}" == *facilitator* ]] || {
  printf 'first MEMBER must be facilitator, got %q\n' "${MEMBERS[0]}" >&2
  exit 1
}
printf 'OK: %d MEMBERS, facilitator first\n' "${#MEMBERS[@]}"
