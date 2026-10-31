#!/usr/bin/env bash
# 171：封包／影片的共用有界工具；Go/X11基底沿r169，新增NumPy、Noto及SquashFS。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
  --memory "${LOGH3_MEM:-4g}" --cpus "${LOGH3_CPUS:-2}" --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  -v "$ROOT:/src" -v "$ROOT/org_game/extracted:/orig:ro" \
  -v "$ROOT/workplace/gocache:/gocache" -v "$ROOT/workplace/gomodcache:/gomodcache" \
  -e HOME=/tmp -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e GOPROXY=off -e GOSUMDB=off \
  -w /src "${LOGH3_RELEASE_IMAGE:-logh3-release:20261031-r171}" "$@"
