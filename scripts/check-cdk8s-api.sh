#!/usr/bin/env bash

# Verify that the purecdk8s core package exposes an API compatible with the
# original cdk8s core package.
set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"
cd "$repository_root"

readonly source_dir="$repository_root/cdk8s/v2"
readonly tool_module="$repository_root/tools/cdk8s-api-check"

(
  cd "$tool_module"
  go run . "$source_dir"
)
