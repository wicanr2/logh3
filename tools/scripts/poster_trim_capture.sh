#!/usr/bin/env bash
# 154：只擷取可丟棄HTML效果稿；掛載來源不存在時直接失敗。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BATCH="$ROOT/workplace/poster154"
variant="${1:?需A或B}"
[[ "$variant" == A || "$variant" == B ]]
test -d "$BATCH"
test -f "$BATCH/$variant.html"
test "$(stat -c '%u:%g' "$BATCH")" = "$(id -u):$(id -g)"
exec timeout 2m docker run --rm --network none --memory 1g --cpus 2 --pids-limit 256 \
  --log-opt max-size=5m -u "$(id -u):$(id -g)" \
  --mount "type=bind,source=$BATCH,target=/work" -e FONTCONFIG_FILE=/work/fonts.conf \
  --entrypoint sh mcr.microsoft.com/playwright:v1.56.1-noble -c \
  'exec /ms-playwright/chromium-1194/chrome-linux/chrome --headless --no-sandbox \
    --disable-gpu --disable-dev-shm-usage --no-first-run --allow-file-access-from-files \
    --hide-scrollbars --force-color-profile=srgb --force-device-scale-factor=1 \
    --window-size=1280,720 --virtual-time-budget=2000 \
    --screenshot="/work/$1.png" "file:///work/$1.html" >"/work/$1-browser.log" 2>&1' sh "$variant"
