#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
MANIFEST="$ROOT_DIR/smoke/manifest.json"

CASE="${CASE:-}"
SMOKE_TIERS="${SMOKE_TIERS:-pr}"

list_cases() {
  python3 - "$MANIFEST" "$SMOKE_TIERS" <<'PY'
import json, sys
manifest_path, tiers = sys.argv[1], sys.argv[2]
tier_set = {t.strip() for t in tiers.split(',') if t.strip()}
with open(manifest_path, "r", encoding="utf-8") as f:
    m = json.load(f)
for c in m.get("cases", []):
    if c.get("status") != "active":
        continue
    if tier_set and c.get("tier") not in tier_set:
        continue
    print(c.get("id"))
PY
}

if [ -n "$CASE" ]; then
  "$ROOT_DIR/smoke/scripts/run-case.sh" "$CASE"
  exit 0
fi

found=0
while IFS= read -r case_id; do
  [ -n "$case_id" ] || continue
  found=1
  "$ROOT_DIR/smoke/scripts/run-case.sh" "$case_id"
done <<EOF
$(list_cases)
EOF

if [ "$found" -eq 0 ]; then
  echo "no active smoke cases found for tiers: $SMOKE_TIERS" >&2
  exit 1
fi
