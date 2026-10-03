#!/usr/bin/env bash
# #59：固定原版父狀態，注入清單分頁的邊界值；拒絕覆寫收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?ally-base|<empire-counts|empire-hidden|ally-counts|ally-hidden|morale49|morale50|battle|occupied>-<a|b>}"
[[ "$mode" =~ ^(ally-base|(empire-counts|empire-hidden|ally-counts|ally-hidden|ally-visible-counts|ally-visible-hidden|ally-normal|star-battle|morale49|morale50|battle|occupied)-(a|b))$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
name="list59-r59-$mode"
test ! -e "$ROOT/workplace/oracle/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
args=(-tick 20000 -click-hold 3000000)
if [[ "$mode" == ally-base ]]; then
  end=170000000
  args+=(-press 1,3,8,enter -press-at 105000000 -press-every 2000000
    -rclicks 130000000:300:200,155000000:300:200
    -clicks 40000000:300:190,95000000:300:190)
else
  case "$mode" in
    ally-*) parent=list59-r59-ally-base; at=171000000; menu=180000000; item=210000000; end=235000000 ;;
    *) parent=hold60-r58-base; at=420500000; menu=421000000; item=446000000; end=470000000 ;;
  esac
  test -s "$ROOT/workplace/oracle/$parent/machine.state"
  args+=(-load-state "/out/oracle/$parent/machine.state")
  case "$mode" in
    ally-visible-counts-*) args+=(-poke "lin:60236@$at=09 07 06 08") ;;
    ally-visible-hidden-*) args+=(-poke "lin:60236@$at=09 ff ff 08") ;;
    *-counts-*) args+=(-poke "lin:5F2F4@$at=09 07 06 08") ;;
    empire-hidden-*) args+=(-poke "lin:5F2F4@$at=ff 07 06 ff") ;;
    ally-hidden-*) args+=(-poke "lin:5F2F4@$at=09 ff ff 08") ;;
    morale49-*) args+=(-poke "lin:60BE0@$at=31") ;;
    morale50-*) args+=(-poke "lin:60BE0@$at=32") ;;
    battle-*) args+=(-poke "lin:5F2DF@$at=14") ;;
    occupied-*) args+=(-poke "lin:5F2DF@$at=10") ;;
    star-battle-*) args+=(-poke "lin:5F2DF@$at=04") ;;
  esac
  case "$mode" in
    empire-*|ally-*|star-battle-*) args+=(-clicks "$menu:448:85,$item:460:85") ;;
    *)
      args+=(-clicks "$menu:448:85,$item:460:105")
      if [[ "$mode" == battle-* || "$mode" == occupied-* ]]; then
        args+=(-clicks "$menu:448:85,$item:460:105,471000000:584:30")
        end=495000000
      fi
      ;;
  esac
fi
bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps "$((end+1))" \
  -peek lin:5F2F4:4,lin:60236:4,lin:5F2DF:1,lin:60BE0:1,lin:5DA4D:1 \
  -save-state "$end:/out/oracle/$name/machine.state" \
  -dump-mem-at "$((end-1)):lin:0:1048576:/out/oracle/$name/mem.bin"
test "$(stat -c %s "$ROOT/workplace/oracle/$name/mem.bin")" = 1048576
test "$(stat -c %s "$ROOT/workplace/oracle/$name/screen.bin")" = 307200
