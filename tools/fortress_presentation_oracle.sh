#!/usr/bin/env bash
# 第133輪：沿完整S3保存狀態在第五步明示配置要塞，不改dosgolem版控檔。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
kind="${1:?iserlohn|empire}"
side="${2:?a|b}"
[[ "$kind" =~ ^(iserlohn|empire)$ && "$side" =~ ^(a|b)$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/tools/scripts/fortress_presentation_native.go"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
test ! -e "$ROOT/workplace/oracle/fortress48-r134-r3-$kind-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=8m
cd "$GOLEM"
bash tools/go.sh /task/tools/fortress_presentation_native_run.sh "$kind" "$side"
