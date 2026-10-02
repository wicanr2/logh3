#!/usr/bin/env bash
# 第43輪固定來源標頭／移鏡；探索與正式收據分開。入口見docs/goal/2026-10-03-battle-discovery-scroll-round-43.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?需要 battle79-scroll-r43-<case>-<explore|a|b>}"
test -d "$ROOT/workplace"
test -d "$GOLEM/workplace"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]]
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]]
files=$(bash "$ROOT/tools/py.sh" /work/tools/scripts/battle_discovery_scroll.py flags "$name")
mapfile -t flags < "$ROOT/workplace/$name-flags.txt"
base=oracle
[[ "$name" != *-explore* ]] || base=explore
test ! -e "$ROOT/workplace/$base/$name"
test ! -e "$GOLEM/workplace/logh3-$name"
DOSGOLEM_GO_CMD=mkdir bash "$GOLEM/tools/go.sh" "/src/workplace/logh3-$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out:ro"
LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/$base" bash "$ROOT/tools/oracle.sh" "$name" "${flags[@]}"
LOGH3_EXTRA_RO="$GOLEM/workplace/logh3-$name" bash "$ROOT/tools/py.sh" -c \
  'from pathlib import Path; import os, shutil, sys; dst=Path("workplace")/sys.argv[1]/sys.argv[2]; assert dst.stat().st_uid==os.getuid(); files=list(Path("/ref").iterdir()); assert len(files)==int(sys.argv[3]); [shutil.copyfile(p,dst/p.name) for p in files]' "$base" "$name" "$files"
