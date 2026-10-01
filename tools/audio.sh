#!/usr/bin/env bash
# 音訊工具的有界 Docker 入口；來源／用途見 docs/internals/17-mds.md。
# 例如：bash tools/audio.sh fluidsynth --version
# /orig 與 /src 唯讀；只有 /work（workplace/）可寫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_AUDIO_IMAGE:-logh3-audio:2}"
[[ -d "$ROOT/workplace" && -d "$ROOT/org_game/extracted" ]] || {
  echo '缺少既有研究輸出或原版目錄，停止掛載。' >&2; exit 2;
}
[[ "$(stat -c %u "$ROOT/workplace")" == "$(id -u)" ]] || {
  echo '研究輸出目錄不屬於目前使用者，停止寫入。' >&2; exit 2;
}
docker image inspect "$IMAGE" >/dev/null
exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
  --memory "${LOGH3_MEM:-2g}" --cpus "${LOGH3_CPUS:-2}" --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$ROOT:/src:ro" -v "$ROOT/org_game/extracted:/orig:ro" \
  -v "$ROOT/workplace:/work" -e HOME=/tmp -w /work "$IMAGE" "$@"
