#!/usr/bin/env bash
# 於既有影片 Docker 內錄製完整 4:3 遊戲視窗。
set -euo pipefail
Xvfb :98 -screen 0 "${LOGH3_RECORD_SIZE:-1280x960}x24" -nolisten tcp >/tmp/s1-xvfb.log 2>&1 &
s1_display_pid=$!
trap 'kill "$s1_display_pid" 2>/dev/null || true; wait "$s1_display_pid" 2>/dev/null || true' EXIT
export DISPLAY=:98 LIBGL_ALWAYS_SOFTWARE=1 XDG_RUNTIME_DIR=/tmp
for ((n=0; n<50; n++)); do
    if xdotool getdisplaygeometry >/dev/null 2>&1; then break; fi
    sleep 0.1
done
xdotool getdisplaygeometry >/dev/null
"$@"
