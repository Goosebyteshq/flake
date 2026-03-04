#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "usage: $0 <case-id>" >&2
  exit 2
fi

CASE_ID="$1"
ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
MANIFEST="$ROOT_DIR/smoke/manifest.json"

read_case_json() {
  python3 - "$MANIFEST" "$CASE_ID" <<'PY'
import json, sys
manifest_path, case_id = sys.argv[1], sys.argv[2]
with open(manifest_path, "r", encoding="utf-8") as f:
    m = json.load(f)
for c in m.get("cases", []):
    if c.get("id") == case_id:
        print(json.dumps(c))
        sys.exit(0)
print(f"case id not found: {case_id}", file=sys.stderr)
sys.exit(1)
PY
}

CASE_JSON="$(read_case_json)"

case_field() {
  local key="$1"
  python3 - "$CASE_JSON" "$key" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
key = sys.argv[2]
val = obj
for part in key.split('.'):
    val = val[part]
print(val)
PY
}

CASE_STATUS="$(case_field status)"
if [ "$CASE_STATUS" != "active" ]; then
  echo "skip $CASE_ID: status=$CASE_STATUS"
  exit 0
fi

CASE_PATH_REL="$(case_field path)"
CASE_PATH="$ROOT_DIR/$CASE_PATH_REL"
if [ ! -d "$CASE_PATH" ]; then
  echo "missing case path: $CASE_PATH" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
  if [ "${SMOKE_KEEP:-}" = "1" ]; then
    echo "smoke artifacts kept at $TMP_DIR"
    return
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

RAW_OUT="$TMP_DIR/raw.out"
SCAN_OUT="$TMP_DIR/scan.out"
STATE_PATH="$TMP_DIR/state.json"

IMAGE_TAG="flake-smoke-${CASE_ID}"

FLAKE_BIN="${FLAKE_BIN:-$ROOT_DIR/.dist-smoke/flake}"
if [ ! -x "$FLAKE_BIN" ]; then
  echo "flake binary not found at $FLAKE_BIN" >&2
  echo "build it with: make smoke-build" >&2
  exit 1
fi

echo "[smoke] build image: $CASE_ID"
docker build -t "$IMAGE_TAG" "$CASE_PATH" >/dev/null

echo "[smoke] run case: $CASE_ID"
docker run --rm "$IMAGE_TAG" >"$RAW_OUT" 2>&1 || true

"$FLAKE_BIN" scan \
  --framework auto \
  --state "$STATE_PATH" \
  --input "$RAW_OUT" \
  --json >"$SCAN_OUT"

python3 - "$CASE_JSON" "$SCAN_OUT" <<'PY'
import json, sys
case = json.loads(sys.argv[1])
scan_path = sys.argv[2]
with open(scan_path, "r", encoding="utf-8") as f:
    out = f.read()
start = out.find('{')
if start < 0:
    print("missing JSON payload in scan output", file=sys.stderr)
    sys.exit(1)
data = json.loads(out[start:])
exp = case["expected"]
if data.get("framework") != exp.get("framework"):
    print(f"framework mismatch: got={data.get('framework')} want={exp.get('framework')}", file=sys.stderr)
    sys.exit(1)
want_counts = exp.get("class_counts", {})
got_counts = data.get("class_counts", {})
for k, v in want_counts.items():
    if got_counts.get(k) != v:
        print(f"class_counts mismatch for {k}: got={got_counts.get(k)} want={v}", file=sys.stderr)
        sys.exit(1)
print(f"[smoke] ok {case['id']} framework={data.get('framework')} class_counts={got_counts}")
PY
