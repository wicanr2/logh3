#!/usr/bin/env bash
# 第56輪：正常入口與明示方向物件字鍵反例；不覆寫舊收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r56-<base|key-high-joined>-<a|b|explore>}"
[[ "$name" =~ ^normal89-r56-(base|key-high-joined)-(a|b|explore)$ ]] || exit 1
kind="${BASH_REMATCH[1]}"
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/oracle"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
poke='lin:5E00C@300000000=7E 16 DC D3'
if [[ "$kind" == key-high-joined ]]; then
  # 明示診斷：在第一個方向物件建立前只改其word+4，驗排序後覆蓋。
  # probe的-poke是單值旗標，兩筆必須用分號合成一次傳入。
  poke+=';lin:71022@407194660=FF FF'
fi
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000001 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
  -poke "$poke" \
  -watch 71022-71023,74C24-74C25 \
  -regs-from 407194000 -regs-max 80 -regs-at 25DB:00FB,35D1:0009,25DB:0BD2,35D1:02AB \
  -save-state "435000000:/out/oracle/$name/machine.state" \
  -dump-mem-at "434999999:lin:0:1048576:/out/oracle/$name/entry.bin"
test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
