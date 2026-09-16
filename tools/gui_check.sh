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
  # vs3 收據在交棒視窗出現後還點了 (400,262)；那點在確認框內、熱區外，
  # 狀態不變但游標會搬過去（docs/re/280）。漏掉就只差兩個游標位置 274 px。
  "vs3|move 300 250; click; type 126; enter; click 400 262"
  "vs4|move 300 250; click; type 126; enter; click 432 263"
  "vs5|move 300 250; click; type 126; enter; click 432 263; rclick 300 240; rclick 300 240; click 544 125; click 560 125"
  # 戰略艦隊命令（docs/spec/051）：dec1 的序列再點右下角清單第 0 列開艦隊視窗；
  # fcmd-done 走完「作戰 → 移動 → 亞爾提那 → 決定」；fcmd-exec 再按 系統指令 → 實行
  # （艦隊走到亞爾提那、那一格變黑、首都經濟力 82339）。
  "fcmd-win|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308"
  "fcmd-done|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263"
  "fcmd-exec|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 312; click 85 128; click 432 263; click 544 125; click 560 125"
  # 待命 → 「指定待命行星」再點布連娜（docs/re/314）。排除小圖最後一列
  # （x 152–279、y 347）：原版那一列讀到地圖緩衝區外（規格 051 已知差異）。
  "fcmd-standby-pick|move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308; click 490 345; click 497 392; click 350 250|152,347,128,1"
)
fail=0
for c in "${cases[@]}"; do
  IFS='|' read -r batch script except <<<"$c"
  oracle="workplace/oracle/$batch/screen.bin"
  if [[ ! -f "$oracle" ]]; then
    echo "  $batch 缺收據，跳過"
    continue
  fi
  tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data \
    -script "$script" -shot-bin /src/workplace/gui-check.bin -shot-after 2 \
    >/dev/null 2>&1
  if [[ -n "$except" ]]; then
    # 有排除矩形的用 python 比（x,y,w,h 之外要逐位元組相同）。
    if tools/py.sh tools/scripts/cmp_except.py workplace/gui-check.bin "$oracle" "$except"; then
      echo "  $batch 排除 ($except) 之外逐位元組相同"
    else
      echo "  $batch **不同**"
      fail=1
    fi
  elif cmp -s workplace/gui-check.bin "$oracle"; then
    echo "  $batch 逐位元組相同"
  else
    echo "  $batch **不同**"
    fail=1
  fi
done
exit $fail
