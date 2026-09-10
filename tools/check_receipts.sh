#!/usr/bin/env bash
# 只比對、不覆蓋：用翻譯後的旗標把收據重跑一次，記下畫面與調色盤差幾個位元組。
#
#   tools/check_receipts.sh                 # 全部（已經比過的會跳過）
#   tools/check_receipts.sh main1e cb1      # 指定幾份（指名的一定重比）
#   JOBS=3 tools/check_receipts.sh          # 並行數（預設 2）
#   FORCE=1 tools/check_receipts.sh         # 全部重比，不跳過
#
# ⚠ **這是幾個鐘頭的工作，會被外力打斷。** 455 份裡有跑十幾億道指令的，
# 而這台機器同時有別人在用——記憶體一緊就被 OOM 砍掉。所以每比完一份
# 就把結果附到 `workplace/receipt-check.tsv`，下次跑自動從那裡接下去；
# 全部重來一次的代價是好幾個鐘頭，而**被砍掉的那一次什麼都沒留下**
# 才是真正貴的地方。
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

# ⚠ **整個本體包在 main() 裡，最後一行才呼叫**——這不是排版偏好。
# bash 是**邊讀邊執行**的：這支要跑好幾個鐘頭，中途有人（包括我自己）
# 改了這個檔，bash 會從新內容的舊位元組位移接下去讀，噴出「未預期的
# 字組附近有語法錯誤」，而**已經比完的那幾百份就此結束**。包成函式之後
# bash 會先把整份解析完，跑起來就與檔案內容脫鉤了。
# （2026-09-11 踩過：跑到第 416 份時被自己的編輯打斷。）
main() {
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
FLAGS="$ROOT/workplace/flags-new"
OUT="$GOLEM/workplace/check"
TSV="$ROOT/workplace/receipt-check.tsv"
JOBS="${JOBS:-2}"
mkdir -p "$OUT"
touch "$TSV"

names=("$@")
picked=$#
if [[ ${#names[@]} -eq 0 ]]; then
  mapfile -t names < <(cd "$FLAGS" && ls *.txt 2>/dev/null | sed 's/\.txt$//')
fi
# ⚠ 這個檢查要在「跳過已比過的」之前——續跑到全部比完時剩下的清單是空的，
# 那是**做完了**，不是「沒有旗標」。
[[ ${#names[@]} -gt 0 ]] || { echo "沒有翻好的旗標——先跑 tools/py.sh tools/scripts/translate_receipt_flags.py --all" >&2; exit 2; }

# 指名的一定重比（要看某一份改了沒）；跑全部時跳過已經有結果的。
if [[ $picked -eq 0 && "${FORCE:-}" != "1" ]]; then
  mapfile -t names < <(printf '%s\n' "${names[@]}" | grep -vxF -f <(cut -f1 "$TSV") || true)
  echo "已經比過 $(wc -l < "$TSV") 份，這次要比 ${#names[@]} 份（FORCE=1 可全部重比）"
fi

one() {
  local name="$1" f="$FLAGS/$1.txt" exe
  local -a fl
  mapfile -t fl < "$f"
  # ⚠ **空的旗標檔要當成「沒比」，不要記成結果。**
  # 踩過：一邊跑這支、一邊重跑 translate_receipt_flags.py --all，
  # 批次讀到正在被重寫的檔案 → 旗標是空的 → probe 跑出別的畫面或
  # 根本沒畫面，而 TSV 記下一筆看起來很正常的「對不上」。
  # 連續 39 份同時「失敗」才看得出來是競態；單獨一份會被當成真差異。
  if [[ ${#fl[@]} -eq 0 ]]; then
    echo "$name	略過：旗標檔是空的（有人正在重寫 flags-new/ 嗎？）" >&2
    return
  fi
  exe=$(sed -n 's/^執行檔: //p' "$ROOT/workplace/oracle/$name/meta.txt")
  rm -f "$OUT/$name.bin" "$OUT/$name.pal"
  # ⚠ `-dump-ega`／`-dump-palette` 放在最後：有些收據自己也帶了，
  # Go 的 flag 取最後一個，這樣才確定倒的是我們要比的那一份。
  (cd "$GOLEM" && DOSGOLEM_CPUS="${DOSGOLEM_CPUS:-2}" \
     DOSGOLEM_ORIG="$ROOT/org_game/extracted" DOSGOLEM_TIMEOUT="${DOSGOLEM_TIMEOUT:-30m}" \
     tools/go.sh run ./cmd/probe -exe "${exe:-/orig/GIN3.COM}" -root /orig -empty-env \
     "${fl[@]}" -dump-ega "/src/workplace/check/$name.png" \
     -dump-palette "/src/workplace/check/$name.pal" >/dev/null 2>&1)
  local s p ref="$ROOT/workplace/oracle/$name/screen.bin"
  # ⚠ **收據不在就不能比。** `cmp` 對不存在的檔會失敗，而失敗的輸出是空的
  # ——`wc -l` 回 0，於是「沒有東西可比」被記成「差 0 ＝ 一致」。
  # 假通過比假失敗更貴：它不會有人去查。
  if [[ ! -s "$ref" ]]; then
    echo "$name	略過：找不到收據 $ref" >&2
    return
  fi
  if [[ -s "$OUT/$name.bin" ]]; then
    s=$(cmp -l "$OUT/$name.bin" "$ref" | wc -l)
    p=$(cmp -l "$OUT/$name.pal" "$ROOT/workplace/oracle/$name/pal.bin" 2>/dev/null | wc -l)
  else
    s=-1; p=-1   # 跑不出畫面：旗標壞了或程式死在半路，與「差很多」不是同一回事
  fi
  # **邊比邊寫**：被砍掉的時候已經比完的那些要留得住。
  flock "$TSV" -c "printf '%s\t%s\t%s\n' '$name' '$s' '$p' >> '$TSV'"
  printf '%s\t%s\t%s\n' "$name" "$s" "$p"
}
export -f one
export ROOT GOLEM FLAGS OUT TSV

if [[ ${#names[@]} -gt 0 ]]; then
  printf '%s\n' "${names[@]}" | xargs -P "$JOBS" -I{} bash -c 'one "$@"' _ {}
fi

# JSON 一律從 TSV 全量重出，所以續跑與一次跑完的結果相同。
python3 - "$TSV" "$ROOT/workplace/receipt-check.json" "$(git -C "$GOLEM" rev-parse --short HEAD)" <<'PY'
import json, sys, datetime
# 指名重比的那幾份會在 TSV 裡多一列——**後面那一列才算數**，
# 不然重比完的結果會被舊的蓋掉，而報表看起來完全正常。
seen = {}
for line in open(sys.argv[1], encoding="utf-8"):
    f = line.rstrip("\n").split("\t")
    if len(f) == 3:
        seen[f[0]] = {"name": f[0], "screen_diff": int(f[1]), "pal_diff": int(f[2])}
rows = sorted(seen.values(), key=lambda r: r["name"])
bad = [r["name"] for r in rows if r["screen_diff"] != 0 or r["pal_diff"] != 0]
json.dump({"generated": datetime.datetime.now().astimezone().isoformat(timespec="seconds"),
           "golem_head": sys.argv[3], "checked": len(rows),
           "mismatched": bad, "results": rows},
          open(sys.argv[2], "w", encoding="utf-8"), ensure_ascii=False, indent=2)
open(sys.argv[2], "a", encoding="utf-8").write("\n")
print(f"\n比對 {len(rows)} 份：{len(rows)-len(bad)} 份逐位元組一致、{len(bad)} 份對不上")
PY


} # main

main "$@"
