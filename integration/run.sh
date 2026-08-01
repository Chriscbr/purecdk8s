#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

usage() {
  echo "usage: $0 EXAMPLE [upstream|pure|both]" >&2
}

case_name=${1:-}
mode=${2:-check}
if [[ -z "$case_name" || "$#" -gt 2 || "$case_name" == . || "$case_name" == .. || "$case_name" == */* || ! -d "$root/integration/examples/$case_name" ]]; then
  usage
  exit 2
fi

case "$mode" in
  check)
    implementations=(upstream pure)
    keep_output=false
    ;;
  upstream|pure)
    implementations=("$mode")
    keep_output=true
    ;;
  both)
    implementations=(upstream pure)
    keep_output=true
    ;;
  *)
    usage
    exit 2
    ;;
esac

scratch=$(mktemp -d "${TMPDIR:-/tmp}/purecdk8s-${case_name}.XXXXXX")
if [[ "$keep_output" == false ]]; then
  trap 'rm -rf "$scratch"' EXIT
else
  echo "Saving output to: $scratch"
fi

for implementation in "${implementations[@]}"; do
  image="purecdk8s-integration-$implementation"
  docker build --target "$implementation" --tag "$image" --file "$root/integration/Dockerfile" "$root"
  mkdir -p "$scratch/$implementation"
  docker run --rm \
    --user "$(id -u):$(id -g)" \
    --volume "$root/integration/examples/$case_name:/examples/$case_name:ro" \
    --volume "$scratch/$implementation:/output" \
    "$image"
done

if [[ "$mode" == check || "$mode" == both ]]; then
  diff -ru "$scratch/upstream" "$scratch/pure"
  echo "$case_name: upstream and purecdk8s output are identical."
fi
