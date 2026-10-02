#!/usr/bin/env bash
# #89讀檔座位；入口第46輪goal。拒絕覆寫，父狀態由同版本冷啟動產生。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:?normal89-r46-<menu|empire|alliance|versus|hover0|hover1|hover2>-<a|b>}"
[[ "$name" =~ ^normal89-r46-(menu|empire|alliance|versus|vsconfirm|vshandoff|hover[012])-[ab]$ ]] || exit 1
kind="${name#normal89-r46-}"; kind="${kind%-?}"
suffix="${name: -1}"
test -d "$ROOT/workplace"
test -d "$ROOT/org_game/extracted"
test -d "$GOLEM/workplace"
test ! -e "$ROOT/workplace/oracle/$name"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
if [[ "$kind" == menu ]]; then
  bash "$ROOT/tools/oracle.sh" "$name" -steps 230000000 -tick 20000 \
    -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
    -clicks 40000000:300:230,95000000:300:230,190000000:300:282 -click-hold 3000000 \
    -save-state "229999999:/out/oracle/$name/menu.state" \
    -dump-mem-at "229999999:lin:0:1048576:/out/oracle/$name/menu-mem.bin"
  exit
fi
parent="$ROOT/workplace/oracle/normal89-r46-menu-$suffix/menu.state"
test -f "$parent"
flag=-clicks
case "$kind" in
  empire) y=168 ;; alliance) y=188 ;; versus) y=208 ;;
  vsconfirm|vshandoff) y=208 ;;
  hover0) y=168; flag=-rclicks ;; hover1) y=188; flag=-rclicks ;; hover2) y=208; flag=-rclicks ;;
esac
clicks="240000000:318:$y"
steps=300000000
case "$kind" in
  vsconfirm) clicks+=",260000000:432:230" ;;
  vshandoff) clicks+=",260000000:432:230,310000000:544:85,313000000:560:85,330000000:432:230"; steps=350000000 ;;
esac
bash "$ROOT/tools/oracle.sh" "$name" \
  -load-state "/out/oracle/normal89-r46-menu-$suffix/menu.state" -steps "$steps" -tick 20000 \
  "$flag" "$clicks" -click-hold 3000000 -watch 5DA50-5DA5D \
  -regs-from 230000000 -regs-max 8 -regs-at 042F:0006 \
  -dump-mem-at "$((steps-1)):lin:0:1048576:/out/oracle/$name/loaded-mem.bin" \
  -peek lin:5DA44:64
