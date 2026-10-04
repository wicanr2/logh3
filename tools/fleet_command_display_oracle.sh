#!/usr/bin/env bash
# 第77輪#48／#79命令顯示；凍結父狀態，拒絕覆寫舊收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?combat79-r77-<wait|hover1..3|popup1..3>-<a|b>}"
[[ "$name" =~ ^combat79-r77-(wait|hover[123]|popup[123])-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/workplace/oracle/combat79-r62-command-a/command.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/py.sh" /work/tools/scripts/fleet_command_display_plan.py "$name"
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
