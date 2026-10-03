#!/usr/bin/env bash
# 第50輪#89；固定S3自然路徑，監看移動後的航向與座標寫入者。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r50-arrival-<explore|a|b>}"
[[ "$name" =~ ^normal89-r50-arrival-(explore|a|b|spot-explore|spot-a|spot-b)$ ]] || exit 1
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
if [[ "$name" == *-spot-* ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 800000000 -tick 20000 \
    -clicks 301000000:544:85,304000000:560:85,405000000:240:200,450000000:100:60,455000000:100:70 -click-hold 3000000 \
    -poke 'lin:5E00C@300000000=7E 16 DC D3;lin:5E00C@435000000=78 56 34 12' \
    -regs-from 435000000 -regs-max 512 \
    -regs-at 181B:0262,181B:028C,2526:00F8,2526:033B,2526:04B1,3826:0298,3765:006C,25DB:0E78 \
    -dump-mem-at "434999999:lin:0:1048576:/out/$outdir/$name/entry.bin;799999999:lin:0:1048576:/out/$outdir/$name/end.bin"
  exit
fi
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000000 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
  -poke 'lin:5E00C@300000000=7E 16 DC D3' \
  -watch 60FBE-60FC0,60FD0-60FDB,611A6-611A8,611B8-611C3,6138E-61390,613A0-613AB,67900-67913,6813A-6814D,687EE-68801 \
  -regs-from 300000000 -regs-max 100 -regs-at 113A:047D,297B:08D8,297B:0ABF,12AA:000E \
  -dump-mem-at "300000001:lin:0:1048576:/out/$outdir/$name/loaded.bin;404999999:lin:0:1048576:/out/$outdir/$name/pre-battle.bin;434999999:lin:0:1048576:/out/$outdir/$name/entry.bin"
