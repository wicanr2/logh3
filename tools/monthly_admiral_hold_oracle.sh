#!/usr/bin/env bash
# #60：正常月初建造計劃的提督清單，固定父狀態與保持雙跑。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?parent|a|b|close-a|close-b}"
[[ "$mode" =~ ^(parent|a|b|close-a|close-b)$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
name="hold60-r58-month-clean-$mode"
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/oracle"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
args=(-tick 20000)
if [[ "$mode" == parent ]]; then
  end=225000000
  args+=(-press 1,3,8,enter -press-at 105000000 -press-every 2000000
    -clicks 40000000:300:170,95000000:300:170,130000000:448:330,155000000:448:310,168000000:464:330
    -click-hold 3000000)
elif [[ "$mode" == a || "$mode" == b ]]; then
  parent=hold60-r58-month-clean-parent
  test -s "$ROOT/workplace/oracle/$parent/machine.state"
  end=236300000
  args+=(-load-state "/out/oracle/$parent/machine.state"
    -clicks 231000000:617:350 -click-hold 300000)
else
  parent="hold60-r58-month-clean-${mode#close-}"
  test -s "$ROOT/workplace/oracle/$parent/machine.state"
  end=245000000
  args+=(-load-state "/out/oracle/$parent/machine.state"
    -rclicks 237000000:100:100 -click-hold 3000000)
fi
bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps "$((end+1))" \
  -peek lin:5E1D1:1 -watch 5E1D1-5E1D1 -watch-file "/out/oracle/$name/writes.txt" \
  -save-state "$end:/out/oracle/$name/machine.state" \
  -dump-mem-at "$((end-1)):lin:0:1048576:/out/oracle/$name/mem.bin"
test "$(stat -c %s "$ROOT/workplace/oracle/$name/mem.bin")" = 1048576
test "$(stat -c %s "$ROOT/workplace/oracle/$name/screen.bin")" = 307200
