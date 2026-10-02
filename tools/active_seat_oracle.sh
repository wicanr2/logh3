#!/usr/bin/env bash
# 第48輪#89，正常活動陣營／座位；入口docs/goal/2026-10-03-active-seat-round-48.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r48-<clock|alliance|versus>-<a|b>}"
[[ "$name" =~ ^normal89-r48-(clock|alliance|versus)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-base.state"
test -f "$ROOT/workplace/oracle/normal89-r46-menu-a/menu.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$name" == *-clock-* ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /src/workplace/fcmd/fcmd-base.state -steps 450000000 -tick 20000 \
    -clicks 421000000:544:85,424000000:560:85 -click-hold 3000000 \
    -watch 5DA4D-5DA4F -regs-from 420000000 -regs-max 16 \
    -regs-at 04F8:01B9,04F8:030E,2D10:0005 \
    -dump-mem-at "420000010:lin:0:1048576:/out/oracle/$name/initial.bin;426100000:lin:0:1048576:/out/oracle/$name/thinking.bin;426215238:lin:0:1048576:/out/oracle/$name/ai-before.bin;427763580:lin:0:1048576:/out/oracle/$name/ai-after.bin;430530800:lin:0:1048576:/out/oracle/$name/processing.bin;449999999:lin:0:1048576:/out/oracle/$name/command.bin"
elif [[ "$name" == *-alliance-* ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /out/oracle/normal89-r46-menu-a/menu.state -steps 435000000 -tick 20000 \
    -clicks 240000000:318:188,301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
    -poke 'lin:5E00C@300000000=7E 16 DC D3' \
    -watch 5DA4D-5DA4F -regs-from 230000000 -regs-max 16 \
    -regs-at 04F8:01B9,04F8:030E,12AA:000E \
    -dump-mem-at "299999999:lin:0:1048576:/out/oracle/$name/loaded.bin;434999999:lin:0:1048576:/out/oracle/$name/entry.bin"
else
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /out/oracle/normal89-r46-menu-a/menu.state -steps 350000000 -tick 20000 \
    -clicks 240000000:318:208,260000000:432:230,310000000:544:85,313000000:560:85,330000000:432:230 -click-hold 3000000 \
    -watch 5DA4D-5DA4F -regs-from 230000000 -regs-max 16 \
    -regs-at 04F8:01B9,04F8:030E \
    -dump-mem-at "259999999:lin:0:1048576:/out/oracle/$name/loaded.bin;299999999:lin:0:1048576:/out/oracle/$name/empire.bin;329999999:lin:0:1048576:/out/oracle/$name/handoff.bin;349999999:lin:0:1048576:/out/oracle/$name/alliance.bin"
fi
