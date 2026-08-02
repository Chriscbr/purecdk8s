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
readonly tool_dir="$repository_root/tools/cdk8splus-api-check"
readonly upstream_package="github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2"
readonly local_package="github.com/Chriscbr/purecdk8s/cdk8splus${minor_version}/v2"

if [[ ! -d "$pure_dir" ]]; then
  printf 'unable to locate %s\n' "$pure_dir" >&2
  exit 2
fi

(
  cd "$tool_dir"
  # The forward-ported packages continue to reference cdk8splus34 types.
  go run github.com/Chriscbr/purecdk8s/tools/api-compat-check/cmd \
    --name "cdk8splus${minor_version}" \
    --upstream "$upstream_package" \
    --local "$local_package" \
    --source "$pure_dir" \
    --replace 'github.com/Chriscbr/purecdk8s/cdk8splus34/v2=github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2,cdk8splus34' \
    --replace 'github.com/Chriscbr/purecdk8s/cdk8s/v2=github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2,cdk8s' \
    --replace 'github.com/Chriscbr/purecdk8s/constructs/v10=github.com/aws/constructs-go/constructs/v10,constructs' \
    --replace 'github.com/Chriscbr/purecdk8s/jsii=github.com/aws/jsii-runtime-go,jsii'
)
