#!/usr/bin/env bash
# 第49輪#89，月底旗標自然0與同盟旗標1的隔離對照。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r49-<natural|due>-<a|b>}"
[[ "$name" =~ ^normal89-r49-(natural|due)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
poke='lin:5E00C@300000000=7E 16 DC D3'
[[ "$name" == *-due-* ]] && poke+=';lin:5DB25@300000000=01'
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000000 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
  -poke "$poke" -watch 5DB24-5DB25 -regs-from 300000000 -regs-max 20 \
  -regs-at 0654:0034,0654:011E,2703:03AF,2B5F:0103,12AA:000E \
  -dump-mem-at "300000001:lin:0:1048576:/out/oracle/$name/loaded.bin;434999999:lin:0:1048576:/out/oracle/$name/entry.bin"
