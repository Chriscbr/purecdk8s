#!/usr/bin/env bash
set -euo pipefail

# Symbol names are compared bytewise so sort and comm agree on every locale.
export LC_ALL=C

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

temp_dir=$(mktemp -d "${TMPDIR:-/tmp}/purecdk8s-api-check.XXXXXX")
trap 'rm -f "$temp_dir"/*; rmdir "$temp_dir"' EXIT

collect_types() {
  find "$1" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' -exec \
    sed -nE 's/^type ([A-Z][A-Za-z0-9_]*)[[:space:]].*/\1/p' {} + | sort -u
}

collect_functions() {
  find "$1" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' -exec \
    sed -nE 's/^func ([A-Z][A-Za-z0-9_]*)\(.*/\1/p' {} + | sort -u
}

collect_types "$upstream_dir" > "$temp_dir/upstream-types"
collect_types "$pure_dir" > "$temp_dir/pure-types"
collect_functions "$upstream_dir" > "$temp_dir/upstream-functions"
collect_functions "$pure_dir" > "$temp_dir/pure-functions"

comm -23 "$temp_dir/upstream-types" "$temp_dir/pure-types" > "$temp_dir/missing-types"
comm -23 "$temp_dir/upstream-functions" "$temp_dir/pure-functions" > "$temp_dir/missing-functions"
comm -13 "$temp_dir/upstream-types" "$temp_dir/pure-types" > "$temp_dir/extra-types"
comm -13 "$temp_dir/upstream-functions" "$temp_dir/pure-functions" > "$temp_dir/extra-functions"

missing=0
if [[ -s "$temp_dir/missing-types" ]]; then
  echo "missing exported ${package} types:" >&2
  sed 's/^/  /' "$temp_dir/missing-types" >&2
  missing=1
fi
if [[ -s "$temp_dir/missing-functions" ]]; then
  echo "missing exported ${package} package functions:" >&2
  sed 's/^/  /' "$temp_dir/missing-functions" >&2
  missing=1
fi
if (( missing )); then
  exit 1
fi

UPSTREAM_DIR="$upstream_dir" REFERENCE_LABEL="$reference" go run "$root/scripts/check-cdk8splus-api-shape.go" "$minor_version"

echo "${package} API names cover ${reference} ($(wc -l < "$temp_dir/upstream-types" | tr -d ' ') types, $(wc -l < "$temp_dir/upstream-functions" | tr -d ' ') functions)."
if [[ -s "$temp_dir/extra-types" || -s "$temp_dir/extra-functions" ]]; then
  echo "purecdk8s-only names are allowed and were found:"
  sed 's/^/  type: /' "$temp_dir/extra-types"
  sed 's/^/  function: /' "$temp_dir/extra-functions"
fi
