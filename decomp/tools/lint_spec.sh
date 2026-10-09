#!/usr/bin/env bash
# 描述檔與突變清單的快速檢查（Docker 內，唯讀掛載）。
#   decomp/tools/lint_spec.sh <函式id>
set -euo pipefail
ROOT="${GIN3_REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
D="${DECOMP_DIR:-$ROOT/decomp}"
id="${1:?用法: lint_spec.sh <函式id>}"
docker run --rm --network none --memory 128m --cpus 1 --pids-limit 16 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  -v "$D:/d:ro" golang:1.24-bookworm python3 -I /d/tools/lint_spec.py "$id" /d
