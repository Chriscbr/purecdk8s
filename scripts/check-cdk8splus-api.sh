#!/usr/bin/env bash
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

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
package="cdk8splus${minor_version}"
version=v2.0.46
pure_dir="$root/$package/v2"

if [[ -n "${UPSTREAM_DIR:-}" ]]; then
	upstream_dir=$UPSTREAM_DIR
	reference="${package} override source"
elif [[ "$minor_version" == "34" ]]; then
	module="github.com/cdk8s-team/cdk8s-plus-go/${package}/v2"
	upstream_dir=$(go mod download -json "${module}@${version}" | sed -nE 's/^[[:space:]]*"Dir": "([^"]+)",/\1/p')
	reference="${module}@${version}"
else
	upstream_dir="$root/cdk8splus34/v2"
	reference="implemented cdk8splus34/v2 API"
fi

if [[ ! -d "$upstream_dir" || ! -d "$pure_dir" ]]; then
  echo "unable to locate cdk8s+ source directories" >&2
  exit 2
fi

UPSTREAM_DIR="$upstream_dir" REFERENCE_LABEL="$reference" go run "$root/scripts/check-cdk8splus-api-shape.go" "$minor_version"
