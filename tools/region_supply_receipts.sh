#!/usr/bin/env bash
# #84 的完整回合收據；入口 docs/re/353、規格 039。
# tools/region_supply_receipts.sh [批次名] [stationed|gui]；預設 supply84-stationed。
# 正式收據與 supply84-explore 分開，所有注入都記入 flags.txt。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:-supply84-stationed}"
mode="${2:-stationed}"
[[ "$mode" == stationed || "$mode" == gui ]] || exit 2
[[ "$name" =~ ^[a-zA-Z0-9-]+$ ]] || exit 2
cd "$ROOT"
test -d "$GOLEM/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-base.state"
POKE='lin:60BA4@420000050=01;lin:60BE0@420000050=28;lin:668B8@420000050=28;lin:668BA@420000050=64 00;lin:668C0@420000050=00 00 00 00;lin:668C4@420000050=00 00 00 00'
inject=(-poke "$POKE")
[[ "$mode" == stationed ]] || inject=()
dumps=''
entries=('initial:420000010' 'move-finished:432635158' 'distributed:432905201' 'fleet-supplied:432905949' \
  'ships-filled:432909112' 'resupplied:433600000' 'empty-supplied:451152514' \
  'empty-filled:451154823' 'owner-pass:451788485' 'refill-before:456172977' \
  'refill-after:456183314' 'command:499999999')
[[ "$mode" == stationed ]] || entries=('initial:420000010' 'command:499999999')
for entry in "${entries[@]}"; do
  IFS=: read -r stage step <<<"$entry"
  [[ -z "$dumps" ]] || dumps+=';'
  dumps+="$step:lin:0:1048576:/src/workplace/$name-$stage.bin"
done
tools/oracle.sh "$name" -load-state /src/workplace/fcmd/fcmd-base.state \
  -steps 500000000 -tick 20000 -clicks 421000000:544:85,424000000:560:85 \
  -click-hold 3000000 "${inject[@]}" -dump-mem-at "$dumps" \
  -regs-at 2703:051D,2703:0540,1930:1572,1AAB:0FAF,2703:12EB,2703:08DE,2B5F:02B2,0654:001A \
  -regs-from 420000000 -regs-max 80
LOGH3_EXTRA_RO="$GOLEM/workplace" tools/py.sh -c '
import hashlib,json,pathlib,struct,sys
out=pathlib.Path("workplace/oracle",sys.argv[1]); flags=out.joinpath("flags.txt").read_text().splitlines()
parent=flags[flags.index("-load-state")+1]
local=pathlib.Path("/ref")/parent.removeprefix("/src/workplace/")
memory=[]
for dump in flags[flags.index("-dump-mem-at")+1].split(";"):
    at,rest=dump.split(":",1); prefix,addr,length,path=rest.split(":",3)
    assert prefix=="lin" and addr=="0" and length=="1048576"
    raw=(pathlib.Path("/ref")/path.removeprefix("/src/workplace/")).read_bytes()
    assert len(raw)==1048576
    stage=pathlib.Path(path).stem.removeprefix(sys.argv[1]+"-")
    out.joinpath(stage+".bin").write_bytes(raw)
    memory.append({"stage":stage,"step":int(at),"address_space":"dosgolem 執行期線性位址","start":0,"bytes":len(raw),"sha256":hashlib.sha256(raw).hexdigest()})
seed=struct.unpack_from("<I",out.joinpath("initial.bin").read_bytes(),0x54B30+0x94DC)[0]
assert seed==0xD3DC167E
out.joinpath("inputs.json").write_text(json.dumps({"golem_head":sys.argv[2],"golem_working_tree":sys.argv[3],"original_image":"golang:1.24-bookworm","python_image":"python:3.12-slim","parent":{"path":parent,"sha256":hashlib.sha256(local.read_bytes()).hexdigest()},"memory":memory,"original_seed_hex":hex(seed),"original_seed_setup":"執行前從完整原版檢查點固定；initial.bin 實讀 DS:94DC","remake_seed_hex":"0xd3dc167e","remake_seed_setup":"正式 strategyRandSeed，執行前固定；不影響未受控正式局的裁決","rng":"局部庫存公式不取亂數；完整正常回合只裁決相同初始局面及輸入的庫存與指定畫面，不宣稱整段骰序一致。","canvas":[640,480],"mouse_y_to_screen_y":40},ensure_ascii=False,indent=2)+"\n")
' "$name" "$(git -C "$GOLEM" rev-parse HEAD)" "$(git -C "$GOLEM" status --porcelain)"
