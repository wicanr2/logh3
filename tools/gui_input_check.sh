#!/usr/bin/env bash
# 真正走 X11 事件的 GUI 黑箱驗收：不使用 -script，讓 xdotool 把實際滑鼠
# 座標與按鍵送進 Ebiten。驗證第二層選單、縮放後命中與雙游標修正。
set -euo pipefail

go build -o /tmp/logh3-gui-input ./cmd/gin3

click_at() {
  window_id="$1"
  button="$2"
  x="$3"
  y="$4"
  xdotool mousemove --window "$window_id" "$x" "$y"
  sleep 0.08
  xdotool mousedown "$button"
  sleep 0.08
  xdotool mouseup "$button"
  sleep 0.08
}

key_at() {
  window_id="$1"
  key="$2"
  xdotool keydown --window "$window_id" "$key"
  sleep 0.08
  xdotool keyup --window "$window_id" "$key"
  sleep 0.08
}

run_case() {
  name="$1"
  width="$2"
  height="$3"
  main_x="$4"
  main_y="$5"
  report_x="$6"
  report_y="$7"
  menu_x="$8"
  menu_y="$9"
  load_x="${10}"
  load_y="${11}"
  want_cursor="${12}"
  receipt="/tmp/logh3-gui-${name}.txt"
  pixels="/tmp/logh3-gui-${name}.bin"

  /tmp/logh3-gui-input -orig /orig -data /src/docs/internals/data \
    -intro=false -shot-bin "$pixels" -state-out "$receipt" -shot-after 240 &
  game_pid=$!
  window_id="$(xdotool search --sync --onlyvisible --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  # 標題出現時 GLFW 還可能在套用啟動尺寸；等第一批更新格後再模擬使用者拖曳。
  sleep 0.3
  xdotool windowsize "$window_id" "$width" "$height"
  sleep 0.2

  # 主選單 → 防拷 → 月報兩層 → 系統指令 → 第二項「取進度」。
  click_at "$window_id" 1 "$main_x" "$main_y"
  key_at "$window_id" 1
  key_at "$window_id" 3
  key_at "$window_id" 8
  key_at "$window_id" Return
  click_at "$window_id" 3 "$report_x" "$report_y"
  click_at "$window_id" 3 "$report_x" "$report_y"
  click_at "$window_id" 1 "$menu_x" "$menu_y"
  click_at "$window_id" 1 "$load_x" "$load_y"

  wait "$game_pid"
  test "$(stat -c %s "$pixels")" = 307200
  grep -qx 'state=StateLoadGame' "$receipt" || {
    echo "  ${name}: 未進入取進度畫面；實際收據：" >&2
    cat "$receipt" >&2
    return 1
  }
  grep -qx "window=${width}x${height}" "$receipt" || {
    cat "$receipt" >&2
    return 1
  }
  grep -qx "cursor=${want_cursor}" "$receipt" || {
    cat "$receipt" >&2
    return 1
  }
  # Ebiten v2.8.8 的 CursorModeHidden 數值是 1。
  grep -qx 'cursor_mode=1' "$receipt" || {
    cat "$receipt" >&2
    return 1
  }
  echo "  ${name}: 第二層選單、${width}x${height} 視窗、邏輯游標 ${want_cursor} 通過"
}

run_fleet_case() {
  receipt=/tmp/logh3-gui-fleet.txt
  pixels=/tmp/logh3-gui-fleet.bin
  /tmp/logh3-gui-input -orig /orig -data /src/docs/internals/data \
    -intro=false -shot-bin "$pixels" -state-out "$receipt" -shot-after 480 &
  game_pid=$!
  window_id="$(xdotool search --sync --onlyvisible --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  sleep 0.3

  click_at "$window_id" 1 300 210
  for key in 1 3 8 Return; do key_at "$window_id" "$key"; done
  click_at "$window_id" 3 300 240
  click_at "$window_id" 3 300 240
  click_at "$window_id" 1 352 125 # 戰略命令
  click_at "$window_id" 1 400 125 # 出擊命令
  click_at "$window_id" 1 400 85  # 待命提督第一列
  click_at "$window_id" 1 488 330 # 提督詳細資料「出擊」
  click_at "$window_id" 1 368 350 # 艦隊編成「艦隊編號」

  wait "$game_pid"
  test "$(stat -c %s "$pixels")" = 307200
  grep -qx 'state=StateFleetNumber' "$receipt" || {
    echo '  fleet: 艦隊編成的下一層清單未開啟；實際收據：' >&2
    cat "$receipt" >&2
    return 1
  }
  grep -qx 'cursor=368,350' "$receipt"
  grep -qx 'cursor_mode=1' "$receipt"
  echo '  fleet: 正常玩家路徑可點入艦隊編號下一層清單'
}

run_startup_scale_case() {
  receipt=/tmp/logh3-gui-scale2.txt
  pixels=/tmp/logh3-gui-scale2.bin
  /tmp/logh3-gui-input -orig /orig -data /src/docs/internals/data \
    -intro=false -scale 2 -shot-bin "$pixels" -state-out "$receipt" -shot-after 60 &
  game_pid=$!
  xdotool search --sync --onlyvisible --name '^銀河英雄傳說III SP$' >/dev/null
  wait "$game_pid"
  grep -qx 'window=1280x960' "$receipt" || {
    echo "  scale2: -scale 2 未建立 1280x960 視窗；實際收據：" >&2
    cat "$receipt" >&2
    return 1
  }
  echo "  scale2: -scale 2 啟動尺寸 1280x960 通過"
}

# 1×：X11 座標就是 640×480 邏輯座標。
run_case 1x 640 480 300 210 300 240 544 125 560 145 560,145

# 1.5×：同一組邏輯點改送等比例的實際視窗座標，驗證 Ebiten 的縮放映射。
run_case 1_5x 960 720 450 315 450 360 816 188 840 218 560,145

# 命令列的穩定入口也要能直接用整數倍率啟動。
run_startup_scale_case

# 另一族的巢狀介面：戰略命令 → 出擊命令 → 提督 → 艦隊編成 → 艦隊編號。
run_fleet_case
