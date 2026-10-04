#!/usr/bin/env bash
# 來源、權利及重現契約：docs/playtest/067-fleet-sound.md。
# 輸出為 workplace 下尚未存在的批次名稱。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ $# == 1 && "$1" =~ ^[A-Za-z0-9][A-Za-z0-9_-]*$ ]] || {
  echo '用法：bash tools/sound_to_ogg.sh <新批次名稱>' >&2; exit 2;
}
[[ -f "$ROOT/org_game/extracted/GIN3SP.VOC" && -d "$ROOT/workplace" && ! -e "$ROOT/workplace/$1" ]] || {
  echo '缺原版或輸出已存在，停止。' >&2; exit 2;
}
[[ "$(stat -c %u "$ROOT/workplace")" == "$(id -u)" ]] || exit 2
cd "$ROOT"
bash tools/gui.sh build -o workplace/fleet-sound-convert tools/scripts/fleet_sound_convert.go
bash tools/audio.sh /src/workplace/fleet-sound-convert -out "/work/$1"
