"""第93輪來源身份、分級原指令及產品收據；Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path.cwd()
assert str(root) == "/src" and os.getuid() == 1000
target = Path("docs/internals/data/fleet-order-action.json")

def digest(b):
    return hashlib.sha256(b).hexdigest()

def create(path, text):
    with Path(path).open("x") as f:
        f.write(text)

if sys.argv[1] == "--prepare-replay":
    subprocess.run(["python3", "tools/scripts/fleet_player_gui_prepare.py", "order93"], check=True)
    src = Path("internal/game/fleetcombat.go")
    dest = Path("workplace/order79-r93-replay-game.go")
    create(dest, src.read_text()+"\nfunc (g *Game) FleetChainProbeSeed(seed uint32) { g.battleRand = rules.NewRand(seed) }\n")
    subprocess.run(["gofmt", "-w", str(dest)], check=True)
    create("workplace/order79-r93-replay-overlay.json", json.dumps({"Replace": {str(src.resolve()): str(dest.resolve())}}, indent=2)+"\n")
    sys.exit(0)

if sys.argv[1] == "--prepare-gui-input":
    s = Path("tools/scripts/fleet_player_gui_driver.py").read_text()
    start = s.index("        send_input({'Type':4,'Button':3})")
    end = s.index("        camera_gestures.append", start)
    s = s[:start]+'''        # 第81輪工具確認X server送達，正式Update收到後立即放開。
        send_input({'Type':4,'Button':3,'Status':str(statuspath),'Edge':edge})
        after=wait(lambda s:s is not None and not s['buttons'][1],'鏡頭放開',5)
'''+s[end:]
    s = s.replace("            send_input(down)", '''            if kind=='right' and action['before']['state']=='StateBattle' and action['before']['idle'] and action['before']['stage']==-1:
                down['Status']=str(statuspath);down['Edge']=edge
            send_input(down)''')
    s = s.replace("    (out/'failed-trace.json').write_text", "    (out/'failed-camera.json').write_text(json.dumps(camera_gestures,ensure_ascii=False,indent=2)+'\\n')\n    (out/'failed-trace.json').write_text")
    create("workplace/order79-r93-gui-driver-v3.py", s)
    sys.exit(0)

if sys.argv[1] == "--prepare-command-gui":
    subprocess.run(["python3", "tools/scripts/fleet_player_gui_prepare.py", "order93cmd"], check=True)
    p = Path("workplace/round79-fleet-gui-game-order93cmd.go")
    s = p.read_text()
    needle = 'return map[string]any{"state":'
    assert s.count(needle) == 1
    extra = '''var flags []map[string]any
    field := [4]int{}
    if g.res.Battle != nil {
        f := g.res.Battle.Field
        field = [4]int{f.X,f.Y,f.Cols*f.TileW,f.Rows*f.TileH}
        for _,u := range g.battleUnits {
            if u.Squadron != 0 || g.slot == nil {continue}
            pool:=g.battleUnitPool(u)
            flags=append(flags,map[string]any{"fleet":u.Fleet,"faction":u.Faction,"out":u.Out,
                "action":u.Action,"pool":pool,"raw_action":int(g.slot.Ship[pool*rules.ShipRecord+0x1e]),
                "x":f.X+u.X-g.battleView.Col*f.TileW,"y":f.Y+u.Y-g.battleView.Row*f.TileH})
        }
    }
    layout:=[3]int{}
    if g.res.CommandMenus != nil { l:=g.res.CommandMenus.Layout;layout=[3]int{l.HighlightDX,l.HighlightDY,l.RowPitch} }
    '''
    s = s.replace(needle, extra+needle).replace('"state": g.state.String(),', '"state": g.state.String(), "flagships":flags,"field":field,"command_xy":[2]int{g.battleCmdX,g.battleCmdY},"command_layout":layout,', 1)
    assert '"flagships":flags' in s
    p.write_text(s)
    subprocess.run(["gofmt", "-w", str(p)], check=True)
    sys.exit(0)

if sys.argv[1] == "--check-gui-save":
    folder = Path("workplace/order79-r93-command-gui")
    r = json.loads((folder/"receipt.json").read_text())
    ship = (folder/"SHIP.DT0").read_bytes()
    pool = r["flagship"]["pool"]
    assert len(ship) == 352*78
    assert ship[pool*78+0x1e] == r["after"]["raw_action"] == 127
    assert len(list(folder.glob("*.DT0"))) == 6
    create("workplace/order79-r93-gui-save-check.json", json.dumps({
        "scope": "實際SHIP.DT0原池行動碼核對；其餘五檔存在與非空由來源身份保存",
        "ship_sha256": digest(ship), "pool": pool, "action": 127, "ship_bytes": len(ship),
        "six_files_nonempty": all(p.stat().st_size > 0 for p in folder.glob("*.DT0")),
    }, ensure_ascii=False, indent=2)+"\n")
    print("真實GUI的實際SHIP保存行動碼7F相符")
    sys.exit(0)

if sys.argv[1] == "--check":
    d = json.loads(target.read_text())
    for name, expected in d["sources"].items():
        b = Path(name).read_bytes()
        assert len(b) == expected["bytes"] and digest(b) == expected["sha256"], name
    print("來源核對通過", len(d["sources"]), digest(target.read_bytes()))
    sys.exit(0)

assert sys.argv[1] == "--freeze" and not target.exists()
sources = {}
def read(name):
    p = Path(name)
    assert p.stat().st_uid == 1000, name
    b = p.read_bytes()
    sources[str(p)] = {"bytes": len(b), "sha256": digest(b)}
    return b

a = read("workplace/order79-r93-a.json")
assert a == read("workplace/order79-r93-b.json")
native = json.loads(a)
assert len(native["rows"]) == 8
for r in native["rows"]:
    assert r["ship_equal"] and r["before_ship_hex"] == r["after_ship_hex"]
    assert r["seed_before_hex"] == r["seed_after_hex"]
    assert len(bytes.fromhex(r["before_ship_hex"])) == 352*78

semantics = {
    "1CC5:0849": ("26884748", "移動成功重寫FLEET+48為0"),
    "1CC5:08F1": ("26c6474801", "戰鬥成功重寫FLEET+48為1"),
    "1CC5:08A2": ("9a780edb25", "逐分隊建立命令顯示物件"),
    "1CC5:094B": ("9a780edb25", "逐分隊建立命令顯示物件"),
    "1CC5:08A9": ("9a930fdb25", "移除暫時顯示物件"),
    "1CC5:0952": ("9a930fdb25", "移除暫時顯示物件"),
    "2583:00BC": ("2688474c", "成功選點寫FLEET+4C抵達朝向"),
}
ram = read("workplace/oracle/cleanup79-r88-entry-a/mem-in.bin")
graded = []
for seg in ("1CC5", "2583", "25DB"):
    prefix = "workplace/ida/order79-r93-"+seg
    export = json.loads(read(prefix+".json"))
    com = read(prefix+".com")
    assert export["kernel_version"] == "9.4" and export["uid"] == 1000
    assert export["input_sha256"] == digest(com)
    offset = int(seg,16)*16
    assert com[256:] == ram[offset:offset+len(com)-256]
    for row in export["instructions"]:
        raw = bytes.fromhex(row["original_bytes_hex"])
        at = int(row["runtime_linear_hex"], 16)
        assert ram[at:at+len(raw)] == raw
        loc = row["original_location"]
        if loc in semantics:
            expected, semantic = semantics[loc]
            assert row["original_bytes_hex"] == expected
            graded.append({**row, "grade": "proven", "semantic": semantic, "warning": "", "source": "研究427；八案原版雙跑及原bytes資料流"})
assert len(graded) == len(semantics)

for pattern in ("workplace/order79-r93-*", "workplace/finish79-r93-*", "workplace/round79-fleet-gui-*-order93.*", "workplace/round79-fleet-gui-*-order93cmd.*"):
    for p in sorted(Path(".").glob(pattern)):
        if p.is_dir():
            for f in sorted(p.rglob("*")):
                if f.is_file(): read(f)
        elif p.is_file():
            read(p)
for name in (
    "internal/game/fleetorder.go", "internal/game/fleetorder_test.go",
    "tools/fleet_order_action_oracle.sh", "tools/scripts/fleet_order_action_prepare.py",
    "tools/scripts/fleet_order_action_receipt.py", "tools/scripts/battle_finish_player_probe.py",
    "tools/scripts/fleet_player_gui_prepare.py", "tools/scripts/fleet_player_gui_plan.go",
    "tools/scripts/fleet_player_gui_driver.py", "tools/scripts/fleet_player_chain_replay.go",
    "tools/scripts/sound_tail_input.go", "workplace/round79-input-v2",
    "tools/scripts/fleet_order_action_gui.py", "workplace/round79-capture",
): read(name)

data = {
    "schema": "logh3.fleet-order-action.v1", "date": "2026-10-05",
    "issue": 79, "parent_commit": "0f33179705a67b5e987247dfbc39e1ac9a2eaa42",
    "classification": "原版局部診斷與正常remake玩家回歸；原版完整玩家對拍未驗",
    "native": {"rows": 8, "repeat_equal": True, "ship_bytes_per_run": 219648,
        "program_sha256": native["program_sha256"], "dosgolem_head": native["dosgolem_head"],
        "selection_substitute": "2583:013D，原位置與傳入朝向；完整注入見原版JSON"},
    "instructions": graded, "sources": sources,
    "limitations": ["自然勝利／占領GUI未驗", "原版選點UI未重錄", "原版整回合同態未驗", "110 READY；062 DRAFT；Issue79 OPEN"],
}
create(target, json.dumps(data, ensure_ascii=False, indent=2)+"\n")
print("來源已固定", len(sources), digest(target.read_bytes()))
