#!/usr/bin/env bash
# 參謀任命三個入口的正常玩家路徑截圖（第173輪）。必須由 tools/gui.sh xrun 呼叫：
#
#   tools/gui.sh xrun ./tools/staff_pick_gui.sh
#
# 真實 X11 輸入（xdotool），不用 -script 或 direct-entry。兩陣營各走一趟：
#   月初報告 → 次月份預算計劃 → 參謀 → 第一列 → 決定 → 結束
#   → 船艦訂造報告 → 建造計劃 → 參謀 → 第一列 → 決定 → 決定（回訂造報告）
#   → 右鍵關報告到戰略地圖 → 戰略命令 → 生產計劃 → 第0列減 → 參謀 → 第一列 → 決定 → 結束
# 每一步用 ImageMagick `import` 抓視窗，輸出到 $LOGH3_STAFF_GUI_OUT（預設 workplace/staff173/gui）。
set -euo pipefail
out="${LOGH3_STAFF_GUI_OUT:-/src/workplace/staff173/gui}"
mkdir -p "$out"
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
# LOGH3_STAFF_GUI_BIN 指定發行包解開的 gin3 時，不重編，素材由執行檔旁的目錄自動尋找（發行驗證用）。
bin="${LOGH3_STAFF_GUI_BIN:-}"
res=()
if [[ -z "$bin" ]]; then
  go build -o /tmp/logh3-staff-gui ./cmd/gin3
  bin=/tmp/logh3-staff-gui
  res=(-orig /orig -data /src/docs/internals/data)
fi

click_at() { # <鍵 1/3> <x> <y>
  xdotool mousemove --window "$window_id" "$2" "$3"
  sleep 0.3
  xdotool mousedown "$1"
  sleep 0.2
  xdotool mouseup "$1"
  sleep 0.6
}
shot() { # <名字>
  import -window "$window_id" "$out/$faction-$1.png"
  echo "$faction-$1"
}

for faction in empire alliance; do
  row=210
  [[ "$faction" == alliance ]] && row=230
  "$bin" "${res[@]}" -intro=false -scale 1 \
    -save "/tmp/logh3-staff-save-$faction" > "$out/$faction.log" 2>&1 &
  game_pid=$!
  window_id="$(xdotool search --sync --onlyvisible --all --pid "$game_pid" --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  sleep 0.3
  xdotool windowmove "$window_id" 0 0
  sleep 0.3
  click_at 1 300 "$row"
  for key in 1 3 8 Return; do
    xdotool keydown --window "$window_id" "$key"
    sleep 0.2
    xdotool keyup --window "$window_id" "$key"
    sleep 0.2
  done
  sleep 0.8
  shot 00-month-report
  # 次月份預算計劃（內政專任參謀）
  click_at 1 432 340; shot 01-budget
  click_at 1 415 370; shot 02-budget-staff-list
  click_at 1 450 92;  shot 03-budget-staff-detail
  click_at 1 497 330; shot 04-budget-staff-decided
  click_at 1 463 369; shot 05-budget-end
  # 建造計劃（軍事專任參謀）
  click_at 1 448 370; shot 06-ship-order
  click_at 1 448 350; shot 07-build-plan
  click_at 1 464 370; shot 08-build-staff-list
  click_at 1 450 92;  shot 09-build-staff-detail
  click_at 1 497 330; shot 10-build-staff-decided
  click_at 1 512 370; shot 11-build-commit
  click_at 3 300 240; shot 12-rclick1
  click_at 3 300 240; shot 13-rclick2
  # 戰略命令 → 生產計劃（軍事專任參謀、七列箭頭）
  click_at 1 352 125; shot 14-strategy-menu
  click_at 1 400 185; shot 15-production
  click_at 1 232 210; shot 16-production-dec0
  click_at 1 464 370; shot 17-production-staff-list
  click_at 1 450 112; shot 18-production-staff-detail
  click_at 1 497 330; shot 19-production-staff-decided
  click_at 1 511 369; shot 20-production-end
  kill "$game_pid" 2>/dev/null || true
  wait "$game_pid" 2>/dev/null || true
  game_pid=''
done
