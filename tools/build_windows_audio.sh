#!/usr/bin/env bash
# 固定Wine9父映像及全部新增套件；只補ALSA/Pulse缺件，入口playtest/093。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test -f "$ROOT/tools/Dockerfile.windows-audio"
test -f "$ROOT/tools/windows-audio-packages.lock"
[[ "$(docker image inspect colonization-windows-verification:20261001-r1 --format '{{.Id}}')" == \
   sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201 ]]
exec timeout 10m docker build --memory 2g --cpu-period 100000 --cpu-quota 200000 \
    --network default -f "$ROOT/tools/Dockerfile.windows-audio" \
    -t logh3-windows-audio:1 "$ROOT/tools"
