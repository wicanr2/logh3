#!/usr/bin/env bash
# #79第④步自然入口與參戰門檻；只用固定dosgolem，不覆寫歷史收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?combat79-r62-natural-<a|b>或combat79-r62-gate-<案例>-<a|b>}"
[[ "$name" =~ ^combat79-r62-(natural-[ab]|gate-[a-z0-9-]+-[ab]|command-[ab])$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$name" == *-command-* ]]; then
  bash "$ROOT/tools/py.sh" /work/tools/scripts/fleet_combat_entry_plan.py "$name"
  mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
  bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
  exit
fi
if [[ "$name" == *-gate-* ]]; then
  test -f "$ROOT/workplace/oracle/combat79-r62-natural-a/entry.state"
  bash "$ROOT/tools/py.sh" /work/tools/scripts/fleet_combat_entry_plan.py "$name"
  mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
  limit=484700000
  [[ "$name" == *-near-human-* ]] && limit=486000000
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state "/out/oracle/combat79-r62-natural-a/entry.state" \
    -steps "$limit" -tick 20000 -regs-from 484357727 -regs-max 128 \
    -regs-at 181B:028C,181B:0376,181B:0385,1DD6:0009,1DD6:000D,1DD6:0040,1DD6:0069,1DD6:0152,1DD6:030F,1ED2:0001,1ED2:0097,1ED2:01FB,1ED2:0205,1ED2:02AB,1ED2:03E9,1629:0005,165F:034F,1F6F:000F,3CD5:000C \
    "${flags[@]}"
  exit
fi
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 490000000 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200,450000000:100:60,455000000:100:70 -click-hold 3000000 \
  -poke 'lin:5E00C@300000000=7E 16 DC D3;lin:5E00C@435000000=78 56 34 12' \
  -save-state "484357727:/out/oracle/$name/entry.state" \
  -regs-from 484357700 -regs-max 512 \
  -regs-at 181B:028C,181B:0376,181B:0385,1DD6:0009,1DD6:000D,1DD6:0040,1DD6:0048,1DD6:0152,1DD6:026B,1DD6:030F,1ED2:0001,1F6F:000F,3CD5:000C \
  -dump-mem-at "484357727:lin:0:1048576:/out/oracle/$name/entry-mem.bin;489999999:lin:0:1048576:/out/oracle/$name/end-mem.bin"
