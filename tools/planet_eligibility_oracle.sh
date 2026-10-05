#!/usr/bin/env bash
# #91原版鎖定條件矩陣；CPU與堆疊注入完整記錄，非正常玩家入口。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
run="${1:?a|b}"
mode="${2:-minimum}"
[[ "$run" == a || "$run" == b ]] || exit 1
[[ "$mode" == minimum || "$mode" == matrix || "$mode" == matrix2 || "$mode" == pipeline ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/tools/scripts/planet_eligibility_native.go"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
test ! -e "$ROOT/workplace/oracle/planet91-r111-$mode-$run"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/tools/planet_eligibility_native_run.sh "$run" "$mode"
