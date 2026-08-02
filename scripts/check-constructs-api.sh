#!/usr/bin/env bash

# Verify that the purecdk8s constructs package has exactly the API exposed by the
# original constructs-go package.
set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"
cd "$repository_root"

readonly upstream_package="github.com/aws/constructs-go/constructs/v10"
readonly local_package="github.com/purecdk8s/purecdk8s/constructs/v10"
readonly tool_module="$repository_root/tools/api-check"

report="$(mktemp "${TMPDIR:-/tmp}/constructs-api.XXXXXX")"
trap 'rm -f "$report"' EXIT

(
  cd "$tool_module"
  go run golang.org/x/exp/cmd/apidiff "$upstream_package" "$local_package"
) >"$report"

if [[ -s "$report" ]]; then
  printf 'constructs API differs from %s:\n' "$upstream_package" >&2
  cat "$report" >&2
  exit 1
fi

printf 'constructs API matches %s.\n' "$upstream_package"
