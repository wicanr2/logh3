#!/usr/bin/env bash
# 第75輪自動命令診斷；沿固定父state，不覆寫歷史收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?command79-r75-<案例>-<a|b>}"
[[ "$name" =~ ^command79-r75-[a-z0-9-]+-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/workplace/oracle/combat79-r63-target-a/selected.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/py.sh" /work/tools/scripts/combat_player_command_plan.py "$name"
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
