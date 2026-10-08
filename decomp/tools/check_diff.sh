#!/usr/bin/env bash
# 通用差異檢查：編譯目標函式的 C 參考實作（加上台帳中已 diff-tested 的函式，供呼叫），
# 讀向量檔重跑並比對。
#
#   decomp/tools/check_diff.sh <向量檔>
# 只編譯目標與已通過的函式，其他人寫到一半的檔案不會影響這裡。
# 只讀向量檔、基底傾印（與向量檔同目錄）與 decomp/ 原始碼，不需要原版與執行器。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
vec="$(realpath "${1:?用法: check_diff.sh <向量檔>}")"
test -f "$vec"
id="$(grep -m1 '^H ' "$vec" | sed -E 's/.*func=([^ ]+).*/\1/')"
test -n "$id"
test -f "$ROOT/decomp/src/f_$id.c" || { echo "找不到 decomp/src/f_$id.c" >&2; exit 2; }
srcs="/d/src/f_$id.c"
if [[ -d "$ROOT/decomp/ledger/status.d" ]]; then
  for t in "$ROOT"/decomp/ledger/status.d/*.tsv; do
    other="$(basename "$t" .tsv)"
    [[ "$other" == "$id" ]] && continue
    [[ -f "$ROOT/decomp/src/f_$other.c" ]] && srcs="$srcs /d/src/f_$other.c"
  done
fi
build="$(mktemp -d "$ROOT/workplace/decomp-build.XXXXXX")"
trap 'rm -rf "$build"' EXIT
docker run --rm --network none --memory 1g --cpus 2 --pids-limit 64 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  -v "$ROOT/decomp:/d:ro" -v "$build:/build" -v "$(dirname "$vec"):/v:ro" \
  golang:1.24-bookworm sh -c '
    set -e
    gcc -std=gnu99 -O1 -Wall -Wextra -Werror -I/d/include '"$srcs"' /d/test/check_diff.c -o /build/check_diff
    /build/check_diff /v/'"$(basename "$vec")"'
  '
