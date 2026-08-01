#!/usr/bin/env bash
set -euo pipefail

export HOME=/tmp/purecdk8s-integration/home
export GOCACHE=/tmp/purecdk8s-integration/go-cache
export GOPATH=/tmp/purecdk8s-integration/go
export GOMODCACHE="$GOPATH/pkg/mod"
mkdir -p "$HOME" "$GOCACHE" "$GOMODCACHE"

case "$CDK8S_IMPLEMENTATION" in
  upstream)
    command=cdk8s
    ;;
  pure)
    if command -v node >/dev/null || command -v npm >/dev/null; then
      echo "pure image unexpectedly contains Node.js or npm" >&2
      exit 1
    fi
    command=purecdk8s
    ;;
  *)
    echo "unknown CDK8S_IMPLEMENTATION: $CDK8S_IMPLEMENTATION" >&2
    exit 1
    ;;
esac

found=false
for source in /examples/*; do
  [[ -d "$source" ]] || continue
  found=true
  name=${source##*/}
  project="/tmp/purecdk8s-integration/projects/$name"
  mkdir -p "$project" "/output/$name"
  cp -R "$source/." "$project/"

  echo "[$CDK8S_IMPLEMENTATION] $name"
  cd "$project"
  if [[ "$CDK8S_IMPLEMENTATION" == pure ]]; then
    migrate-purecdk8s "$project"
  fi
  "$command" import --no-save
  go mod tidy
  go test ./...
  if [[ "$CDK8S_IMPLEMENTATION" == pure ]] && go list -m all | grep -E 'github.com/aws/(constructs-go/constructs/v10|jsii-runtime-go)|github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2' >/dev/null; then
    echo "$name retained an upstream runtime dependency" >&2
    exit 1
  fi
  "$command" synth
  cp -R "$project/dist/." "/output/$name/"
done

if [[ "$found" == false ]]; then
  echo "no integration examples found" >&2
  exit 1
fi
