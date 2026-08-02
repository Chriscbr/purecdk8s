#!/usr/bin/env bash

# Verify that the purecdk8s constructs package has exactly the API and Go doc
# comments exposed by the original constructs-go package.
set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"
cd "$repository_root"

readonly upstream_package="github.com/aws/constructs-go/constructs/v10"
readonly local_package="github.com/Chriscbr/purecdk8s/constructs/v10"
readonly source_dir="$repository_root/constructs/v10"
readonly tool_dir="$repository_root/tools/constructs-api-check"

(
  cd "$tool_dir"
  go run github.com/Chriscbr/purecdk8s/tools/api-compat-check/cmd \
    --name constructs \
    --upstream "$upstream_package" \
    --local "$local_package" \
    --source "$source_dir"
)
