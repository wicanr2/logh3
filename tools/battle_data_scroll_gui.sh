#!/usr/bin/env bash
# tools/gui.sh xrun bash tools/battle_data_scroll_gui.sh：正常S3同盟資料表。
set -euo pipefail
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
go build -o /tmp/logh3-data-scroll ./cmd/gin3
click_at() {
  xdotool mousemove --window "$window_id" "$(( $1 * scale ))" "$(( $2 * scale ))"
  sleep 0.2
  xdotool mousedown 1
  sleep 0.2
  xdotool mouseup 1
  sleep 0.2
}
for scale in 1 2; do
  base="/src/workplace/${LOGH3_DATA_GUI_PREFIX:-round57-data-gui}-${scale}x"
  test ! -e "$base.bin"
  /tmp/logh3-data-scroll -orig /orig -data /src/docs/internals/data -intro=false -scale "$scale" \
    -save /tmp/logh3-data-scroll-save -shot-bin "$base.bin" -shot "$base.png" \
    -state-out "$base.txt" -shot-after 2400 > "$base.log" 2>&1 &
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
  sleep 17
  click_at 400 80
  click_at 400 100
  click_at 440 390
  wait "$game_pid"
  game_pid=''
  test "$(stat -c %s "$base.bin")" = 307200
  grep -qx 'state=StateBattle' "$base.txt"
  grep -qx "window=$((640 * scale))x$((480 * scale))" "$base.txt"
  grep -qx 'cursor=440,390' "$base.txt"
  cmp "$base.bin" "/src/workplace/${LOGH3_DATA_PLAYER_DIR:-round57-player-a}/one/screen.bin"
  echo "正常鍵鼠S3同盟資料偏移1：倍率$scale，完整畫面與正常Game收據相同"
done
