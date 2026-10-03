#!/usr/bin/env bash
# 第52輪：由已驗空白Input續跑，移入分隊；新收據不覆寫舊輸出。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r52-<flagship|squad|fleet-target|squad-target>-<explore|a|b>}"
[[ "$name" =~ ^normal89-r52-(flagship|squad|fleet-target|squad-target)-(explore|a|b)$ ]] || exit 1
kind="${BASH_REMATCH[1]}"
side="${BASH_REMATCH[2]}"
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
parent="oracle/normal89-r51-empty-a/machine.state"
[[ "$side" == b ]] && parent="oracle/normal89-r51-empty-b/machine.state"
test -f "$ROOT/workplace/$parent"
outdir=oracle
[[ "$side" == explore ]] && outdir=explore
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/$outdir"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
extra=()
mx=248; my=208
case "$kind" in
  squad|squad-target) mx=249; my=184 ;;
esac
case "$kind" in
  fleet-target) extra=(-poke 'lin:611C0@457000000=40 02 90 01') ;;
  squad-target) extra=(-poke 'lin:681D6@457000000=40 02 90 01') ;;
esac
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state "/out/$parent" -steps 475000001 -tick 20000 \
  "${extra[@]}" -mouse-at 460000000 -mouse-x "$mx" -mouse-y "$my" \
  -regs-from 460000000 -regs-max 16 -regs-at 1727:0677,1727:0736,39D7:08F4,39D7:0936 \
  -save-state "475000000:/out/$outdir/$name/machine.state" \
  -dump-mem-at "474999999:lin:0:1048576:/out/$outdir/$name/end.bin"
test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
