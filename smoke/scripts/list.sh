#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
MANIFEST="$ROOT_DIR/smoke/manifest.json"

python3 - "$MANIFEST" <<'PY'
import json, sys
with open(sys.argv[1], "r", encoding="utf-8") as f:
    m = json.load(f)
print("id\tlanguage\tframework\ttier\tstatus")
for c in m.get("cases", []):
    print(f"{c.get('id')}\t{c.get('language')}\t{c.get('framework')}\t{c.get('tier')}\t{c.get('status')}")
PY
