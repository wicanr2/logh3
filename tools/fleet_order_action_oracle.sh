#!/usr/bin/env bash
# 第93輪旗艦行動碼診斷；使用既有dosgolem副本，不覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b}"
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/workplace/order79-r93-native.go"
test -f "$ROOT/workplace/order79-r93-native-run.sh"
test ! -e "$ROOT/workplace/order79-r93-$side.json"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/workplace/order79-r93-native-run.sh "/task/workplace/order79-r93-$side.json"
