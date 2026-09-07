#!/bin/sh
# 在虛擬顯示下執行命令。Ebiten 在 package init 就會開 GLFW，沒有 DISPLAY
# 會直接 panic，所以需要顯示的 package 一律透過這支執行。
#
# 不用 xvfb-run：它在 -u <uid>:<gid> 且該 uid 不在 /etc/passwd 時會無限等待。
set -e
Xvfb :99 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
XVFB_PID=$!
DISPLAY=:99
export DISPLAY
# 等 X server 接受連線；xdpyinfo 不一定裝，改用一次性 X 連線測試。
i=0
while [ $i -lt 50 ]; do
    if xset q >/dev/null 2>&1; then break; fi
    i=$((i + 1))
    sleep 0.2
done
if ! xset q >/dev/null 2>&1; then
    echo "with-xvfb：X server 未就緒" >&2
    tail -5 /tmp/xvfb.log >&2 || true
    kill "$XVFB_PID" 2>/dev/null || true
    exit 1
fi
"$@"
STATUS=$?
kill "$XVFB_PID" 2>/dev/null || true
exit $STATUS
