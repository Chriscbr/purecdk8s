#!/usr/bin/env bash

# Verify that the purecdk8s core package exposes an API and Go doc comments
# compatible with the original cdk8s core package.
set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"
cd "$repository_root"

readonly source_dir="$repository_root/cdk8s/v2"
readonly tool_dir="$repository_root/tools/cdk8s-api-check"
readonly upstream_package="github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
readonly local_package="github.com/Chriscbr/purecdk8s/cdk8s/v2"

(
  cd "$tool_dir"
  go run github.com/Chriscbr/purecdk8s/tools/api-compat-check/cmd \
    --name cdk8s \
    --upstream "$upstream_package" \
    --local "$local_package" \
    --source "$source_dir" \
    --replace 'github.com/Chriscbr/purecdk8s/constructs/v10=github.com/aws/constructs-go/constructs/v10,constructs' \
    --replace 'github.com/Chriscbr/purecdk8s/jsii=github.com/aws/jsii-runtime-go,jsii'
)
