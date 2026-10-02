#!/usr/bin/env bash
# #89 回合計數器證據；入口 docs/goal/2026-10-03-turn-counter-round-47.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r47-<clock|battle>-<a|b>}"
[[ "$name" =~ ^normal89-r47-(clock|battle)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-base.state"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$name" == *-clock-* ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /src/workplace/fcmd/fcmd-base.state -steps 450000000 -tick 20000 \
    -clicks 421000000:544:85,424000000:560:85 -click-hold 3000000 \
    -watch 5DA48-5DA49 -regs-from 420000000 -regs-max 24 \
    -regs-at 0374:03E8,056A:078E,0374:03DD \
    -dump-mem-at "420000010:lin:0:1048576:/out/oracle/$name/initial.bin;435711455:lin:0:1048576:/out/oracle/$name/before.bin;435711456:lin:0:1048576:/out/oracle/$name/after.bin;449999999:lin:0:1048576:/out/oracle/$name/command.bin"
else
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000000 -tick 20000 \
    -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
    -poke 'lin:5E00C@300000000=7E 16 DC D3' \
    -watch 5DA48-5DA49 -regs-from 300000000 -regs-max 24 \
    -regs-at 0374:03E8,12AA:000E,0374:03DD \
    -dump-mem-at "300000001:lin:0:1048576:/out/oracle/$name/loaded.bin;434999999:lin:0:1048576:/out/oracle/$name/entry.bin"
fi
