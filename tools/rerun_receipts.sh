#!/usr/bin/env bash
# 重跑收據。**執行器一改就要跑這支**——同一組輸入在新版執行器下跑出來的
# 畫面可能不同，那份收據與所有依賴它的對拍結論同時失效（AGENTS.md）。
#
#   tools/rerun_receipts.sh              # 全部
#   tools/rerun_receipts.sh -m mouse-x   # 只重跑命令列含這個字串的
#   tools/rerun_receipts.sh -n 'main1e|cmd1e'  # 只重跑名字對得上的
#   JOBS=6 tools/rerun_receipts.sh       # 並行數（預設 4）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JOBS="${JOBS:-4}"
PREV="$ROOT/workplace/oracle-prev"   # 覆蓋前的備份（一份批次只留最原始的那次）
match=''
namere=''
while [[ "${1:-}" == -* ]]; do
  case "$1" in
    -m) match="$2"; shift 2 ;;
    -n) namere="$2"; shift 2 ;;
    *) echo "不認得的參數 $1" >&2; exit 2 ;;
  esac
done

n=0
for d in "$ROOT"/workplace/oracle/*/; do
  name=$(basename "$d")
  [[ -f "$d/meta.txt" ]] || continue
  # 旗標**優先讀翻譯過的新語法**（`workplace/flags-new/`，由
  # `tools/scripts/translate_receipt_flags.py` 產生）。收據裡那兩份記的是
  # 錄製當時的舊語法，直接餵給新版 probe 會被拒或誤讀——而誤讀不會報錯，
  # 只會跑出另一個畫面然後把它寫回收據。
  if [[ -f "$ROOT/workplace/flags-new/$name.txt" ]]; then
    mapfile -t flagv < "$ROOT/workplace/flags-new/$name.txt"
    flags="${flagv[*]}"
  elif [[ -f "$d/flags.txt" ]]; then
    mapfile -t flagv < "$d/flags.txt"
    flags="${flagv[*]}"
  else
    flags=$(sed -n 's/^probe 旗標: //p' "$d/meta.txt")
    read -r -a flagv <<<"$flags"
  fi
  [[ -n "$flags" ]] || continue
  [[ -z "$match" || "$flags" == *"$match"* ]] || continue
  [[ -z "$namere" || "$name" =~ $namere ]] || continue
  exe=$(sed -n 's/^執行檔: //p' "$d/meta.txt")
  before=$(sha256sum "$d/screen.bin" 2>/dev/null | cut -d' ' -f1 || echo -)
  # ⚠ **覆蓋之前先留一份。** 收據不進版控，寫回去就是唯一那份沒了——
  # 而「原本停在哪個畫面」正是判斷「新的是不是同一個狀態」的唯一依據。
  # 已經有備份就不再蓋（重跑第二次時，備份要維持是**最原始**的那一份）。
  if [[ ! -d "$PREV/$name" ]]; then
    mkdir -p "$PREV"
    cp -a "$d" "$PREV/$name"
  fi
  (
    # **失敗要印出來。** 原本 oracle.sh 一失敗，子 shell 就被 set -e 帶走，
    # 一行都不印——那份收據等於沒有重跑過，而且看不出來。
    if ! LOGH3_EXE="${exe:-/orig/GIN3.COM}" "$ROOT/tools/oracle.sh" "$name" "${flagv[@]}" >/dev/null 2>&1; then
      echo "  $name **重跑失敗**（旗標可能含空格；看 $d/flags.txt）"
      exit 0
    fi
    after=$(sha256sum "$d/screen.bin" | cut -d' ' -f1)
    [[ "$before" == "$after" ]] && echo "  $name 不變" || echo "  $name **變了**（$before → $after）"
  ) &
  n=$((n+1)); (( n % JOBS )) || wait
done
wait
echo "重跑 $n 份收據"
