#!/usr/bin/env bash
# #85 全表收據；入口 docs/re/354、規格 066。
# tools/runtime_fields_receipts.sh [批次名] [gui|stationed]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOLEM="${DOSGOLEM_DIR:-$HOME/cht/dosgolem-logh3}"
name="${1:-runtime85-normal}"; mode="${2:-gui}"
[[ "$name" =~ ^[a-zA-Z0-9-]+$ ]] || exit 2
[[ "$mode" == gui || "$mode" == stationed ]] || exit 2
cd "$ROOT"
test -d "$GOLEM/workplace"
test -d "$ROOT/org_game/extracted"
test -f "$GOLEM/workplace/fcmd/fcmd-base.state"
if [[ "$mode" == stationed ]]; then
  exec tools/region_supply_receipts.sh "$name" stationed
fi
tools/oracle.sh "$name" -load-state /src/workplace/fcmd/fcmd-base.state \
  -steps 500000000 -tick 20000 -clicks 421000000:544:85,424000000:560:85 -click-hold 3000000 \
  -watch 5F4BE-5F4BF,53178-5317B \
  -regs-at 2AF5:00CB,2AF5:0036,2AF5:060D,2E54:1108 -regs-from 427000000 -regs-max 300 \
  -dump-mem-at "420000010:lin:0:1048576:/src/workplace/$name-initial.bin;427763580:lin:0:1048576:/src/workplace/$name-ai-ready.bin;430530800:lin:0:1048576:/src/workplace/$name-map-before.bin;430535000:lin:0:1048576:/src/workplace/$name-map-after.bin;499999999:lin:0:1048576:/src/workplace/$name-command.bin"
LOGH3_EXTRA_RO="$GOLEM/workplace" tools/py.sh -c '
import hashlib,json,pathlib,struct,sys
out=pathlib.Path("workplace/oracle",sys.argv[1]);flags=out.joinpath("flags.txt").read_text().splitlines()
parent=flags[flags.index("-load-state")+1];local=pathlib.Path("/ref")/parent.removeprefix("/src/workplace/")
memory=[]
for dump in flags[flags.index("-dump-mem-at")+1].split(";"):
    at,rest=dump.split(":",1);prefix,addr,length,path=rest.split(":",3)
    assert prefix=="lin" and addr=="0" and length=="1048576"
    raw=(pathlib.Path("/ref")/path.removeprefix("/src/workplace/")).read_bytes();assert len(raw)==1048576
    stage=pathlib.Path(path).stem.removeprefix(sys.argv[1]+"-");out.joinpath(stage+".bin").write_bytes(raw)
    memory.append({"stage":stage,"step":int(at),"bytes":len(raw),"sha256":hashlib.sha256(raw).hexdigest()})
seed=struct.unpack_from("<I",out.joinpath("initial.bin").read_bytes(),0x5E00C)[0];assert seed==0xD3DC167E
out.joinpath("inputs.json").write_text(json.dumps({"golem_head":sys.argv[2],"golem_working_tree":sys.argv[3],"original_image":"golang:1.24-bookworm","parent":{"path":parent,"sha256":hashlib.sha256(local.read_bytes()).hexdigest()},"memory":memory,"original_seed_hex":hex(seed),"remake_seed_hex":hex(seed),"seed_setup":"執行前固定原版檢查點及 remake strategyRandSeed；initial.bin 實讀 DS:94DC","flags":flags,"address_space":"dosgolem 執行期線性位址；CS:IP 另列","scope":"正常玩家編成與實行；無狀態注入；只裁決指定回合與表格。"},ensure_ascii=False,indent=2)+"\n")
' "$name" "$(git -C "$GOLEM" rev-parse HEAD)" "$(git -C "$GOLEM" status --porcelain)"
