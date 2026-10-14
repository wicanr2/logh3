#!/usr/bin/env bash
# 第136輪原版公開戰術保存／冷啟動載入，不改原版檔案或局面。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?e|f}"
[[ "$side" =~ ^[ef]$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/tools/scripts/battle_load_native.go"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/battle-load-r136-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=10m
cd "$GOLEM"
bash tools/go.sh /task/tools/battle_load_native_run.sh "$side"
