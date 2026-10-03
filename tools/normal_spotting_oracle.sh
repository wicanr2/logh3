#!/usr/bin/env bash
# #79正常S3首回合的第三階段邊界；沿第50輪固定執行器與玩家輸入。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal79-r61-<a|b>}"
[[ "$name" =~ ^normal79-r61-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 484357729 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200,450000000:100:60,455000000:100:70 -click-hold 3000000 \
  -poke 'lin:5E00C@300000000=7E 16 DC D3;lin:5E00C@435000000=78 56 34 12' \
  -regs-from 435000000 -regs-max 512 \
  -regs-at 181B:0262,181B:028C,2526:00F8,2526:033B,2526:04B1,3826:0298,3765:006C,25DB:0E78 \
  -dump-at "483778107:/out/oracle/$name/entry.png;483798431:/out/oracle/$name/label.png;484103338:/out/oracle/$name/seat1.png;484357728:/out/oracle/$name/end.png" \
  -dump-mem-at "434999999:lin:0:1048576:/out/oracle/$name/battle.bin;483778107:lin:0:1048576:/out/oracle/$name/entry-mem.bin;483798431:lin:0:1048576:/out/oracle/$name/label-mem.bin;484103338:lin:0:1048576:/out/oracle/$name/seat1-mem.bin;484357728:lin:0:1048576:/out/oracle/$name/end-mem.bin"
