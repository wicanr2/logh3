#!/usr/bin/env bash
# Ebiten 顯示層的同狀態驗收：走玩家路徑，把 Ebiten 畫出來的色號陣列
# 與原版收據逐位元組比。
#
#   tools/gui_check.sh
#
# **這是顯示層唯一算數的驗收**：視窗開得起來、看起來像，都不算。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

cases=(
  "r0|move 300 210"
  "cb1|move 300 210; click"
  "key1|move 300 210; click; type 138; enter"
  "cbAlly|move 300 230; click"
  "keyAlly|move 300 230; click; type 138; enter"
  "main1e|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340"
  "load1e|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145"
  "vs3|move 300 250; click; type 126; enter"
  "vs5|move 300 250; click; type 126; enter; click 432 263; rclick 300 240; rclick 300 240; click 544 125; click 560 125"
)
fail=0
for c in "${cases[@]}"; do
  batch="${c%%|*}"; script="${c#*|}"
  oracle="workplace/oracle/$batch/screen.bin"
  if [[ ! -f "$oracle" ]]; then
    echo "  $batch 缺收據，跳過"
    continue
  fi
  tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data \
    -script "$script" -shot-bin /src/workplace/gui-check.bin -shot-after 2 \
    >/dev/null 2>&1
  if cmp -s workplace/gui-check.bin "$oracle"; then
    echo "  $batch 逐位元組相同"
  else
    echo "  $batch **不同**"
    fail=1
  fi
done
exit $fail
