#!/usr/bin/env bash
# 只重跑 tools/fleetcmd_receipts.sh 的一批或一段，不用整條鏈從冷啟動跑：
#
#   tools/fleetcmd_rerun.sh <批次名>     只跑那一行 rec（含反斜線續行）
#   tools/fleetcmd_rerun.sh -s <段號>    跑整段（從 "# N." 標題到下一個標題）
#
# 標頭的變數（DS、HOLD、STOCK、ADM、TABS…）照主腳本的定義取；rec 需要的
# 上一個檢查點（workplace/fcmd/<from>.state）要先存在——沒有就先跑它的上一批。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
MAIN=tools/fleetcmd_receipts.sh
eval "$(sed -n '/^set -euo/,/^cd "\$ROOT"/p' "$MAIN" | grep -v '^set -euo\|^ROOT=\|^cd ')"
if [[ "${1:-}" == -s ]]; then
  sec=$2
  chunk=$(awk -v s="$sec" '
    $0 ~ "^# "s"\\. " {p=1; next}
    p && $0 ~ /^# [0-9]+\. / {exit}
    p {print}' "$MAIN")
  [[ -n "$chunk" ]] || { echo "找不到第 $sec 段" >&2; exit 1; }
  eval "$chunk"
else
  name=${1:?用法：fleetcmd_rerun.sh <批次名> | -s <段號>}
  line=$(awk -v n="$name" '
    $0 ~ "^rec "n" " {p=1}
    p {print; if ($0 !~ /\\$/) {exit}}' "$MAIN")
  [[ -n "$line" ]] || { echo "找不到批次 $name" >&2; exit 1; }
  eval "$line"
fi
