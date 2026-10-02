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
# 預設跑 GIN3.COM；結局那類獨立執行檔用 LOGH3_EXE 換掉。
EXE="${LOGH3_EXE:-/orig/GIN3.COM}"
out="${LOGH3_ORACLE_OUTPUT_ROOT:-$ROOT/workplace/oracle}/$name"
mkdir -p "$out"

head=$(git -C "$GOLEM" rev-parse --short HEAD)
dirty=$(git -C "$GOLEM" status --porcelain | wc -l)

cd "$GOLEM"
# ⚠ **`-empty-env` 不是可有可無的旗標。** 環境字串區非空時，`GIN3.COM`
# 的載入器會去掃描它，走上一條與執行器的記憶體配置打架的路，最後
# `retf` 進垃圾（`docs/re/269`）。少了它，這個遊戲一個畫面都跑不出來。
DOSGOLEM_ORIG="$ROOT/org_game/extracted" tools/go.sh run ./cmd/probe \
  -exe "$EXE" -root /orig -empty-env "$@" \
  -dump-screen "/src/workplace/oracle-tmp-$name.bin" \
  -dump-screen-png "/src/workplace/oracle-tmp-$name.png" \
  -dump-palette "/src/workplace/oracle-tmp-pal-$name.bin" > "$out/report.txt" 2>&1 || true
cp "$GOLEM/workplace/oracle-tmp-$name.png" "$out/screen.png"
cp "$GOLEM/workplace/oracle-tmp-$name.bin" "$out/screen.bin"
cp "$GOLEM/workplace/oracle-tmp-pal-$name.bin" "$out/pal.bin"

{
  echo "批次: $name"
  echo "日期: $(date -Iseconds)"
  echo "dosgolem HEAD: $head（未提交檔案 $dirty 個）"
  echo "執行檔: $EXE"
  echo "probe 旗標: $*"
  echo "GIN3.COM SHA-256: $(sha256sum "$ROOT/org_game/extracted/GIN3.COM" | cut -d' ' -f1)"
  echo "screen.bin SHA-256: $(sha256sum "$out/screen.bin" | cut -d' ' -f1)"
  echo "pal.bin SHA-256: $(sha256sum "$out/pal.bin" | cut -d' ' -f1)"
} > "$out/meta.txt"

# **旗標要能原樣拿回來。** meta.txt 那一行是給人看的：旗標裡有空格
# （`-cmdline " 9 2"`、`-poke ...=01 01 01`）時，用不帶引號的展開重跑會拆錯。
# 這裡一個旗標一行另存一份；旗標裡不會有換行，所以逐行讀是安全的。
printf '%s\n' "$@" > "$out/flags.txt"

cat "$out/meta.txt"
