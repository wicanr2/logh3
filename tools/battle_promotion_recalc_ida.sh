#!/usr/bin/env bash
# IDA位址基底與原版CS:IP分列；沿固定第84輪匯出器。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for seg in 297B 090A; do
  test -f "$ROOT/workplace/ida/promotion79-r85-reviewed-$seg.com"
  test ! -e "$ROOT/workplace/ida/promotion79-r85-reviewed-$seg.json"
done
cd "$ROOT"
timeout 90s bash tools/ida.sh raw bash -c 'set -euo pipefail
task_dir="$(mktemp -d /tmp/promotion79-r85-ida.XXXXXX)"
trap '\''rm -rf "$task_dir"'\'' EXIT
for seg in 297B 090A; do
 cp "/work/promotion79-r85-reviewed-$seg.com" "$task_dir/promotion79-r84-$seg.com"
 if [[ "$seg" == 297B ]]; then ranges="0209:025D 05B2:083F 0E8B:10DE 116C:12FB"; else ranges="0E6D:0E88 0E88:ret"; fi
 idat -A "-L/work/promotion79-r85-reviewed-$seg.log" "-S/idascripts/battle_promotion_review_ida.py /work/promotion79-r85-reviewed-$seg.json $seg $ranges" "$task_dir/promotion79-r84-$seg.com"
done'
