#!/usr/bin/env bash
# 通用差異檢查：編譯 decomp/src 下所有 C 參考實作，讀向量檔重跑並比對。
#
#   decomp/tools/check_diff.sh <向量檔>
# 只讀向量檔、基底傾印（與向量檔同目錄）與 decomp/ 原始碼，不需要原版與執行器。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
vec="$(realpath "${1:?用法: check_diff.sh <向量檔>}")"
test -f "$vec"
build="$(mktemp -d "$ROOT/workplace/decomp-build.XXXXXX")"
trap 'rm -rf "$build"' EXIT
docker run --rm --network none --memory 1g --cpus 2 --pids-limit 64 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  -v "$ROOT/decomp:/d:ro" -v "$build:/build" -v "$(dirname "$vec"):/v:ro" \
  golang:1.24-bookworm sh -c '
    set -e
    gcc -std=gnu99 -O1 -Wall -Wextra -Werror -I/d/include /d/src/f_*.c /d/test/check_diff.c -o /build/check_diff
    /build/check_diff /v/'"$(basename "$vec")"'
  '
