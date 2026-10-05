#!/usr/bin/env bash
# 第109輪三個原版分支；明示注入，不當正常玩家收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
kind="${1:?counterfirst|empire|multiple}"
side="${2:?a|b}"
revision="${3:-r1}"
[[ "$kind" =~ ^(counterfirst|empire|multiple)$ ]] || exit 1
[[ "$side" == a || "$side" == b ]] || exit 1
[[ "$revision" == r1 || ( "$revision" == r2 && "$kind" == multiple ) ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/tools/scripts/landing_frame_native.go"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
test ! -e "$ROOT/workplace/oracle/landing48-r109-branch-$revision-$kind-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$ROOT:/task"
export DOSGOLEM_GO_CMD=bash DOSGOLEM_TIMEOUT=5m
cd "$GOLEM"
bash tools/go.sh /task/tools/landing_frame_branch_native_run.sh "$kind" "$side" "$revision"
