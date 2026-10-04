#!/usr/bin/env bash
# #79 子步8-3完整保存診斷。先以GUI容器執行產生器一次；每次輸出拒絕覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b}"
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/workplace/scout79-r83-native-v7.go"
test -f "$ROOT/workplace/oracle/ownership79-r82-entry-a/entry.state"
test ! -e "$ROOT/workplace/scout79-r83-native-v7-$side.json"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/tools/scout_fuel_native_run.sh "/task/workplace/scout79-r83-native-v7-$side.json"
