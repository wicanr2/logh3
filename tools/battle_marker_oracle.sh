#!/usr/bin/env bash
# #79正常S3進場的小地圖標記收據，版本與輸入固定；只接受新批次。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?用法: bash tools/battle_marker_oracle.sh battle79-marker-r33-fixed-<a|b>}"
[[ "$name" =~ ^battle79-marker-r33-fixed-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-s3-base.state"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
LOGH3_EXTRA_RO="$GOLEM/workplace/fcmd" bash "$ROOT/tools/py.sh" -c \
  'from pathlib import Path; import hashlib; assert hashlib.sha256(Path("/ref/fcmd-s3-base.state").read_bytes()).hexdigest() == "0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7"'
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /src/workplace/fcmd/fcmd-s3-base.state \
  -steps 435000000 -tick 20000 \
  -clicks 301000000:544:85,304000000:560:85,405000000:240:200 \
  -click-hold 3000000 -poke 'lin:5E00C@300000000=78 56 34 12' \
  -regs-from 300000000 -regs-max 128 \
  -regs-at 297B:08D8,26E5:00AC,1727:02E2 \
  -dump-mem-at "434999999:lin:0:1048576:/out/oracle/$name/mem.bin"
