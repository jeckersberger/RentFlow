#!/bin/bash
# Static audit for migration file naming/version collisions.

set -euo pipefail

ROOT="${1:-services}"
errors=0
services=0
migrations=0

if [[ ! -d "$ROOT" ]]; then
  echo "ERROR: services root not found: $ROOT" >&2
  exit 2
fi

while IFS= read -r -d '' dir; do
  svc="$(basename "$(dirname "$dir")")"
  ((services += 1))
  declare -A seen=()

  while IFS= read -r -d '' file; do
    base="$(basename "$file")"
    ((migrations += 1))

    if [[ ! "$base" =~ ^([0-9]+)_[A-Za-z0-9._-]+\.sql$ ]]; then
      echo "ERROR: $svc has invalid migration filename: $base" >&2
      ((errors += 1))
      continue
    fi

    version="${BASH_REMATCH[1]}"
    normalized="$(printf '%d' "$((10#$version))")"
    if [[ -n "${seen[$normalized]:-}" ]]; then
      echo "ERROR: $svc has duplicate migration version $version: ${seen[$normalized]} and $base" >&2
      ((errors += 1))
    else
      seen[$normalized]="$base"
    fi
  done < <(find "$dir" -maxdepth 1 -type f -name '*.sql' ! -name '*.down.sql' -print0 | sort -z)

done < <(find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name migrations -print0 | sort -z)

echo "Migration audit: $migrations migrations across $services service directories"

if (( errors > 0 )); then
  echo "Migration audit FAILED with $errors error(s)" >&2
  exit 1
fi

echo "Migration audit passed"
