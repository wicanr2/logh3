#!/usr/bin/env bash
# #86：正式 GUI 正常入口 → 新編艦隊 → 回合 → 存進度 → 取進度 → 再存。
# 入口與驗收邊界見 docs/spec/067-live-fleet-save.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
OUT=workplace/playtest/save86
tools/py.sh -c 'from pathlib import Path
import os
p=Path("/work/workplace/playtest/save86")
p.mkdir(parents=True,exist_ok=True)
assert p.stat().st_uid==os.getuid() and p.stat().st_gid==os.getgid()
for f in p.iterdir(): assert f.stat().st_uid==os.getuid() and f.stat().st_gid==os.getgid(),f
for f in p.glob("*.DT*"): f.unlink()
'
script="move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 544 125; click 560 125; tick 18; click 544 125; click 560 165; click 300 142; rclick 300 142; click 544 125; click 560 145; click 300 122; click 544 125; click 560 165; click 300 162"
printf '%s\n' "$script" > "$OUT/script.txt"
tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data \
  -save "/src/$OUT" -script "$script" -shot "/src/$OUT/gui.png" \
  -state-out "/src/$OUT/state.txt" -shot-after 2
LOGH3_EXTRA_RO="$ROOT" tools/py.sh -c 'from pathlib import Path
import json,hashlib,struct,sys
p=Path("/work/workplace/playtest/save86")
files={n:(p/n).read_bytes() for ext in "01" for n in [s+".DT"+ext for s in ("STELLAR","PLANET","ADMIRAL","FLEET","SHIP","WORK")]}
assert "state=StateSaveGame" in (p/"state.txt").read_text()
for stem,size in (("FLEET",3904),("SHIP",27456)):
 a,b=files[stem+".DT0"],files[stem+".DT1"]
 assert len(a)==size and a==b,(stem,"存讀後全表不同")
f=files["FLEET.DT0"];counts=[0,0]
for i in range(32):
 r=f[i*122:(i+1)*122]
 if r[4]!=255:counts[r[1]]+=1
assert counts==[1,3],counts
for region in range(32):
 o=region*126+40
 assert files["STELLAR.DT0"][o:o+2]==files["STELLAR.DT1"][o:o+2]
assert files["PLANET.DT0"]==files["PLANET.DT1"]
receipt={"kind":"remake 正常 GUI 存讀回歸；不宣稱兩側整表原版 parity","git_head":sys.argv[1],"working_tree":sys.argv[2],"image":sys.argv[3],"initial_seed":"0xD3DC167E（正常新遊戲執行前固定）","input":"script.txt；正常入口，無注入","fleet_counts":counts,"fleet_bytes":3904,"ship_bytes":27456,"fleet_ship_different":0,"region_fleet_counts_different":0,"sha256":{n:hashlib.sha256(b).hexdigest() for n,b in files.items()},"gui_sha256":hashlib.sha256((p/"gui.png").read_bytes()).hexdigest()}
receipt["command"]="tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -save /src/workplace/playtest/save86 -script <script.txt> -shot /src/workplace/playtest/save86/gui.png -state-out /src/workplace/playtest/save86/state.txt -shot-after 2"
receipt["source_sha256"]={n:hashlib.sha256((Path("/ref")/n).read_bytes()).hexdigest() for n in ("internal/rules/fleetsave.go","internal/rules/save.go","internal/game/fleetsave.go","internal/game/game.go","internal/game/originalbrain.go","tools/live_fleet_save_check.sh")}
(p/"receipt.json").write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
print("正常 GUI：帝國 1／同盟 3 支艦隊；FLEET 3904／SHIP 27456 bytes 存讀零差異，艦隊計數與行星完整保留。")
' "$(git rev-parse HEAD)" "$(git status --porcelain)" "${LOGH3_GUI_IMAGE:-logh3-ebiten:2}"
