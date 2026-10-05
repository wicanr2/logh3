#!/usr/bin/env bash
# #69：唯讀 Windows 玩家目錄，正常 GUI 保存及重新啟動讀回。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for dir in workplace/music69-windows workplace/music69-windows/saveload69-r106-build org_game/extracted tools docs/internals/data; do
    test -d "$ROOT/$dir"
done
[[ "$(stat -c '%u:%g' "$ROOT/workplace/music69-windows")" == "$(id -u):$(id -g)" ]]
exec timeout "${LOGH3_TIMEOUT:-8m}" docker run --rm --network none \
    --memory 4g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" --entrypoint sh \
    -v "$ROOT/workplace/music69-windows:/out" \
    -v "$ROOT/workplace/music69-windows/saveload69-r106-build:/player:ro" \
    -v "$ROOT/org_game/extracted:/orig:ro" \
    -v "$ROOT/docs/internals/data:/data:ro" \
    -v "$ROOT/tools:/tools:ro" \
    -e HOME=/out/wine-home -e WINEPREFIX=/out/wine-prefix -e WINEARCH=win64 \
    -e WINEDEBUG=-all -e DISPLAY=:98 -e LANG=C.UTF-8 -e LC_ALL=C.UTF-8 \
    -e LP_NUM_THREADS=1 -w /player \
    colonization-windows-verification:20261001-r1 \
    /tools/music_wine_entry.sh python3 /tools/scripts/windows_save_load.py "$@"
