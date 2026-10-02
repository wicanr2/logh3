#!/usr/bin/env bash
# 必須由tools/gui.sh xrun呼叫；088正常X11輸入，不用-script或direct-entry。
set -euo pipefail
game_pid=''
trap 'if [[ -n "$game_pid" ]]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi' EXIT
go build -o /tmp/logh3-load-seats ./cmd/gin3
click_at() {
  xdotool mousemove --window "$window_id" "$(( $2 * scale ))" "$(( $3 * scale ))"
  sleep 0.08
  xdotool mousedown "$1"
  sleep 0.08
  xdotool mouseup "$1"
  sleep 0.08
}
for name in hover empire alliance versus empire2x; do
  base="/src/workplace/round46-gui-$name"
  test ! -e "$base.bin"
  scale=1
  [[ "$name" == empire2x ]] && scale=2
  /tmp/logh3-load-seats -orig /orig -data /src/docs/internals/data -intro=false \
    -save /tmp/logh3-load-seats-save -shot-bin "$base.bin" -shot "$base.png" \
    -state-out "$base.txt" -shot-after 420 > "$base.log" 2>&1 &
  game_pid=$!
  window_id="$(xdotool search --sync --onlyvisible --name '^銀河英雄傳說III SP$' | head -1)"
  xdotool windowfocus "$window_id"
  sleep 0.3
  xdotool windowsize "$window_id" "$((640 * scale))" "$((480 * scale))"
  sleep 0.2
  click_at 1 300 270
  for key in 1 3 8 Return; do
    xdotool keydown --window "$window_id" "$key"
    sleep 0.08
    xdotool keyup --window "$window_id" "$key"
    sleep 0.08
  done
  click_at 1 300 322
  case "$name" in
    hover) click_at 3 318 228; state=StateLoadFaction; cursor=318,228 ;;
    empire|empire2x) click_at 1 318 208; state=StateCommandPhase; cursor=318,208 ;;
    alliance) click_at 1 318 228; state=StateCommandPhase; cursor=318,228 ;;
    versus) click_at 1 318 248; state=StateVersusInput; cursor=318,248 ;;
  esac
  wait "$game_pid"
  game_pid=''
  test "$(stat -c %s "$base.bin")" = 307200
  grep -qx "state=$state" "$base.txt"
  grep -qx "window=$((640 * scale))x$((480 * scale))" "$base.txt"
  grep -qx "cursor=$cursor" "$base.txt"
  grep -qx 'cursor_mode=1' "$base.txt"
  echo "$name: $state，游標$cursor，倍率$scale通過"
done
