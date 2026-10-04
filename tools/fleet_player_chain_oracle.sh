#!/usr/bin/env bash
# 第76輪正式玩家鏈所需的名冊診斷，固定執行器，拒絕覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?fleet79-r76-order-<base|reverse|rotate|interleave|number>-<a|b>}"
[[ "$name" =~ ^fleet79-r76-order-(base|reverse|rotate|interleave|number)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$ROOT/workplace/oracle/combat79-r62-natural-a/entry.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
bash "$ROOT/tools/py.sh" /work/tools/scripts/fleet_player_chain_plan.py "$name"
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
