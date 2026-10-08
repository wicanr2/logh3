#!/usr/bin/env bash
# 用 C 重寫版重算向量檔並逐筆比對。只讀向量檔與 decomp/ 原始碼，不需要原版與執行器。
#
#   decomp/tools/check_vectors.sh <向量檔>
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
vec="$(realpath "${1:?用法: check_vectors.sh <向量檔>}")"
test -f "$vec"
build="$(mktemp -d "$ROOT/workplace/decomp-build.XXXXXX")"
trap 'rm -rf "$build"' EXIT
docker run --rm --network none --memory 512m --cpus 2 --pids-limit 64 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  -v "$ROOT/decomp:/d:ro" -v "$build:/build" -v "$vec:/vectors.txt:ro" \
  golang:1.24-bookworm sh -c '
    set -e
    gcc -std=gnu99 -O1 -Wall -Wextra -Werror -I/d/include \
      /d/src/gin3ps_203A_0424_fleet_attack.c /d/test/check_fleet_attack.c -o /build/check_fleet_attack
    /build/check_fleet_attack /vectors.txt
  '
