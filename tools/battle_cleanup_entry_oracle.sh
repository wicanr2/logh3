#!/usr/bin/env bash
# #79：固定第五回合自然到8-1入口；不覆寫歷史父保存。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
side="${1:?a|b}"
[[ "$side" == a || "$side" == b ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
test ! -e "$ROOT/workplace/oracle/cleanup79-r88-entry-$side"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "cleanup79-r88-entry-$side" \
  -load-state /src/workplace/fcmd/eng-t5.state \
  -steps 1187954189 -tick 20000 \
  -clicks 1140000000:240:200 -click-hold 3000000 \
  -poke lin:5E00C@1187900000=78,56,34,12 \
  -regs-from 1187900000 -regs-max 40 \
  -regs-at 181B:03F0,185C:001D,185C:02DA \
  -save-state "1187954188:/out/oracle/cleanup79-r88-entry-$side/entry.state" \
  -dump-mem-at "1187954188:lin:0:1048576:/out/oracle/cleanup79-r88-entry-$side/mem-in.bin"
