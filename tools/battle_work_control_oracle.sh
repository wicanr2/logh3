#!/usr/bin/env bash
# 第55輪#89：正常S3首戰WORK控制與經費的原始寫入。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r55-work-<a|b>}"
[[ "$name" =~ ^normal89-r55-work-(a|b)$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/oracle"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state -steps 435000001 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 -click-hold 3000000 \
  -poke 'lin:5E00C@300000000=7E 16 DC D3' \
  -watch 5DA46-5DA4B,5DA6A-5DA71,5DAB2-5DAB9 \
  -regs-from 300000000 -regs-max 40 -regs-at 12AA:000E,12AA:02D4,297B:10DE,1588:02CE \
  -save-state "435000000:/out/oracle/$name/machine.state" \
  -dump-mem-at "434999999:lin:0:1048576:/out/oracle/$name/entry.bin"
test -s "$LOGH3_ORACLE_OUTPUT_ROOT/$name/machine.state"
