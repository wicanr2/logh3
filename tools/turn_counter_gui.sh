#!/usr/bin/env bash
# tools/gui.sh xrun bash tools/turn_counter_gui.sh，正常主選單與S3首戰。
set -euo pipefail
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
go build -o /tmp/logh3-turn-counter ./cmd/gin3
click_at() {
  xdotool mousemove --window "$window_id" "$(( $1 * scale ))" "$(( $2 * scale ))"
  sleep 0.2
  xdotool mousedown 1
  sleep 0.2
  xdotool mouseup 1
  sleep 0.2
}
for scale in 1 2; do
  base="/src/workplace/${LOGH3_TURN_GUI_PREFIX:-round47-gui-pass}-${scale}x"
  test ! -e "$base.bin"
  /tmp/logh3-turn-counter -orig /orig -data /src/docs/internals/data -intro=false -scale "$scale" \
    -save /tmp/logh3-turn-counter-save -shot-bin "$base.bin" -shot "$base.png" \
    -state-out "$base.txt" -shot-after 1200 > "$base.log" 2>&1 &
  game_pid=$!
  window_id="$(xdotool search --sync --onlyvisible --all --pid "$game_pid" --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  sleep 0.3
  # 使用產品-scale啟動，並把命令落點保持在Xvfb內。
  xdotool windowmove "$window_id" 0 0
  sleep 0.2
  click_at 300 270
  for key in 1 3 8 Return; do
    xdotool keydown --window "$window_id" "$key"
    sleep 0.08
    xdotool keyup --window "$window_id" "$key"
    sleep 0.08
  done
  click_at 300 322
  click_at 318 208
  click_at 544 125
  click_at 560 125
  xdotool mousemove --window "$window_id" "$((240 * scale))" "$((240 * scale))"
  wait "$game_pid"
  game_pid=''
  test "$(stat -c %s "$base.bin")" = 307200
  grep -qx 'state=StateBattle' "$base.txt"
  grep -qx "window=$((640 * scale))x$((480 * scale))" "$base.txt"
  grep -qx 'cursor=240,240' "$base.txt"
  grep -qx 'cursor_mode=1' "$base.txt"
  cmp "$base.bin" "/src/workplace/${LOGH3_TURN_PLAYER_DIR:-round47-normal-a}/entry/screen.bin"
  echo "正常S3首戰：倍率$scale，整張正式畫面與正常Game收據相同"
done
