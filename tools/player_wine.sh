#!/usr/bin/env bash
# #69 Windows 玩家入口隔離驗證；沿用音樂工具的 Wine image 與有界 Xvfb。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for dir in workplace/music69-windows org_game/extracted tools docs/internals/data; do
    test -d "$ROOT/$dir"
done
[[ "$(stat -c '%u:%g' "$ROOT/workplace/music69-windows")" == "$(id -u):$(id -g)" ]]
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
    --memory 4g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" --entrypoint sh \
    -v "$ROOT/workplace/music69-windows:/out" \
    -v "$ROOT/org_game/extracted:/orig:ro" \
    -v "$ROOT/docs/internals/data:/data:ro" \
    -v "$ROOT/tools:/tools:ro" \
    -e HOME=/out/wine-home -e WINEPREFIX=/out/wine-prefix -e WINEARCH=win64 \
    -e WINEDEBUG=-all -e DISPLAY=:98 -e LANG=C.UTF-8 -e LC_ALL=C.UTF-8 -w /out \
    "${LOGH3_WINE_IMAGE:-kol-wine-verify:ubuntu-noble-20261007-r1}" \
    /tools/music_wine_entry.sh "$@"
