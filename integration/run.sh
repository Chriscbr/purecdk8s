#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

usage() {
  echo "usage: $0 EXAMPLE [upstream|pure|both]" >&2
  echo "set VERBOSE=1 to stream Docker and application output" >&2
}

run_quietly() {
  local label=$1
  local log=$2
  shift 2

  if [[ "${VERBOSE:-}" == "1" ]]; then
    "$@"
    return
  fi
  if ! "$@" >"$log" 2>&1; then
    echo "❌ $label failed; output follows:" >&2
    cat "$log" >&2
    exit 1
  fi
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
  run_quietly "building $implementation integration image" "$scratch/$implementation-build.log" \
    docker build --quiet --target "$implementation" --tag "$image" --file "$root/integration/Dockerfile" "$root"
  mkdir -p "$scratch/$implementation"
  echo "[$implementation] $case_name"
  run_quietly "running $implementation $case_name" "$scratch/$implementation-run.log" docker run --rm \
    --user "$(id -u):$(id -g)" \
    --volume "$root/integration/examples/$case_name:/examples/$case_name:ro" \
    --volume "$scratch/$implementation:/output" \
    "$image"
done

if [[ "$mode" == check || "$mode" == both ]]; then
  if ! diff -ru "$scratch/upstream" "$scratch/pure"; then
    echo "❌ $case_name: upstream and purecdk8s output differ." >&2
    exit 1
  fi
  echo "✅ $case_name: upstream and purecdk8s output are identical."
fi
