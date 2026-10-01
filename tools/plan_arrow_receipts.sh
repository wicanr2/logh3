#!/usr/bin/env bash
# #80／#82 的原版收據；契約與證據：docs/spec/003、docs/re/350。
# tools/plan_arrow_receipts.sh [批次名]；不給名字就全部重錄。
# 不改 dosgolem；各批次用獨立目錄，避免與探索輸出互相覆蓋。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
only="${1:-}"
cd "$ROOT"
test -d "$GOLEM/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-plan-tax.state" || {
  echo '缺徵稅檢查點；先用 tools/fleetcmd_receipts.sh 重生正常玩家路徑。' >&2
  exit 1
}
rec() {
  local name=$1; shift
  [[ -z "$only" || "$only" == "$name" ]] || return 0
  tools/oracle.sh "$name" "$@" -peek lin:5DB04:2,lin:5DAC2:2
  LOGH3_EXTRA_RO="$GOLEM/workplace" tools/py.sh -c '
import hashlib,json,pathlib,sys
out=pathlib.Path("workplace/oracle",sys.argv[1]); flags=out.joinpath("flags.txt").read_text().splitlines()
parent=None
if "-load-state" in flags:
    p=flags[flags.index("-load-state")+1]
    local=pathlib.Path("/ref")/p.removeprefix("/src/workplace/")
    parent={"path":p,"sha256":hashlib.sha256(local.read_bytes()).hexdigest()}
out.joinpath("inputs.json").write_text(json.dumps({"golem_head":sys.argv[2],"golem_working_tree":sys.argv[3],"original_image":"golang:1.24-bookworm","python_image":"python:3.12-slim","parent":parent,"rng":"從完整檢查點還原；箭頭操作不含亂數判定","canvas":[640,480],"mouse_y_to_screen_y":40},ensure_ascii=False,indent=2)+"\n")
' "$name" "$(git -C "$GOLEM" rev-parse HEAD)" "$(git -C "$GOLEM" status --porcelain)"
}
tax() { local name=$1; shift; rec "$name" -load-state /src/workplace/fcmd/fcmd-plan-tax.state "$@"; }
for entry in 'tax-up:471:249' 'tax-down:487:249' 'pct-up:471:289' 'pct-down:487:289'; do
  IFS=: read -r key x y <<<"$entry"
  tax "planarrow-$key" -steps 490000000 -mouse-x "$x" -mouse-y "$y" -mouse-at 452000000 \
    -clicks "462000000:$x:$y" -click-hold 100000
done
tax planarrow-hover -steps 461900000 -mouse-x 471 -mouse-y 249 -mouse-at 452000000
tax planarrow-moved -steps 491000000 -mouse-x 471 -mouse-y 249 -mouse-at 452000000 \
  -clicks 462000000:471:249,490000000:600:340:0:1 -click-hold 100000
tax planarrow-updown -steps 520000000 -mouse-x 471 -mouse-y 249 -mouse-at 452000000 \
  -clicks 462000000:471:249,502000000:487:249 -click-hold 100000
tax planarrow-reopen -steps 570000000 -mouse-x 471 -mouse-y 249 -mouse-at 452000000 \
  -clicks 462000000:471:249,502000000:463:329,532000000:352:85,557000000:400:125 -click-hold 100000
tax planarrow-clamp -steps 490000000 -mouse-x 471 -mouse-y 249 -mouse-at 452000000 \
  -poke lin:5DB04@451000000=04 -clicks 462000000:471:249 -click-hold 100000
# 月初預算從冷啟動走主選單、防拷問答、月初會計報告；與 budget1 同一份輸入。
rec planarrow-budget-base -steps 165000000 -tick 20000 \
  -clicks 40000000:300:170::3000000,95000000:300:170::3000000,130000000:432:300::300000 \
  -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -save-state 164999999:/src/workplace/planarrow-budget.state
for entry in 'tax-up:471:249' 'tax-down:487:249' 'pct-up:471:289' 'pct-down:487:289'; do
  IFS=: read -r key x y <<<"$entry"
  if [[ -z "$only" || "$only" == "planarrow-budget-$key" ]]; then
    test -f "$GOLEM/workplace/planarrow-budget.state"
    rec "planarrow-budget-$key" -load-state /src/workplace/planarrow-budget.state -steps 210000000 \
      -mouse-x "$x" -mouse-y "$y" -mouse-at 170000000 -clicks "185000000:$x:$y" -click-hold 100000
  fi
done
