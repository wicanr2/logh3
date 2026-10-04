#!/usr/bin/env bash
# #79 子步8-2：由既有第五回合狀態自然進場，收據各自保存、不覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?ownership79-r82-<案例>-<a|b>}"
[[ "$name" =~ ^ownership79-r82-(entry|control|capture0|capture1|partial0|partial1|capture0-complete|capture1-complete|partial0-complete|partial1-complete|capture0-second|capture1-second|capture0-finish|capture1-finish)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ ! "$name" =~ ^ownership79-r82-entry- ]]; then
  test -f "$ROOT/workplace/oracle/ownership79-r82-entry-a/entry.state"
  flags=(-load-state /out/oracle/ownership79-r82-entry-a/entry.state
    -steps 1188500000 -tick 20000
    -regs-from 1187957002 -regs-max 40
    -regs-at 185C:02DA,185C:03C6,185C:03BD,185C:0491,185C:060D,185C:09BF,1213:007B,3826:0405,12AA:0245,3958:0215,3765:006C)
  poke="lin:5DA51@1187957003=00,00"
  case "$name" in
    *capture0*) poke+=";lin:54571@1187957003=00;lin:545C3@1187957003=00;lin:54615@1187957003=00" ;;
    *capture1*) poke+=";lin:54571@1187957003=01;lin:545C3@1187957003=01;lin:54615@1187957003=01" ;;
    *partial0*) poke+=";lin:5FFAA@1187957003=00" ;;
    *partial1*) poke+=";lin:5FFAA@1187957003=01" ;;
  esac
  limit=1188500000
  if [[ "$name" == *-complete-* ]]; then
    limit=1189100000
  elif [[ "$name" == *-second-* ]]; then
    limit=1189300000
    flags+=(-clicks 1188600000:240:200 -click-hold 100000)
  elif [[ "$name" == *-finish-* ]]; then
    limit=1191000000
    flags+=(-clicks 1188600000:240:200,1189400000:240:200 -click-hold 100000)
  fi
  # 每案只有一個steps上限，不依賴重複旗標覆蓋。
  flags[3]="$limit"
  flags+=(-poke "$poke"
    -save-state "$((limit-1)):/out/oracle/$name/end.state"
    -dump-mem-at "$((limit-1)):lin:0:1048576:/out/oracle/$name/mem-out.bin")
  bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
  exit
fi
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/eng-t5.state \
  -steps 1187957003 -tick 20000 \
  -clicks 1140000000:240:200 -click-hold 3000000 \
  -poke lin:5E00C@1187900000=78,56,34,12 \
  -regs-from 1187900000 -regs-max 40 \
  -regs-at 181B:03F0,185C:001D,185C:02DA,185C:03C6,185C:0491,1213:007B,3826:0405 \
  -save-state "1187957002:/out/oracle/$name/entry.state" \
  -dump-mem-at "1187957002:lin:0:1048576:/out/oracle/$name/mem-in.bin"
