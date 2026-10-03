#!/usr/bin/env bash
# tools/gui.sh xrun bash tools/list_tab_edge_gui.sh：正常同盟入口開星系清單。
set -euo pipefail
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
go build -o /tmp/logh3-list-edge ./cmd/gin3
click_at() {
  xdotool mousemove --window "$window_id" "$(( $1 * scale ))" "$(( $2 * scale ))"
  sleep 0.2
  xdotool mousedown "${3:-1}"
  sleep 0.2
  xdotool mouseup "${3:-1}"
  sleep 0.2
}
for scale in 1 2; do
  base="/src/workplace/round59-list-tab-gui-${scale}x"
  test ! -e "$base.bin"
  /tmp/logh3-list-edge -orig /orig -data /src/docs/internals/data -intro=false -scale "$scale" \
    -save /tmp/logh3-list-edge-save -shot-bin "$base.bin" -shot "$base.png" \
    -state-out "$base.txt" -shot-after 900 > "$base.log" 2>&1 &
  game_pid=$!
  window_id="$(timeout 15s xdotool search --sync --onlyvisible --all --pid "$game_pid" --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  xdotool windowmove "$window_id" 0 0
  sleep 0.5
  click_at 300 230
  for key in 1 3 8 Return; do
    xdotool keydown --window "$window_id" "$key"
    sleep 0.08
    xdotool keyup --window "$window_id" "$key"
    sleep 0.08
  done
  click_at 300 240 3
  click_at 300 240 3
  click_at 448 125
  click_at 460 125
  wait "$game_pid"
  game_pid=''
  test "$(stat -c %s "$base.bin")" = 307200
  grep -qx 'state=StateStarList' "$base.txt"
  grep -qx "window=$((640*scale))x$((480*scale))" "$base.txt"
  grep -qx 'cursor=460,125' "$base.txt"
  cmp "$base.bin" /src/workplace/round59-game-a/normal-ally/screen.bin
  echo "正常同盟星系清單：倍率$scale，307200像素與正常Game一致"
done
