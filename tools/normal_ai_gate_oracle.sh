#!/usr/bin/env bash
# #89 的 AI 入口閘門；來源與限制見第45輪 goal。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r45-<natural|zero>-<a|b>}"
[[ "$name" =~ ^normal89-r45-(natural|zero)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
poke='lin:5E00C@300000000=7E 16 DC D3'
if [[ "$name" == *-zero-* ]]; then
  poke+=';lin:5DA76@300000000=00'
fi
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000000 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
  -poke "$poke" -regs-from 300000000 -regs-max 32 \
  -regs-at 2D10:0005,2D10:000C,2D10:0017,2D10:0021,2D10:0030,2D10:0037,2E54:000C,12AA:000E \
  -watch 5DA76-5DA77 \
  -dump-mem-at "300000001:lin:0:1048576:/out/oracle/$name/loaded-mem.bin;306103453:lin:0:1048576:/out/oracle/$name/ai-before.bin;306103458:lin:0:1048576:/out/oracle/$name/ai-after.bin;434999999:lin:0:1048576:/out/oracle/$name/entry-mem.bin" \
  -peek lin:5E00C:4,lin:5DA44:64
