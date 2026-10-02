#!/usr/bin/env bash
# 第40輪原版固定分支雙跑。入口與限制見 docs/playtest/046-battle-spotting-phase.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?需要 battle79-discovery-r40-<分支>-<a|b>}"
test -d "$ROOT/workplace"
test -d "$GOLEM/workplace"
test ! -e "$ROOT/workplace/oracle/$name"
test ! -e "$GOLEM/workplace/logh3-$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
files=$(bash "$ROOT/tools/py.sh" /work/tools/scripts/battle_discovery_plan.py "$name")
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
DOSGOLEM_GO_CMD=mkdir bash "$GOLEM/tools/go.sh" "/src/workplace/logh3-$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out:ro"
bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
LOGH3_EXTRA_RO="$GOLEM/workplace/logh3-$name" bash "$ROOT/tools/py.sh" -c \
  'from pathlib import Path; import os, shutil, sys; dst=Path("workplace/oracle")/sys.argv[1]; assert dst.stat().st_uid==os.getuid(); files=list(Path("/ref").iterdir()); assert len(files)==int(sys.argv[2]); [shutil.copyfile(p,dst/p.name) for p in files]' "$name" "$files"
