#!/usr/bin/env bash
# 第66輪沿專用dosgolem，不覆寫前輪原版收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?sequence79-r66-<案例>-<a|b>}"
[[ "$name" =~ ^sequence79-r66-[a-z0-9-]+-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/py.sh" /work/tools/scripts/combat_sequence_plan.py "$name"
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
