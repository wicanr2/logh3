#!/usr/bin/env bash
# 產生 oracle 收據：用 dosgolem 跑原版，倒出色號陣列、調色盤與 PNG。
#
#   tools/oracle.sh <批次名> [probe 的額外旗標...]
#
# ⚠ **收據綁執行器版本。** dosgolem 的行為一改（即使是改對），
# 同一組輸入序列跑出來的畫面就可能不同，收據必須重做。所以這支腳本
# 會把 dosgolem 的 Git HEAD 一起寫進 meta.txt——沒有那一行的收據不算數。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?用法: tools/oracle.sh <批次名> [額外旗標...]}"; shift
out="$ROOT/workplace/oracle/$name"
mkdir -p "$out"

head=$(git -C "$GOLEM" rev-parse --short HEAD)
dirty=$(git -C "$GOLEM" status --porcelain | wc -l)

cd "$GOLEM"
DOSGOLEM_ORIG="$ROOT/org_game/extracted" tools/go.sh run ./cmd/probe \
  -exe /orig/GIN3.COM -root /orig "$@" \
  -dump-ega "/src/workplace/oracle-tmp.png" \
  -dump-palette "/src/workplace/oracle-tmp-pal.bin" > "$out/report.txt" 2>&1 || true
cp "$GOLEM/workplace/oracle-tmp.png" "$out/screen.png"
cp "$GOLEM/workplace/oracle-tmp.bin" "$out/screen.bin"
cp "$GOLEM/workplace/oracle-tmp-pal.bin" "$out/pal.bin"

{
  echo "批次: $name"
  echo "日期: $(date -Iseconds)"
  echo "dosgolem HEAD: $head（未提交檔案 $dirty 個）"
  echo "probe 旗標: $*"
  echo "GIN3.COM SHA-256: $(sha256sum "$ROOT/org_game/extracted/GIN3.COM" | cut -d' ' -f1)"
  echo "screen.bin SHA-256: $(sha256sum "$out/screen.bin" | cut -d' ' -f1)"
  echo "pal.bin SHA-256: $(sha256sum "$out/pal.bin" | cut -d' ' -f1)"
} > "$out/meta.txt"
cat "$out/meta.txt"
