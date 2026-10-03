#!/usr/bin/env bash
# #59：正常一回合及明示索敵邊界；固定父狀態，不覆寫收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
mode="${1:?normal|clear|peace|level1|late1|late255，附-a或-b}"
[[ "$mode" =~ ^(normal|clear|peace|level1|late1|late255)-[ab]$ ]] || exit 1
test -d "$ROOT/workplace/oracle"
test -d "$ROOT/org_game/extracted"
test -s "$ROOT/workplace/oracle/hold60-r58-base/machine.state"
[[ "$(git -C "$GOLEM" rev-parse HEAD)" == f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c ]] || exit 1
[[ -z "$(git -C "$GOLEM" status --porcelain)" ]] || exit 1
name="intel59-r60-$mode"
test ! -e "$ROOT/workplace/oracle/$name"
export DOSGOLEM_EXTRA_MOUNT="$ROOT/workplace:/out"
args=(-load-state /out/oracle/hold60-r58-base/machine.state -tick 20000
  -click-hold 3000000 -clicks 421000000:544:85,424000000:560:85,451000000:448:85,476000000:460:85)
if [[ "$mode" != normal-* ]]; then
  level=01
  [[ "$mode" == late255-* ]] && level=ff
  at=420000010
  # 固定版本已量得043B進入步數437804861；晚注入只用於隔離導出規則。
  [[ "$mode" == late* ]] && at=437804865
  [[ "$mode" == peace-* ]] && at=432932986
  poke="lin:668C9@$at=$level"
  [[ "$mode" == peace-* ]] && poke+=";lin:60BEC@$at=02"
  # 晚注入FF案例改成同盟陣營，以核對+2B分支；不宣稱正常局面。
  [[ "$mode" == late255-* ]] && poke+=";lin:60BA5@$at=01"
  # 星域表+2A/+2B，故意留下舊值，以驗證玩家命令入口整表清零。
  for ((r=0; r<32; r++)); do
    printf -v addr '%X' "$((0x5F2CC+r*126+0x2A))"
    poke+=";lin:$addr@420000010=fe ff"
  done
  args+=(-poke "$poke")
fi
watch=5F2F6-5F2F7
[[ "$mode" == clear-* || "$mode" == peace-* ]] && watch=668C9-668C9
bash "$ROOT/tools/oracle.sh" "$name" "${args[@]}" -steps 510000000 \
  -watch "$watch" -regs-from 420000000 -regs-max 40 \
  -regs-at 0374:0AF2,04F8:003E,2703:0214,2703:02E4,2703:043B \
  -dump-mem-at "420000011:lin:0:1048576:/out/oracle/$name/initial.bin;432932988:lin:0:1048576:/out/oracle/$name/clear-input.bin;432950000:lin:0:1048576:/out/oracle/$name/clear-output.bin;437804867:lin:0:1048576:/out/oracle/$name/input.bin;449999999:lin:0:1048576:/out/oracle/$name/command.bin;509999999:lin:0:1048576:/out/oracle/$name/mem.bin"
for f in initial input command mem clear-input clear-output; do
  test "$(stat -c %s "$ROOT/workplace/oracle/$name/$f.bin")" = 1048576
done
test "$(stat -c %s "$ROOT/workplace/oracle/$name/screen.bin")" = 307200
