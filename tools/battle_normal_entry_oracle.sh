#!/usr/bin/env bash
# #79正常冷啟動→S3→首次戰略實行；不注入局面。
# 契約與限制見docs/goal/2026-10-03-battle-normal-entry-round-44.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?用法: bash tools/battle_normal_entry_oracle.sh battle79-normal-r44-<a|b>}"
[[ "$name" =~ ^battle79-normal-r44-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM/workplace"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "$name" -steps 435000000 -tick 20000 \
  -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -clicks 40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:168,301000000:544:85,304000000:560:85,405000000:240:200 \
  -click-hold 3000000 -poke 'lin:5E00C@300000000=7E 16 DC D3' \
  -regs-from 300000000 -regs-max 32 -regs-at 12AA:000E,297B:08D8,26E5:00AC \
  -dump-mem-at "299999999:lin:0:1048576:/out/oracle/$name/loaded-mem.bin;434999999:lin:0:1048576:/out/oracle/$name/entry-mem.bin" \
  -save-state "434999999:/out/oracle/$name/entry.state" \
  -peek lin:5E00C:4,lin:5DA44:32,lin:5E3C4:128
