#!/usr/bin/env bash
set -euo pipefail

project=${1:?usage: migrate.sh PROJECT}

find "$project" -type f -name '*.go' -exec sed -i \
  -e 's|github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2|github.com/purecdk8s/purecdk8s/cdk8s/v2|g' \
  -e 's|github.com/aws/constructs-go/constructs/v10|github.com/purecdk8s/purecdk8s/constructs/v10|g' \
  -e 's|github.com/aws/jsii-runtime-go|github.com/purecdk8s/purecdk8s/jsii|g' \
  {} +

cd "$project"
go mod edit \
  -droprequire github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2 \
  -droprequire github.com/aws/constructs-go/constructs/v10 \
  -droprequire github.com/aws/jsii-runtime-go \
  -require github.com/purecdk8s/purecdk8s@v0.0.0 \
  -replace github.com/purecdk8s/purecdk8s=/purecdk8s

