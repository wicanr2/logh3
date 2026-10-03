#!/usr/bin/env bash
# #60：固定版本、正常新遊戲父路徑與指令數保持探針。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?base|tab|star|tax|<tab|star|tax|pct>-<hold>-<a|b|move|clock>}"
[[ "$mode" =~ ^(base|tab|star|tax|(tab|star|tax|pct)-[0-9]+-(a|b|move|clock))$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
name="hold60-r58-$mode"
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/oracle"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
args=(-tick 20000)
parent=''
watch=5E21C-5E21D
case "$mode" in
  base)
    end=420000000
    args+=(-press 1,3,8,enter -press-at 105000000 -press-every 2000000
      -rclicks 130000000:300:200,155000000:300:200
      -clicks 40000000:300:170,95000000:300:170,180000000:352:85,210000000:400:85,240000000:400:45,270000000:488:290,320000000:480:310,380000000:510:310
      -click-hold 3000000)
    ;;
  tab)
    parent=base; end=445000000; watch=5E108-5E10B
    args+=(-clicks 421000000:560:250 -click-hold 3000000)
    ;;
  star)
    parent=base; end=470000000
    args+=(-clicks 421000000:448:85,446000000:460:85 -click-hold 3000000)
    ;;
  tax)
    parent=base; end=470000000; watch=5DB04-5DB05
    args+=(-clicks 421000000:352:85,446000000:400:125 -click-hold 3000000)
    ;;
  *)
    owner="${mode%%-*}"; rest="${mode#*-}"; hold="${rest%%-*}"; variant="${rest##*-}"
    parent="$owner"
    case "$owner" in
      tab) click=451000000; x=617; y=348; watch=5E109-5E109 ;;
      star) click=476000000; x=615; y=350 ;;
      tax) click=476000000; x=471; y=249; watch=5DB04-5DB05 ;;
      pct) parent=tax; click=476000000; x=471; y=289; watch=5DAC2-5DAC3 ;;
    esac
    end=$((click+hold+5000000))
    args+=(-clicks "$click:$x:$y" -click-hold "$hold")
    if [[ "$variant" == move ]]; then
      args+=(-mouse-x 100 -mouse-y 100 -mouse-at "$((click+50000))")
    elif [[ "$variant" == clock ]]; then
      watch=5DA2E-5DA2F
    fi
    ;;
esac
if [[ -n "$parent" ]]; then
  parent="hold60-r58-$parent"
  test -s "$ROOT/workplace/oracle/$parent/machine.state"
  test -s "$ROOT/workplace/oracle/$parent/flags.txt"
  args+=(-load-state "/out/oracle/$parent/machine.state")
fi
bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps "$((end+1))" \
  -peek lin:5E21C:2,lin:5E108:4,lin:5DB04:2,lin:5DAC2:2 \
  -watch "$watch" -watch-file "/out/oracle/$name/writes.txt" \
  -save-state "$end:/out/oracle/$name/machine.state" \
  -dump-mem-at "$((end-1)):lin:0:1048576:/out/oracle/$name/mem.bin"
test "$(stat -c %s "$ROOT/workplace/oracle/$name/mem.bin")" = 1048576
test "$(stat -c %s "$ROOT/workplace/oracle/$name/screen.bin")" = 307200
