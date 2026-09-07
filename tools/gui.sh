#!/usr/bin/env bash
# Ebiten（GUI）走 docker：建置要 X11／OpenGL 的標頭，無頭跑要 Xvfb。
#
#   tools/gui.sh build ./cmd/gin3          # 建置
#   tools/gui.sh run ./cmd/gin3 -orig /orig -shot /src/workplace/gui.png
#   LOGH3_NET=1 tools/gui.sh mod download  # 抓相依（唯一需要網路的一步）
#
# 預設 `--network none`；只有抓相依時開網路（`LOGH3_NET=1`）。
# 螢幕是 Xvfb ＋ Mesa 的軟體 OpenGL，所以沒有顯示卡也跑得動。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_GUI_IMAGE:-logh3-ebiten:1}"
mkdir -p "$ROOT/workplace/gocache" "$ROOT/workplace/gomodcache"
NET=(--network none)
[[ "${LOGH3_NET:-0}" == 1 ]] && NET=()
MOUNTS=()
[[ -d "$ROOT/org_game/extracted" ]] && MOUNTS+=(-v "$ROOT/org_game/extracted:/orig:ro")
PRE=()
case "${1:-}" in
  run|test) PRE=(with-xvfb) ;;
esac
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm "${NET[@]}" \
  --memory "${LOGH3_MEM:-4g}" --cpus "${LOGH3_CPUS:-4}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$ROOT:/src" \
  -v "$ROOT/workplace/gocache:/gocache" -v "$ROOT/workplace/gomodcache:/gomodcache" \
  -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e HOME=/tmp -e GOFLAGS=-mod=mod \
  -e LIBGL_ALWAYS_SOFTWARE=1 -e XDG_RUNTIME_DIR=/tmp \
  "${MOUNTS[@]}" -w /src "$IMAGE" "${PRE[@]}" go "$@"
