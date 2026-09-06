#!/usr/bin/env bash
# 重建建造計劃畫面的「同狀態對拍」證據：用 -poke 把百分比陣列寫進原版記憶體，
# 每個局面倒出一張 640×480 色號陣列。清單來源就是
# workplace/explore/manifest.json，跑完直接覆蓋回去給
# internal/screen 的 TestBuildPlanStatesMatchOriginal 用。
#
#   tools/buildplan_states.sh [局面名...]      # 不給就重建 manifest 裡全部
#
# 證據不進 Git（workplace/ 被忽略），所以它必須能一鍵重做——
# 要靠翻對話紀錄拼命令列的東西，實際上就不會重做。
#
# ⚠ 收據綁執行器版本：dosgolem 一改，跑出來的畫面就可能不同。
# 這支腳本把 dosgolem 的 Git HEAD 寫進 workplace/explore/buildplan/meta.txt。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
OUT="$ROOT/workplace/explore/buildplan"
JOBS="${JOBS:-4}"

# 從主選單走到建造計劃畫面的點擊序列（同一組序列 ⇒ 游標落點相同 ⇒ 可整張比）
BASE='300,170,40000000,3000000;300,170,95000000,3000000;448,330,130000000,3000000;448,310,155000000,3000000'
# 百分比陣列在記憶體裡有兩份（現值與顯示用副本），兩份都要寫，
# 在第 4 次點擊之前的 154,000,000 步注入。
POKE_A=5DB06; POKE_B=5DB0D; POKE_AT=154000000
STEPS=175000000
# 開頭的版權保護要答對才進得去；-tick 決定計時器頻率，換掉會整組跑掉。
COMMON=(-tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000)

mkdir -p "$OUT"
mapfile -t PLAN < <(python3 - "$ROOT/workplace/explore/manifest.json" "$@" <<'PY'
import json, sys
mf = json.load(open(sys.argv[1], encoding='utf-8'))
want = set(sys.argv[2:])
for key, pct in mf['states'].items():
    if key.startswith('../'):      # 正常玩家路徑的收據，不是 poke 造的
        continue
    name = key.split('/')[-1].removesuffix('.bin')
    if want and name not in want:
        continue
    print(name, ' '.join('%02X' % v for v in pct))
PY
)

cd "$GOLEM"
head=$(git rev-parse --short HEAD)
dirty=$(git status --porcelain | wc -l)
mkdir -p workplace/bp

run_one() {
  local name="$1" hex="$2"
  DOSGOLEM_ORIG="$ROOT/org_game/extracted" tools/go.sh run ./cmd/probe \
    -exe /orig/GIN3.COM -root /orig "${COMMON[@]}" -steps "$STEPS" -clicks "$BASE" \
    -poke "lin:$POKE_A@$POKE_AT=$hex;lin:$POKE_B@$POKE_AT=$hex" \
    -dump-ega "/src/workplace/bp/$name.png" >/dev/null 2>&1
  cp "workplace/bp/$name.bin" "$OUT/$name.bin"
  echo "  $name 完成"
}

n=0
for line in "${PLAN[@]}"; do
  set -- $line; name="$1"; shift; hex="$*"
  run_one "$name" "$hex" &
  n=$((n+1)); (( n % JOBS )) || wait
done
wait

{
  echo "dosgolem HEAD: $head（工作樹修改數 $dirty）"
  echo "點擊序列: $BASE"
  echo "注入: lin:$POKE_A@$POKE_AT 與 lin:$POKE_B@$POKE_AT，各 7 bytes 的百分比"
  echo "共用旗標: ${COMMON[*]}"
  echo "步數: $STEPS"
  echo "重建時間: $(date -Is)"
} > "$OUT/meta.txt"
echo "共 ${#PLAN[@]} 個局面，收據資訊寫入 $OUT/meta.txt"
