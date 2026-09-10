#!/usr/bin/env bash
# 只比對、不覆蓋：用翻譯後的旗標把收據重跑一次，記下畫面與調色盤差幾個位元組。
#
#   tools/check_receipts.sh                 # 全部
#   tools/check_receipts.sh main1e cb1      # 指定幾份
#   JOBS=3 tools/check_receipts.sh          # 並行數（預設 2）
#
# 與 `tools/rerun_receipts.sh` 的差別是**它不寫回 workplace/oracle/**。
# 換執行器的時候要先知道「哪幾份對不上」，而重跑會把舊畫面蓋掉——
# 蓋掉之後就再也比不出來了（收據沒有版控）。
#
# 輸出 `workplace/receipt-check.json`：一份「批次 → 差幾個位元組」的表，
# 給 worklist 的 verify 綁用（`docs/worklist.json`）。
#
# ⚠ **旗標從 `workplace/flags-new/` 讀**，那是
# `tools/scripts/translate_receipt_flags.py` 產生的新語法版本；
# 收據 `meta.txt` 裡那一行是舊語法，直接餵給新版 probe 會被拒或誤讀。
set -uo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
FLAGS="$ROOT/workplace/flags-new"
OUT="$GOLEM/workplace/check"
JOBS="${JOBS:-2}"
mkdir -p "$OUT"

names=("$@")
if [[ ${#names[@]} -eq 0 ]]; then
  mapfile -t names < <(cd "$FLAGS" && ls *.txt 2>/dev/null | sed 's/\.txt$//')
fi
[[ ${#names[@]} -gt 0 ]] || { echo "沒有翻好的旗標——先跑 tools/py.sh tools/scripts/translate_receipt_flags.py --all" >&2; exit 2; }

one() {
  local name="$1" f="$FLAGS/$1.txt" exe
  local -a fl
  mapfile -t fl < "$f"
  exe=$(sed -n 's/^執行檔: //p' "$ROOT/workplace/oracle/$name/meta.txt")
  rm -f "$OUT/$name.bin" "$OUT/$name.pal"
  # ⚠ `-dump-ega`／`-dump-palette` 放在最後：有些收據自己也帶了，
  # Go 的 flag 取最後一個，這樣才確定倒的是我們要比的那一份。
  (cd "$GOLEM" && DOSGOLEM_CPUS="${DOSGOLEM_CPUS:-2}" \
     DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_TIMEOUT="${DOSGOLEM_TIMEOUT:-30m}" \
     tools/go.sh run ./cmd/probe -exe "${exe:-/orig/GIN3.COM}" -root /orig -empty-env \
     "${fl[@]}" -dump-ega "/src/workplace/check/$name.png" \
     -dump-palette "/src/workplace/check/$name.pal" >/dev/null 2>&1)
  local s p
  if [[ -s "$OUT/$name.bin" ]]; then
    s=$(cmp -l "$OUT/$name.bin" "$ROOT/workplace/oracle/$name/screen.bin" 2>/dev/null | wc -l)
    p=$(cmp -l "$OUT/$name.pal" "$ROOT/workplace/oracle/$name/pal.bin" 2>/dev/null | wc -l)
  else
    s=-1; p=-1   # 跑不出畫面：旗標壞了或程式死在半路，與「差很多」不是同一回事
  fi
  printf '%s\t%s\t%s\n' "$name" "$s" "$p"
}
export -f one
export ROOT GOLEM FLAGS OUT

tmp=$(mktemp)
printf '%s\n' "${names[@]}" | xargs -P "$JOBS" -I{} bash -c 'one "$@"' _ {} | tee "$tmp"

python3 - "$tmp" "$ROOT/workplace/receipt-check.json" "$(git -C "$GOLEM" rev-parse --short HEAD)" <<'PY'
import json, sys, datetime
rows = []
for line in open(sys.argv[1], encoding="utf-8"):
    f = line.rstrip("\n").split("\t")
    if len(f) == 3:
        rows.append({"name": f[0], "screen_diff": int(f[1]), "pal_diff": int(f[2])})
rows.sort(key=lambda r: r["name"])
bad = [r["name"] for r in rows if r["screen_diff"] != 0 or r["pal_diff"] != 0]
json.dump({"generated": datetime.datetime.now().astimezone().isoformat(timespec="seconds"),
           "golem_head": sys.argv[3], "checked": len(rows),
           "mismatched": bad, "results": rows},
          open(sys.argv[2], "w", encoding="utf-8"), ensure_ascii=False, indent=2)
open(sys.argv[2], "a", encoding="utf-8").write("\n")
print(f"\n比對 {len(rows)} 份：{len(rows)-len(bad)} 份逐位元組一致、{len(bad)} 份對不上")
PY
rm -f "$tmp"
