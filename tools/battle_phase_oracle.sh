#!/usr/bin/env bash
# #79固定移動逐圈收據；只接受新批次，原版snapshot唯讀掛載。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?用法: bash tools/battle_phase_oracle.sh battle79-phase-r34-<a|b>}"
[[ "$name" =~ ^battle79-phase-r34-[ab]$ ]] || exit 1
test -d "$ROOT/workplace"
test -d "$GOLEM/workplace"
test -f "$ROOT/workplace/oracle/battle79-move-entry-a/machine.state"
test ! -e "$ROOT/workplace/oracle/$name"
test ! -e "$GOLEM/workplace/logh3-$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
bash "$ROOT/tools/py.sh" /work/tools/scripts/battle_phase_plan.py "$name"
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
DOSGOLEM_GO_CMD=mkdir bash "$GOLEM/tools/go.sh" "/src/workplace/logh3-$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out:ro"
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state /out/oracle/battle79-move-entry-a/machine.state \
  -steps 1226638240 -tick 20000 -regs-from 1213409021 -regs-max 64 \
  -regs-at 181B:0135 "${flags[@]}"
LOGH3_EXTRA_RO="$GOLEM/workplace/logh3-$name" bash "$ROOT/tools/py.sh" -c \
  'from pathlib import Path; import os, shutil, sys; dst=Path("workplace/oracle")/sys.argv[1]; assert dst.stat().st_uid==os.getuid(); files=list(Path("/ref").iterdir()); assert len(files)==105; [shutil.copyfile(p,dst/p.name) for p in files]' "$name"
