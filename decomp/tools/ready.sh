#!/usr/bin/env bash
# 列出「可以動手」的規則層函式：class 是 game-pure／game-logic、尚未有 status.d 標記，
# 而且呼叫的每個函式都已經有 status.d 標記（diff-tested 或更高）。
#
#   decomp/tools/ready.sh [排除清單檔]
# 排除清單檔每行第一欄是 SEG:OFF（例如 workplace/decomp/waves/w2-13.txt），用來略過已經派給別人的函式。
# 輸出（tab 分隔）：SEG:OFF、bytes、class。純文字處理，只讀台帳，不寫任何檔案。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
L="$ROOT/decomp/ledger"
excl="${1:-/dev/null}"
done_list="$(mktemp)"
trap 'rm -f "$done_list"' EXIT
ls "$L/status.d" 2>/dev/null | sed 's/\.tsv$//' > "$done_list"
awk -F'\t' -v donef="$done_list" -v exclf="$excl" '
BEGIN {
  while ((getline l < donef) > 0) done[l] = 1
  while ((getline l < exclf) > 0) { if (l ~ /^#/) continue; split(l, p, /[ \t]/); skip[p[1]] = 1 }
}
FILENAME ~ /functions\.tsv$/ {
  if (FNR > 1) { split($3, a, ":"); key[$2] = a[1] "_" a[2]; cls[$2] = $9; sz[$2] = $5; sg[$2] = $3 }
  next
}
FNR > 1 { callees[$1] = callees[$1] " " $2 }
END {
  for (f in key) {
    if ((cls[f] != "game-pure" && cls[f] != "game-logic") || (key[f] in done) || (sg[f] in skip)) continue
    ok = 1
    n = split(callees[f], cs, " ")
    for (i = 1; i <= n; i++) { c = cs[i]; if (c != f && !(key[c] in done)) ok = 0 }
    if (ok) print sg[f] "\t" sz[f] "\t" cls[f]
  }
}' "$L/functions.tsv" "$L/callgraph.tsv" | sort
