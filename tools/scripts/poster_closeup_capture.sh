#!/usr/bin/env bash
# 156：有界既有瀏覽器；只開本機提案，無外網。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BATCH_NAME="${LOGH3_POSTER_BATCH:-poster156}"
BROWSER_SCRIPT="${LOGH3_POSTER_BROWSER:-poster_closeup_browser.mjs}"
[[ "$BATCH_NAME" =~ ^[a-z0-9-]+$ ]]
[[ "$BROWSER_SCRIPT" =~ ^[a-z0-9_]+\.mjs$ ]]
BATCH="$ROOT/workplace/$BATCH_NAME"
test -d "$BATCH"
test -f "$BATCH/proposal.html"
test -f "$ROOT/tools/scripts/$BROWSER_SCRIPT"
test "$(stat -c '%u:%g' "$BATCH")" = "$(id -u):$(id -g)"
exec timeout 3m docker run --rm --network none --memory 1536m --cpus 2 --pids-limit 256 \
  --log-opt max-size=5m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  --mount "type=bind,source=$BATCH,target=/work" \
  --mount "type=bind,source=$ROOT/tools,target=/tools,readonly" \
  --entrypoint node mcr.microsoft.com/playwright:v1.56.1-noble "/tools/scripts/$BROWSER_SCRIPT"
