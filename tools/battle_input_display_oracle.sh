#!/usr/bin/env bash
# 第51輪#89：正常S3首戰後只移動游標，原版檔及既有收據不覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r51-<base|flagship|squad|empty|panel>-<explore|a|b>}"
[[ "$name" =~ ^normal89-r51-(base|flagship|squad|empty|panel)-(explore|explore2|a|b)$ ]] || exit 1
kind="${BASH_REMATCH[1]}"
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
outdir=oracle
[[ "$name" == *-explore* ]] && outdir=explore
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/$outdir"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$kind" == base ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" \
    -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000001 -tick 20000 \
    -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
    -poke 'lin:5E00C@300000000=7E 16 DC D3' \
    -regs-from 405000000 -regs-max 16 -regs-at 1CC5:008E,1629:023B,12AA:02D4 \
    -save-state "435000000:/out/$outdir/$name/machine.state" \
    -dump-mem-at "434999999:lin:0:1048576:/out/$outdir/$name/entry.bin"
  test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
  exit
fi
parent="$ROOT/workplace/oracle/normal89-r51-base-a/machine.state"
[[ "$name" == *-b ]] && parent="$ROOT/workplace/oracle/normal89-r51-base-b/machine.state"
[[ "$name" == *-explore* ]] && parent="$ROOT/workplace/explore/normal89-r51-base-explore2/machine.state"
test -f "$parent"
case "$kind" in
  flagship) mx=248; my=208 ;;
  squad) mx=249; my=184 ;;
  empty) mx=100; my=100 ;;
  panel) mx=500; my=250 ;;
esac
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state "/out/${parent#"$ROOT/workplace/"}" -steps 455000001 -tick 20000 \
  -mouse-at 440000000 -mouse-x "$mx" -mouse-y "$my" \
  -regs-from 435000000 -regs-max 16 -regs-at 1CC5:008E,1629:023B,12AA:02D4,12F1:04C3,12F1:0539,25DB:0AF8 \
  -save-state "455000000:/out/$outdir/$name/machine.state" \
  -dump-mem-at "454999999:lin:0:1048576:/out/$outdir/$name/end.bin"
test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
