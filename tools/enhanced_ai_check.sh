#!/usr/bin/env bash
# #87：正式 CLI 的原版／增強模式，新局、實行、存讀及下一回合。
# 規格與邊界：docs/spec/068-enhanced-ai-player-path.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
tools/py.sh -c 'from pathlib import Path
import os
p=Path("/work/workplace/playtest/ai87")
p.mkdir(parents=True,exist_ok=True)
assert p.stat().st_uid==os.getuid() and p.stat().st_gid==os.getgid()
for run in ("1","2"):
 for mode in ("original","enhanced"):
  q=p/run/mode;q.mkdir(parents=True,exist_ok=True)
  for f in q.iterdir(): assert f.stat().st_uid==os.getuid() and f.stat().st_gid==os.getgid(),f
  for f in q.glob("*.DT*"): f.unlink()
'
script='move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 544 125; click 560 125; tick 18; click 544 125; click 560 165; click 300 142; rclick 300 142; click 544 125; click 560 145; click 300 122; click 544 125; click 560 165; click 300 162; rclick 300 162; click 544 125; click 560 125; tick 18; click 544 125; click 560 165; click 300 182'
for run in 1 2; do
  for mode in original enhanced; do
    level=-1; [[ "$mode" == enhanced ]] && level=1
    out="workplace/playtest/ai87/$run/$mode"
    printf '%s\n' "$script" > "$out/script.txt"
    tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data \
      -ai "$level" -save "/src/$out" -script "$script" \
      -shot "/src/$out/gui.png" -state-out "/src/$out/state.txt" -shot-after 2
  done
done
LOGH3_EXTRA_RO="$ROOT" tools/py.sh -c 'from pathlib import Path
import json,hashlib,sys
p=Path("/work/workplace/playtest/ai87");receipt={}
for mode in ("original","enhanced"):
 a,b=p/"1"/mode,p/"2"/mode
 for name in ("STELLAR","PLANET","ADMIRAL","FLEET","SHIP","WORK"):
  for slot in "012":
   n=name+".DT"+slot;assert (a/n).read_bytes()==(b/n).read_bytes(),(mode,n,"重現不同")
 for name in ("FLEET","SHIP","STELLAR","PLANET"):
  assert (a/(name+".DT0")).read_bytes()==(a/(name+".DT1")).read_bytes(),(mode,name,"存讀改變")
 assert "state=StateSaveGame" in (a/"state.txt").read_text(),mode
 turns=[]
 for slot in "02":
  f=(a/("FLEET.DT"+slot)).read_bytes();counts=[0,0];commands=[]
  for i in range(32):
   r=f[i*122:(i+1)*122]
   if r[4]!=255:
    counts[r[1]]+=1
    if r[1]==1:commands.append([i,r[6],r[7],r[8]])
  assert counts[0]==1 and counts[1]>=3,(mode,slot,counts)
  turns.append({"slot":slot,"counts":counts,"commands":commands})
 receipt[mode]={"turns":turns,"sha256":{x.name:hashlib.sha256(x.read_bytes()).hexdigest() for x in a.iterdir() if x.is_file()}}
assert receipt["original"]["turns"][0]["commands"]!=receipt["enhanced"]["turns"][0]["commands"],"模式未改變命令"
assert receipt["original"]["turns"][1]["commands"]!=receipt["enhanced"]["turns"][1]["commands"],"載入後模式未改變命令"
receipt.update({"kind":"remake 正常 CLI 的增強策略效果／存讀／續回合；不宣稱原版同狀態 parity","git_head":sys.argv[1],"working_tree":sys.argv[2],"image":sys.argv[3],"initial_seed":"0xD3DC167E（執行前固定；續回合沿用正常載入器的既有種子設定）","input":"1/<模式>/script.txt；兩次未注入正常重現","replays":2,"source_sha256":{n:hashlib.sha256((Path("/ref")/n).read_bytes()).hexdigest() for n in ("cmd/gin3/main.go","internal/game/aimode.go","internal/game/game.go","internal/ai/focused.go","tools/enhanced_ai_check.sh")}})
receipt["command"]="tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -ai <-1|1> -save /src/workplace/playtest/ai87/<1|2>/<模式> -script <script.txt> -shot <gui.png> -state-out <state.txt> -shot-after 2"
(p/"receipt.json").write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
print("正常 CLI：原版／增強均有敵方艦隊，命令不同；存讀全表相同，續回合差異保持，兩次重現完全一致。")
' "$(git rev-parse HEAD)" "$(git status --porcelain)" "${LOGH3_GUI_IMAGE:-logh3-ebiten:2}"
