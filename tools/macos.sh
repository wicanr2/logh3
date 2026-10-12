#!/usr/bin/env bash
# #94：既有Go1.24.13／osxcross的有界入口，來源與限制見docs/playtest/106。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE=psychicwar-osxcross:go1.24.13-15.5-r2
EXPECTED=sha256:d4c50a5cf3cf35164b83fb1531fe66f52b9224e826944363bf1bc4d7ba3c8d0c
batch="${1:?需建置批次}"
shift
[[ "$batch" =~ ^[a-z0-9-]+$ ]]
test -d "$ROOT/workplace"
test -d "$ROOT/tools"
test "$(docker image inspect "$IMAGE" --format '{{.Id}}')" = "$EXPECTED"
output="$ROOT/workplace/$batch"
if [[ ! -e "$output" ]]; then
    mkdir "$output"
fi
test -d "$output"
test ! -L "$output"
test "$(stat -c '%u:%g' "$output")" = "$(id -u):$(id -g)"
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
    --memory 4g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
    -v "$ROOT:/src:ro" -v "$output:/out" \
    -e HOME=/tmp -e GOCACHE=/out/gocache -e GOPROXY=off -e GOSUMDB=off \
    -e GOTOOLCHAIN=local -w /out "$IMAGE" "$@"
