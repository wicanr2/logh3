#!/usr/bin/env bash
# 跑 decomp/specs 下所有函式的差異測試與突變對照，印出每個函式一行摘要。任何一個失敗就以非 0 結束。
#
#   decomp/tools/diff_all.sh <新批次名>        批次名會加上函式 id 當子目錄，所以同一個批次名可以重用到不同函式
set -uo pipefail
ROOT="${GIN3_REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
D="${DECOMP_DIR:-$ROOT/decomp}"   # 開發副本用：DECOMP_DIR 指到複製的 decomp/，GIN3_REPO_ROOT 指到真正的 repo
batch="${1:?用法: diff_all.sh <新批次名>}"
fail=0
printf '%-11s %-9s %-13s %-9s %s\n' 函式 案數 指令覆蓋 突變存活 結果
for spec in "$D"/specs/*.json; do
  case "$spec" in *.mutants.json) continue ;; esac
  id="$(basename "$spec" .json)"
  log="$(mktemp)"
  if bash "$D/tools/diff_run.sh" "$id" "$batch" >"$log" 2>&1; then ok=OK; else ok=失敗; fail=1; fi
  cases="$(grep -oE '案數 [0-9]+' "$log" | head -1 | grep -oE '[0-9]+' || echo '?')"
  cov="$(grep -oE '指令覆蓋 [0-9]+/[0-9]+ \([0-9.]+%\)' "$log" | head -1 | sed 's/指令覆蓋 //' || echo '?')"
  surv="$(grep -oE '存活 [0-9]+ 個' "$log" | tail -1 | grep -oE '[0-9]+' || echo '-')"
  printf '%-11s %-9s %-13s %-9s %s\n' "$id" "$cases" "${cov:-?}" "${surv:--}" "$ok"
  [[ "$ok" == OK ]] || sed -n '1,12p' "$log" | sed 's/^/    /'
  rm -f "$log"
done
exit $fail
