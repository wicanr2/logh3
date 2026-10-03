#!/usr/bin/env bash
# tools/gui.sh xrun bash tools/left_hold_gui.sh：正常鍵鼠保持到資料窗上界26。
set -euo pipefail
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
go build -o /tmp/logh3-left-hold ./cmd/gin3
click_at() {
  xdotool mousemove --window "$window_id" "$(( $1 * scale ))" "$(( $2 * scale ))"
  sleep 0.2
  xdotool mousedown 1
  sleep 0.2
  xdotool mouseup 1
  sleep 0.2
}
for scale in 1 2; do
  base="${LOGH3_LEFT_HOLD_GUI_PREFIX:-/src/workplace/round58-left-hold-gui-final}-${scale}x"
  test ! -e "$base.bin"
  /tmp/logh3-left-hold -orig /orig -data /src/docs/internals/data -intro=false -scale "$scale" \
    -save /tmp/logh3-left-hold-save -shot-bin "$base.bin" -shot "$base.png" \
    -state-out "$base.txt" -shot-after 4800 > "$base.log" 2>&1 &
  game_pid=$!
  window_id="$(xdotool search --sync --onlyvisible --all --pid "$game_pid" --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  xdotool windowmove "$window_id" 0 0
  sleep 0.5
  click_at 300 270
  for key in 1 3 8 Return; do
    xdotool keydown --window "$window_id" "$key"
    sleep 0.08
    xdotool keyup --window "$window_id" "$key"
    sleep 0.08
  done
  click_at 300 322
  click_at 318 228
  click_at 544 125
  click_at 560 125
  python3 tools/scripts/left_hold_gui_ready.py "$window_id" "$scale" \
    /src/workplace/round58-hold-player-b/battle-ready
  click_at 400 80
  click_at 400 100
  xdotool mousemove --window "$window_id" "$((440*scale))" "$((390*scale))"
  sleep 0.2
  xdotool mousedown 1
  sleep 5
  xdotool mouseup 1
  wait "$game_pid"
  game_pid=''
  test "$(stat -c %s "$base.bin")" = 307200
  grep -qx 'state=StateBattle' "$base.txt"
  grep -qx "window=$((640*scale))x$((480*scale))" "$base.txt"
  grep -qx 'cursor=440,390' "$base.txt"
  cmp "$base.bin" /src/workplace/round58-hold-player-b/last/screen.bin
  echo "正常鍵鼠保持到26：倍率$scale，307200像素與正常Game／原版一致"
done
