#!/usr/bin/env bash
# #67：S3同盟正常四十列清單，固定輸入擷取0／1／26。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?parent|zero-a|zero-b|one-a|one-b|last-a|last-b}"
[[ "$mode" =~ ^(parent|(zero|one|last)-[ab])$ ]] || exit 1
name="data67-r57-$mode"
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export LOGH3_ORACLE_OUTPUT_ROOT="$ROOT/workplace/oracle"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
test ! -e "$LOGH3_ORACLE_OUTPUT_ROOT/$name"
args=(-tick 20000)
case "$mode" in
  parent)
    end=510000000
    args+=(-press 1,3,8,enter -press-at 105000000 -press-every 2000000
      -clicks 40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:188,301000000:544:85,304000000:560:85
      -click-hold 3000000 -poke 'lin:5E00C@300000001=7E 16 DC D3')
    ;;
  zero-*)
    end=525000000
    parent=data67-r57-parent
    args+=(-load-state "/out/oracle/$parent/machine.state"
      -clicks 511000000:400:40,516000000:400:60 -click-hold 3000000
      -mouse-x 440 -mouse-y 350 -mouse-at 522000000)
    ;;
  one-*|last-*)
    parent="data67-r57-zero-${mode##*-}"
    end=530000000
    clicks=526000000:440:350
    if [[ "$mode" == last-* ]]; then
      end=557000000
      # 第27次仍停在26，驗證上界沒有溢出。
      for ((i=1; i<27; i++)); do
        clicks+=",$((526000000+i*1000000)):440:350"
      done
    fi
    args+=(-load-state "/out/oracle/$parent/machine.state" -clicks "$clicks" -click-hold 100000)
    ;;
esac
if [[ "$mode" != parent ]]; then
  test -s "$ROOT/workplace/oracle/$parent/machine.state"
  test -s "$ROOT/workplace/oracle/$parent/flags.txt"
fi
bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps "$((end+1))" \
  -peek lin:5E6B6:4,lin:5E724:4,lin:5E00C:4 \
  -watch 5E6B6-5E6B7 -watch-file "/out/oracle/$name/writes.txt" \
  -save-state "$end:/out/oracle/$name/machine.state" \
  -dump-mem-at "$((end-1)):lin:0:1048576:/out/oracle/$name/mem.bin"
test "$(stat -c %s "$ROOT/workplace/oracle/$name/mem.bin")" = 1048576
test "$(stat -c %s "$ROOT/workplace/oracle/$name/screen.bin")" = 307200
