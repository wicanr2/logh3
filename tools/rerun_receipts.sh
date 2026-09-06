#!/usr/bin/env bash
# 重跑收據。**執行器一改就要跑這支**——同一組輸入在新版執行器下跑出來的
# 畫面可能不同，那份收據與所有依賴它的對拍結論同時失效（AGENTS.md）。
#
#   tools/rerun_receipts.sh              # 全部
#   tools/rerun_receipts.sh -m mouse-x   # 只重跑命令列含這個字串的
#   tools/rerun_receipts.sh -n 'main1e|cmd1e'  # 只重跑名字對得上的
#   JOBS=6 tools/rerun_receipts.sh       # 並行數（預設 4）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JOBS="${JOBS:-4}"
match=''
namere=''
while [[ "${1:-}" == -* ]]; do
  case "$1" in
    -m) match="$2"; shift 2 ;;
    -n) namere="$2"; shift 2 ;;
    *) echo "不認得的參數 $1" >&2; exit 2 ;;
  esac
done

n=0
for d in "$ROOT"/workplace/oracle/*/; do
  name=$(basename "$d")
  [[ -f "$d/meta.txt" ]] || continue
  flags=$(sed -n 's/^probe 旗標: //p' "$d/meta.txt")
  [[ -n "$flags" ]] || continue
  [[ -z "$match" || "$flags" == *"$match"* ]] || continue
  [[ -z "$namere" || "$name" =~ $namere ]] || continue
  exe=$(sed -n 's/^執行檔: //p' "$d/meta.txt")
  before=$(sha256sum "$d/screen.bin" 2>/dev/null | cut -d' ' -f1 || echo -)
  (
    LOGH3_EXE="${exe:-/orig/GIN3.COM}" "$ROOT/tools/oracle.sh" "$name" $flags >/dev/null 2>&1
    after=$(sha256sum "$d/screen.bin" | cut -d' ' -f1)
    [[ "$before" == "$after" ]] && echo "  $name 不變" || echo "  $name **變了**（$before → $after）"
  ) &
  n=$((n+1)); (( n % JOBS )) || wait
done
wait
echo "重跑 $n 份收據"
