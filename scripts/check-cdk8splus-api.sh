#!/usr/bin/env bash

# Verify that purecdk8s+ exposes an API compatible with cdk8s-plus-go. The
# comparison normalizes purecdk8s' cdk8s and constructs import paths before
# passing the package metadata to apidiff.
set -euo pipefail

if [[ "$#" -ne 1 ]]; then
  echo "usage: $0 <kubernetes-minor-version>" >&2
  echo "example: $0 34" >&2
  exit 2
fi

minor_version=$1
if [[ ! "$minor_version" =~ ^[0-9]+$ ]]; then
  echo "kubernetes-minor-version must be numeric" >&2
  exit 2
fi

repository_root="$(git rev-parse --show-toplevel)"
cd "$repository_root"

readonly pure_dir="$repository_root/cdk8splus${minor_version}/v2"
readonly tool_module="$repository_root/tools/cdk8splus-api-check"

if [[ ! -d "$pure_dir" ]]; then
  printf 'unable to locate %s\n' "$pure_dir" >&2
  exit 2
fi

(
  cd "$tool_module"
  go run . "$minor_version" "$pure_dir"
)
