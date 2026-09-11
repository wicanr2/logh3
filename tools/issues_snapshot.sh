#!/usr/bin/env bash
# 把 GitHub 上的 worklist issue 抓成 `workplace/issues.json`（gitignore）。
#
#   tools/issues_snapshot.sh
#
# **未完成項的權威是 GitHub issue**（`cmd/worklist/issues.go`）。`worklist`
# 那支工具跑在沒有網路的容器裡，所以要先在主機用 `gh` 抓一份快照給它讀。
#
# ⚠ **每次跑 verify／render／score 之前先抓一次。** 快照是某一刻的副本；
# 在 GitHub 上關了一張 issue 而沒有重抓，工具看到的就還是 open。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO="${LOGH3_REPO:-wicanr2/logh3}"
OUT="$ROOT/workplace/issues.json"
mkdir -p "$ROOT/workplace"
tmp="$OUT.tmp"
gh issue list -R "$REPO" --label worklist --state all --limit 1000 \
  --json number,title,body,state,labels > "$tmp"
mv "$tmp" "$OUT"
python3 - "$OUT" <<'PY'
import json, sys
d = json.load(open(sys.argv[1], encoding="utf-8"))
o = sum(1 for x in d if x["state"] == "OPEN")
print(f"抓到 {len(d)} 張 worklist issue（open {o}、closed {len(d)-o}）→ {sys.argv[1]}")
PY
