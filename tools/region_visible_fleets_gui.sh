#!/usr/bin/env bash
# tools/gui.sh xrun bash tools/region_visible_fleets_gui.sh
# 正常帝國新遊戲→編成→實行一回合→一覽表，真實X11鍵鼠。
set -euo pipefail
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
go build -o /tmp/logh3-visible-fleets ./cmd/gin3
click_at() {
  xdotool mousemove --window "$window_id" "$(( $1 * scale ))" "$(( $2 * scale ))"
  sleep 0.2
  xdotool mousedown "${3:-1}"
  sleep 0.2
  xdotool mouseup "${3:-1}"
  sleep 0.2
}
for scale in 1 2; do
  base="/src/workplace/round60-visible-gui-final-${scale}x"
  test ! -e "$base.bin"
  /tmp/logh3-visible-fleets -orig /orig -data /src/docs/internals/data -intro=false -scale "$scale" \
    -save /tmp/logh3-visible-fleets-save -shot-bin "$base.bin" -shot "$base.png" \
    -state-out "$base.txt" -shot-after 2200 > "$base.log" 2>&1 &
  game_pid=$!
  window_id="$(timeout 15s xdotool search --sync --onlyvisible --all --pid "$game_pid" --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  xdotool windowmove "$window_id" 0 0
  sleep 0.5
  click_at 300 210
  for key in 1 3 8 Return; do
    xdotool keydown --window "$window_id" "$key"
    sleep 0.08
    xdotool keyup --window "$window_id" "$key"
    sleep 0.08
  done
  click_at 300 240 3
  click_at 300 240 3
  xdotool mousemove --window "$window_id" "$((600*scale))" "$((340*scale))"
  sleep 0.2
  click_at 352 125
  click_at 400 125
  click_at 400 85
  click_at 488 330
  click_at 480 350
  click_at 510 350
  click_at 544 125
  click_at 560 125
  sleep 3
  click_at 448 125
  click_at 460 125
  wait "$game_pid"
  game_pid=''
  test "$(stat -c %s "$base.bin")" = 307200
  grep -qx 'state=StateStarList' "$base.txt"
  grep -qx "window=$((640*scale))x$((480*scale))" "$base.txt"
  grep -qx 'cursor=460,125' "$base.txt"
  cmp "$base.bin" /src/workplace/round60-visible-game-final-b/screen.bin
  echo "正常帝國一回合後星系清單：倍率$scale，307200像素與原版／Game一致"
done
