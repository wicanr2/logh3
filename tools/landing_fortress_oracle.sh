#!/usr/bin/env bash
# 第135輪要塞型行星登陸；注入診斷，不當正常玩家收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
variant="${1:?fortress|iserlohn}"
kind="${2:?failure|capture}"
side="${3:?a|b}"
[[ "$variant" =~ ^(fortress|iserlohn)$ && "$kind" =~ ^(failure|capture)$ && "$side" =~ ^(a|b)$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/tools/scripts/landing_fortress_native.go"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
test ! -e "$ROOT/workplace/oracle/landing48-r135-$variant-$kind-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/tools/landing_fortress_native_run.sh "$variant" "$kind" "$side"
