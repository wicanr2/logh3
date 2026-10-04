#!/usr/bin/env bash
# 固定GUI3與虛擬PulseAudio，來源 tools/Dockerfile.gui-playback。
# bash tools/playback.sh <新批次名稱> <容器內命令與參數...>
set -euo pipefail
[[ $# -ge 2 && "$1" =~ ^[A-Za-z0-9][A-Za-z0-9_-]*$ ]] || exit 2
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ -d "$ROOT/workplace" && ! -e "$ROOT/workplace/$1" ]] || exit 2
[[ "$(stat -c %u "$ROOT/workplace")" == "$(id -u)" ]] || exit 2
export LOGH3_GUI_IMAGE=logh3-ebiten:3
exec bash "$ROOT/tools/gui.sh" xrun python3 tools/scripts/playback_sink.py "$@"
