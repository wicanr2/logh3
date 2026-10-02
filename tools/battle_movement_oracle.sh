#!/usr/bin/env bash
# #79第二階段原版重播；契約／範圍見docs/re/371-battle-movement-phase.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?用法: bash tools/battle_movement_oracle.sh <entry|init|end> <新批次名>}"
name="${2:?必須指定新批次名}"
[[ "$name" =~ ^battle79-move-[a-z0-9-]+$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM"
test -f "$GOLEM/workplace/fcmd/eng-t5.state"
[[ "$(sha256sum "$GOLEM/workplace/fcmd/eng-t5.state" | cut -d' ' -f1)" == c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1 ]] || exit 1
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
case "$mode" in
  entry) end=1213409022 ;;
  init) end=1213433818 ;;
  end) end=1226638240 ;;
  *) exit 1 ;;
esac
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "$name" -load-state /src/workplace/fcmd/eng-t5.state \
  -clicks 1140000000:240:200,1195000000:240:200,1210000000:240:200 \
  -click-hold 3000000 -tick 20000 -steps "$end" \
  -poke 'lin:5E00C@1211893242=78 56 34 12' \
  -regs-from 1213408900 -regs-max 64 \
  -regs-at 181B:00FA,1C03:0006,1C61:0005,1C03:006A,1C61:0153,181B:0262 \
  -dump-mem-at "1213409021:lin:0:1048576:/out/oracle/$name/mem-in.bin;$((end-1)):lin:0:1048576:/out/oracle/$name/mem-out.bin" \
  -save-state "$((end-1)):/out/oracle/$name/machine.state" \
  -peek lin:5E00C:4,lin:5E3C4:60,lin:61AE4:4
