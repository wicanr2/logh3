#!/usr/bin/env bash
# #69 Windows正式播放器；連至既有音訊image提供的隔離Unix socket。
set -euo pipefail
[[ $# -ge 2 && "$1" =~ ^[A-Za-z0-9][A-Za-z0-9_-]*$ ]] || exit 2
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BATCH="$1"
shift
ENV_ARGS=()
[[ -z "${LOGH3_MESA_THREADS:-}" ]] || ENV_ARGS+=(-e "LP_NUM_THREADS=$LOGH3_MESA_THREADS")
DRIVER="${LOGH3_WINE_AUDIO_DRIVER:-pulse}"
[[ "$DRIVER" == pulse || "$DRIVER" == alsa ]] || exit 2
IMAGE=colonization-windows-verification:20261001-r1
PREFIX=/out/wine-prefix
if [[ "$DRIVER" == alsa ]]; then
    IMAGE=logh3-windows-audio:1
    PREFIX="/out/$BATCH/wine-prefix"
    ENV_ARGS+=(-e "ALSA_CONFIG_PATH=/out/$BATCH/alsa.conf")
fi
for dir in workplace/music69-windows workplace/music69-windows/combined-utf8-full workplace/music69-windows/source103-player org_game/extracted tools docs/internals/data; do
    test -d "$ROOT/$dir"
done
test -f "$ROOT/workplace/music69-windows/$BATCH/READY.json"
test -S "$ROOT/workplace/music69-windows/$BATCH/pulse/native"
[[ "$(stat -c '%u:%g' "$ROOT/workplace/music69-windows/$BATCH")" == "$(id -u):$(id -g)" ]]
exec timeout "${LOGH3_TIMEOUT:-3m}" docker run --rm --network none \
    --memory 4g --cpus "${LOGH3_WINE_CPUS:-2}" --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" --entrypoint sh \
    -v "$ROOT/workplace/music69-windows:/out" \
    -v "$ROOT/workplace/music69-windows/combined-utf8-full:/music:ro" \
    -v "$ROOT/workplace/music69-windows/source103-player:/player:ro" \
    -v "$ROOT/org_game/extracted:/orig:ro" \
    -v "$ROOT/docs/internals/data:/data:ro" \
    -v "$ROOT/tools:/tools:ro" \
    -e HOME=/out/wine-home -e WINEPREFIX="$PREFIX" -e WINEARCH=win64 \
    -e LOGH3_WINE_AUDIO_DRIVER="$DRIVER" -e LOGH3_WINE_AUDIO_BATCH="$BATCH" \
    -e PULSE_SERVER="unix:/out/$BATCH/pulse/native" \
    "${ENV_ARGS[@]}" \
    -e WINEDEBUG=-all -e DISPLAY=:98 -e LANG=C.UTF-8 -e LC_ALL=C.UTF-8 -w /out \
    "$IMAGE" \
    /tools/music_wine_entry.sh python3 /tools/scripts/wine_audio_driver.py "$@"
