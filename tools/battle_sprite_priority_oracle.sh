#!/usr/bin/env bash
# 第54輪#89：正常S3首戰圖示建立／順位，原始檔及既有收據保持。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r54-<init|low-first|high-first|equal>-<explore|a|b>}"
[[ "$name" =~ ^normal89-r54-(init|low-first|high-first|equal)-(explore|a|b)$ ]] || exit 1
kind="${BASH_REMATCH[1]}"
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
outdir=oracle
[[ "$name" == *-explore ]] && outdir=explore
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/$outdir"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$kind" != init ]]; then
  side=a
  [[ "$name" == *-b ]] && side=b
  test -f "$ROOT/workplace/oracle/normal89-r54-init-$side/machine.state"
  p82=80; p87=80
  [[ "$kind" == low-first ]] && p82=70
  [[ "$kind" == high-first ]] && p87=70
  # 明示診斷：池82／87及其顯示物件移到世界480,300，順位與raw池一致。
  # 原物件位置為world-view；dirty FF請執行器正常重畫。不改正式遊戲資料。
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state "/out/oracle/normal89-r54-init-$side/machine.state" \
    -steps 455000001 -tick 20000 \
    -poke "lin:681DE@438000000=$p82" -poke "lin:68364@438000000=$p87" \
    -poke 'lin:681E2@438000000=E0 01 2C 01' -poke 'lin:68368@438000000=E0 01 2C 01' \
    -poke "lin:7106F@438000000=FF 00 $p82" -poke "lin:7110F@438000000=FF 00 $p87" \
    -poke 'lin:71074@438000000=20 01 C8 00' -poke 'lin:71114@438000000=20 01 C8 00' \
    -mouse-at 440000000 -mouse-x 100 -mouse-y 100 \
    -regs-from 435000000 -regs-max 16 -regs-at 25DB:0915,25DB:0939,35D1:02AB \
    -save-state "455000000:/out/$outdir/$name/machine.state" \
    -dump-mem-at "454999999:lin:0:1048576:/out/$outdir/$name/end.bin"
  test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
  exit
fi
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000001 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
  -poke 'lin:5E00C@300000000=7E 16 DC D3' \
  -regs-from 300000000 -regs-max 80 -regs-at 25DB:0007,25DB:0178,25DB:01A8,25DB:01B0,25DB:0915 \
  -save-state "435000000:/out/$outdir/$name/machine.state" \
  -dump-mem-at "434999999:lin:0:1048576:/out/$outdir/$name/entry.bin"
test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
