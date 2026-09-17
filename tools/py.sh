#!/usr/bin/env bash
# Python 分析走 docker，不裝到主機。
#
#   tools/py.sh scripts/foo.py [args...]
#
# 掛載邊界（AGENTS.md「Docker-only 與原版素材邊界」）：
#   /orig            原版素材，**唯讀**
#   /work/workplace  一次性證據，可寫
#   /work/docs       產出的文件，可寫
#   /work/tools      腳本本身，唯讀
#   /work/assets     進版控的原版素材（`assets/original/`，AGENTS.md），可寫
# 沒列在上面的目錄容器看不到——包含 .git。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_PY_IMAGE:-python:3.12-slim}"
mkdir -p "$ROOT/workplace" "$ROOT/assets"
# 需要額外的唯讀參考資料時（例如倚天字型）用 LOGH3_EXTRA_RO 指目錄，掛在 /ref。
EXTRA=()
[[ -n "${LOGH3_EXTRA_RO:-}" ]] && EXTRA+=(-v "$(cd "$LOGH3_EXTRA_RO" && pwd):/ref:ro")

exec timeout "${LOGH3_TIMEOUT:-15m}" docker run --rm --network none \
  --memory "${LOGH3_MEM:-2g}" --cpus "${LOGH3_CPUS:-2}" --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$ROOT/org_game/extracted:/orig:ro" \
  -v "$ROOT/workplace:/work/workplace" \
  -v "$ROOT/docs:/work/docs" \
  -v "$ROOT/tools:/work/tools:ro" \
  -v "$ROOT/assets:/work/assets" \
  "${EXTRA[@]}" \
  -e HOME=/tmp -e RECT="${RECT:-}" -w /work \
  "$IMAGE" python3 "$@"
