#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b}"
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/workplace/assault56-r99-native-v3.go"
test -f "$ROOT/workplace/oracle/cleanup79-r88-entry-a/entry.state"
test ! -e "$ROOT/workplace/assault56-r99-v3-$side.json"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/tools/strategic_assault_edge_native_run.sh "/task/workplace/assault56-r99-v3-$side.json"
