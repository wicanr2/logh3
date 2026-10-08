#!/usr/bin/env bash
# 重新取得 GIN3PS.EXE 解包後的映像、重定位表與標準 MZ 檔（給 IDA 用）。
#
#   decomp/tools/gin3ps_image.sh <新批次名>
#
# 原理：GIN3PS.EXE 是六層封裝（docs/internals/02）。dosgolem 冷開機跑 GIN3.COM，
# 第 4000 萬步點掉開場動畫，GIN3PS 進入六層解包。最內層先把整個映像解壓（尚未套重定位），
# 再逐筆加上載入段。於是：
#   第 58,313,900 步  映像解壓完成、重定位尚未開始（與最終映像只差 5,991 個 word）
#   第 58,339,129 步  最終入口 0349:0079，映像已重定位，啟動碼尚未執行
# 兩份傾印的差異就是重定位表。收據綁執行器版本，步數是這個 HEAD 與旗標下的值。
#
# 輸出：workplace/decomp/image/<批次名>/，全部是原版衍生物，不進 Git。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
HEAD_EXPECT=f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c
PRE_STEP=58313900
FINAL_STEP=58339129
WANT_FINAL_IMAGE_PREFIX=c3fa47da504d3b94   # docs/re/336 §9.3 記錄的舊解包映像雜湊開頭

name="${1:?用法: gin3ps_image.sh <新批次名>}"
[[ "$name" =~ ^[a-z0-9][a-z0-9._-]*$ ]] || { echo "批次名只能用小寫英數與 . _ -" >&2; exit 1; }
test -d "$ROOT/org_game/extracted"
head="$(git -C "$GOLEM" rev-parse HEAD)"
[[ "$head" == "$HEAD_EXPECT" ]] || { echo "dosgolem HEAD 不符：$head" >&2; exit 1; }
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || { echo "dosgolem 工作樹不乾淨" >&2; exit 1; }

out="$ROOT/workplace/decomp/image/$name"
test ! -e "$out" || { echo "批次已存在：$out" >&2; exit 1; }
mkdir -p "$out"

export DOSGOLEM_ORIG="$ROOT/org_game/extracted"
export DOSGOLEM_EXTRA_MOUNT="$out:/out"
flags="-exe /orig/GIN3.COM -root /orig -empty-env -tick 20000 -steps $((FINAL_STEP + 100)) -clicks 40000000:300:170 -click-hold 3000000"
( cd "$GOLEM" && tools/go.sh run ./cmd/probe $flags \
    -dump-mem-at "$PRE_STEP:lin:0:1048576:/out/pre.bin;$FINAL_STEP:lin:0:1048576:/out/final.bin" \
    > "$out/probe.log" 2>&1 )
test "$(stat -c %s "$out/pre.bin")" = 1048576
test "$(stat -c %s "$out/final.bin")" = 1048576

pyrun() {
  docker run --rm --network none --memory 512m --cpus 1 --pids-limit 32 \
    --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
    -v "$out:/i" -v "$ROOT/decomp/tools:/t:ro" golang:1.24-bookworm python3 -I "$@"
}
pyrun /t/derive_relocs.py /i/final.bin /i/pre.bin /i/relocs.json
pyrun /t/build_mz.py /i/pre.bin /i/final.bin /i/relocs.json /i > "$out/build.log"

grep -q "\"image_final_sha256\": \"$WANT_FINAL_IMAGE_PREFIX" "$out/image.json" \
  || { echo "最終映像雜湊與 docs/re/336 記錄不符" >&2; exit 1; }
{
  echo "批次: $name"
  echo "日期: $(date -Iseconds)"
  echo "dosgolem HEAD: $head"
  echo "probe 旗標: $flags"
  echo "GIN3.COM SHA-256: $(sha256sum "$ROOT/org_game/extracted/GIN3.COM" | cut -d' ' -f1)"
  echo "GIN3PS.EXE SHA-256: $(sha256sum "$ROOT/org_game/extracted/GIN3PS.EXE" | cut -d' ' -f1)"
  echo "pre 步數: $PRE_STEP  final 步數: $FINAL_STEP"
} > "$out/meta.txt"
cat "$out/meta.txt"
echo "--- image.json"
cat "$out/image.json"
