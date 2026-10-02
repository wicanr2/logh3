#!/usr/bin/env bash
# #79 第三階段正式固定原版收據；契約見第39輪，禁止覆寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?用法: bash tools/battle_spotting_oracle.sh battle79-spot-r39-<s1|s3>-<a|b>}"
[[ "$name" =~ ^battle79-spot-r39-(s1|s3)-[ab]$ ]] || exit 1
scenario="${BASH_REMATCH[1]}"
if [[ "$scenario" == s1 ]]; then
  parent="battle79-round36-s1-boundary-explore/end.state"
  end=1270680843
  first=1265700000
  files=12
else
  parent="battle79-move-entry-a/machine.state"
  end=1256470039
  first=1226630000
  files=42
fi
test -d "$ROOT/workplace"
test -d "$GOLEM/workplace"
test -f "$ROOT/workplace/oracle/$parent"
test ! -e "$ROOT/workplace/oracle/$name"
test ! -e "$GOLEM/workplace/logh3-$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
bash "$ROOT/tools/py.sh" /work/tools/scripts/battle_spotting_plan.py "$name"
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
DOSGOLEM_GO_CMD=mkdir bash "$GOLEM/tools/go.sh" "/src/workplace/logh3-$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out:ro"
bash "$ROOT/tools/oracle.sh" "$name" -load-state "/out/oracle/$parent" \
  -steps "$end" -tick 20000 -regs-from "$first" -regs-max 256 \
  -regs-at 181B:0262,181B:028C,2526:00F8,2526:033B,3826:0298,3765:006C,25DB:0E78 "${flags[@]}"
LOGH3_EXTRA_RO="$GOLEM/workplace/logh3-$name" bash "$ROOT/tools/py.sh" -c \
  'from pathlib import Path; import os, shutil, sys; dst=Path("workplace/oracle")/sys.argv[1]; assert dst.stat().st_uid==os.getuid(); files=list(Path("/ref").iterdir()); assert len(files)==int(sys.argv[2]); [shutil.copyfile(p,dst/p.name) for p in files]' "$name" "$files"
