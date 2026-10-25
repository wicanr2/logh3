#!/bin/sh
# 有界 Xvfb／Wine 擁有程序；只供 Docker 包裝器，結束即關閉背景程序。
set -eu
mkdir -p "$HOME"
Xvfb :98 -screen 0 "${LOGH3_WINE_SCREEN:-1024x768x24}" -nolisten tcp > /out/xvfb.log 2>&1 &
xvfb_pid=$!
cleanup() {
    wineserver -k >/dev/null 2>&1 || true
    kill "$xvfb_pid" >/dev/null 2>&1 || true
    wait "$xvfb_pid" >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM
"$@"
