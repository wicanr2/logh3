#!/usr/bin/env bash
# 第108輪：原版第六步自然續行，局面注入明示；專用dosgolem不改版控檔。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
kind="${1:?baseline|failure|capture|off|click}"
side="${2:?a|b}"
[[ "$kind" =~ ^(baseline|failure|capture|off|click)$ ]] || exit 1
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/tools/scripts/landing_scene_native.go"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
revision=r5
[[ "$kind" != click ]] || revision=r6
test ! -e "$ROOT/workplace/oracle/landing48-r108-$revision-$kind-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/tools/landing_scene_native_run.sh "$kind" "$side"
