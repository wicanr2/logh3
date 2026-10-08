#!/usr/bin/env bash
# 在 Docker 內執行 ledger.py（python3 -I）。
#
#   decomp/tools/ledger.sh build  <映像批次名>     由 workplace/ida/functions.json 與該批次的映像產生台帳
#   decomp/tools/ledger.sh summary|show|order|mark ...  其餘參數原樣交給 ledger.py（台帳目錄固定 decomp/ledger）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cmd="${1:?用法: ledger.sh build <映像批次名> | summary | show SEG:OFF | order [CLASS...] | mark SEG:OFF LEVEL SOURCE EVIDENCE}"; shift
run() {
  docker run --rm --network none --memory 512m --cpus 1 --pids-limit 32 \
    --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
    -v "$ROOT/decomp:/d" -v "$ROOT/workplace/ida:/ida:ro" \
    -v "$ROOT/workplace/decomp/image:/img:ro" golang:1.24-bookworm python3 -I /d/tools/ledger.py "$@"
}
case "$cmd" in
  build)
    b="${1:?映像批次名}"
    run build /ida/functions.json "/img/$b/relocs.json" "/img/$b/gin3ps-unrelocated.bin" /d/ledger "/img/$b/gin3ps-final.bin" ;;
  summary) run summary /d/ledger ;;
  show) run show /d/ledger "$@" ;;
  order) run order /d/ledger "$@" ;;
  mark) run mark /d/ledger "$@" ;;
  *) echo "未知子命令 $cmd" >&2; exit 2 ;;
esac
