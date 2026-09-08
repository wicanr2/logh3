#!/usr/bin/env bash
# IDA Pro 9.4 headless（idat）走 docker。
#
#   tools/ida.sh analyze ENDING.EXE          # 建 .i64（第一次要跑）
#   tools/ida.sh script  ENDING.EXE x.py ... # 對既有 .i64 跑 IDAPython
#
# ⚠ 知識庫（~/.claude/knowledge-base/retro/ida-pro-9.4.md）的硬規則：
#   - exit code 不能當證據，唯一可信的訊號是**輸出檔本身**。
#   - headless 的 print 不進 stdout，腳本一律寫檔。
#   - 對同一個 .i64 連續開關會弄壞它 → 一支腳本一次處理多個目標。
#   - 把 x.i64 當輸入交給 idat 會從 x 重新載入並覆蓋資料庫。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_IDA_IMAGE:-ida-pro-9.4-idapython:locked-v1}"
WORK="$ROOT/workplace/ida"
mkdir -p "$WORK"
mode="$1"; shift

run() {
  docker run --rm --network none \
    --memory "${LOGH3_IDA_MEM:-6g}" --cpus "${LOGH3_IDA_CPUS:-2}" --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" \
    -v "$WORK:/work" \
    -v "$ROOT/tools/scripts:/idascripts:ro" \
    -e IDA_TRACE_OFFSETS="${IDA_TRACE_OFFSETS:-}" \
    -w /work "$IMAGE" "$@"
}

case "$mode" in
  analyze)
    exe="$1"
    cp -n "$ROOT/org_game/extracted/$exe" "$WORK/$exe"   # 原檔唯讀，複本才給 IDA 寫
    run idat -A -B "/work/$exe"
    ;;
  script)
    exe="$1"; py="$2"; shift 2
    run idat -A "-S/idascripts/$py $*" "/work/$exe.i64"
    ;;
  raw) run "$@" ;;
  *) echo "用法: tools/ida.sh {analyze|script|raw} ..." >&2; exit 2 ;;
esac
