#!/usr/bin/env bash
# 在既有GUI／playback容器內使用完整容納2倍畫面的有界Xvfb。
set -euo pipefail
[[ $# -gt 0 ]]
Xvfb :98 -screen 0 1600x1200x24 -nolisten tcp >/tmp/input121-xvfb.log 2>&1 &
input_display_pid=$!
trap 'kill "$input_display_pid" 2>/dev/null || true; wait "$input_display_pid" 2>/dev/null || true' EXIT
export DISPLAY=:98
for ((n=0; n<50; n++)); do
    if xset q >/dev/null 2>&1; then break; fi
    sleep 0.1
done
xset q >/dev/null
"$@"
