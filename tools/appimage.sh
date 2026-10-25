#!/usr/bin/env bash
# 153：沿既有固定Type-2 runtime建立AppImage；主機只控制Docker與路徑。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BATCH_NAME="${LOGH3_APPIMAGE_BATCH:-appimage153}"
VERSION="${LOGH3_APPIMAGE_VERSION:-v.1.6.1-20261025}"
[[ "$BATCH_NAME" =~ ^[a-z0-9-]+$ ]]
[[ "$VERSION" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
BATCH="$ROOT/workplace/$BATCH_NAME"
test -d "$BATCH/AppDir"
test -f "$BATCH/SOURCE_DATE_EPOCH"
test -f "$ROOT/tools/scripts/appimage_build.sh"
test "$(stat -c '%u:%g' "$BATCH")" = "$(id -u):$(id -g)"
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
  --memory 3g --cpus 2 --pids-limit 128 --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" --entrypoint sh -e LOGH3_APPIMAGE_VERSION="$VERSION" \
  -v "$ROOT/tools:/tools:ro" -v "$BATCH/AppDir:/appdir:ro" -v "$BATCH:/out" \
  -w /out hr-appimage:runtime-recovery-r1 /tools/scripts/appimage_build.sh
