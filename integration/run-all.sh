#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

for example in "$root/integration/examples"/*; do
  [[ -d "$example" ]] || continue
  "$root/integration/run.sh" "${example##*/}"
done
