#!/usr/bin/env bash
# #69 Windows 原生音樂工具的隔離實跑；入口 docs/re/362-windows-music-tools.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test -d "$ROOT/workplace/music69-windows"
test -d "$ROOT/org_game/extracted"
test -d "$ROOT/tools"
[[ "$(stat -c '%u:%g' "$ROOT/workplace/music69-windows")" == "$(id -u):$(id -g)" ]]
exec timeout "${LOGH3_TIMEOUT:-25m}" docker run --rm --network none \
    --memory 4g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" --entrypoint sh \
    -v "$ROOT/workplace/music69-windows:/out" \
    -v "$ROOT/org_game/extracted:/orig:ro" \
    -v "$ROOT/tools:/tools:ro" \
    -e HOME=/out/wine-home -e WINEPREFIX=/out/wine-prefix -e WINEARCH=win64 \
    -e WINEDEBUG=-all -e DISPLAY=:98 -e LANG=C.UTF-8 -e LC_ALL=C.UTF-8 -w /out \
    colonization-windows-verification:20261001-r1 \
    /tools/music_wine_entry.sh "$@"
