"""第97輪正常標題收據；在tools/gui.sh xrun容器內--freeze／--check。"""
from pathlib import Path
import ast, hashlib, json, sys

target=Path("docs/internals/data/title-player.json")
def sha(b): return hashlib.sha256(b).hexdigest()
native=Path("workplace/oracle")
for a,b in [("last-a","last-b"),("waits-b","waits-c"),("timeout-a","timeout-b")]:
    for f in ["screen.bin","pal.bin","flags.txt"]:
        assert (native/("title68-r97-"+a)/f).read_bytes()==(native/("title68-r97-"+b)/f).read_bytes(),(a,b,f)
    for name in [a,b]:
        meta=(native/("title68-r97-"+name)/"meta.txt").read_text()
        assert "dosgolem HEAD: f7523f9（未提交檔案 0 個）" in meta
        assert "沒實作的服務（0 種）" in (native/("title68-r97-"+name)/"report.txt").read_text()
pb,pc=Path("workplace/title68-r97-player-b"),Path("workplace/title68-r97-player-c")
names=sorted(p.name for p in pb.iterdir())
assert len(names)==29 and names==sorted(p.name for p in pc.iterdir())
assert all((pb/n).read_bytes()==(pc/n).read_bytes() for n in names)
player=json.loads((pb/"receipt.json").read_text())
assert player["injection"] is False and player["intro_title_delay_ns"]==5396420663 and player["title_wait_ns"]==15442619818
assert [v["state"] for v in player["events"]]==["StateIntro","StateIntro","StateTitle","StateTitle","StateIntro","StateTitle","StateMainMenu","StateMainGame"]
comparison=json.loads(Path("workplace/title68-r97-comparison/comparison.json").read_text())
assert len(comparison)==2 and all(v["different"]==0 and v["page"]["index_different"]==0 and v["padding"]["different"]==0 for v in comparison)
gui=[]
for name,expected in [("left","StateMainMenu"),("replay","StateIntro")]:
    root=Path("workplace/title68-r97-gui-"+name)
    r=json.loads((root/"receipt.json").read_text())
    assert r["events"][0]["tps"]==60 and r["events"][0]["injection"] is False
    assert r["events"][-1]["state"]["state"]==expected and r["events"][-1]["different"]==0
    for p in root.glob("*-x11-comparison.json"):
        v=json.loads(p.read_text());assert v["pixels"]==307200 and v["different"]==0 and v["max_channel_difference"]==0 and v["bbox"] is None
        gui.append(str(p))
assert len(gui)==5
ida=Path("workplace/ida/title68-r97-loader-ida-v5.json")
v=json.loads(ida.read_text())
assert v["ida_version"]=="9.4" and v["file_offset_zero_ea"]=="10100" and v["uid"]==1000
assert len(v["instructions"])==681 and sum(r["semantic_level"]=="proven" for r in v["instructions"])==9
assert all(r["semantic_source"] if r["semantic_level"]=="proven" else r["warning"] for r in v["instructions"])
raw=Path("workplace/ida/title68-r97-loader.com").read_bytes()
assert sha(raw)==v["input_sha256"]=="0138b89c7e5dbd9bcf708630921835806d01e712d5c8ad7c1d9bbb1b78fb3674"
for frame,name in [("last","TITSPT.RGB"),("title","SPTIT.RGB")]:
    rgb=Path("/orig",name).read_bytes()
    assert (pb/(frame+"-rgb16.bin")).read_bytes()==bytes(((x*4)<<2)|((x*4)>>4) for x in rgb)
standard=Path("workplace/round97-standard-gui.log").read_text()
assert standard.count("\nok  ")+int(standard.startswith("ok  "))==13 and standard.count("[no test files]")==5 and "FAIL" not in standard
for file in ["tools/scripts/title_loader_ida.py","tools/scripts/title_player_gui.py","tools/scripts/title_player_receipt.py"]: ast.parse(Path(file).read_text())

if sys.argv[1]=="--freeze":
    assert not target.exists()
    paths={"docs/goal/2026-10-05-title-player-round-97.md","docs/re/431-title-player.md","docs/spec/112-title-player.md",
        "docs/spec/020-intro-animation.md","docs/playtest/086-title-player.md","docs/internals/19-ibm16-350.md",
        "docs/internals/data/title-player-annotations.json","docs/internals/data/title-asset-receipts.json",
        "docs/internals/data/battle-lifecycle-player.json","tools/scripts/title_loader_ida.py","tools/scripts/title_player_gui.py",
        "tools/scripts/title_player_compare.go","tools/scripts/title_player_receipt.py","tools/scripts/fleet_player_gui_capture.go",
        "tools/oracle.sh","tools/ida.sh","tools/go.sh","tools/gui.sh","workplace/round97-standard-gui.log",
        "workplace/title68-r96-native/title68-r96-title.state","workplace/title68-r96-timeline/title68-r96-tail.state"}
    for d in ["internal/game","internal/bootstrap","internal/audio","internal/gfx","internal/screen","internal/asset","cmd/gin3"]:
        paths.update(str(p) for p in Path(d).glob("*.go"))
    for pattern in ["title68-r97-*","title68-r96-tail.state","title68-r96-title.state"]:
        for root in Path("workplace").glob(pattern):
            if root.is_file(): paths.add(str(root))
            else: paths.update(str(p) for p in root.rglob("*") if p.is_file())
    for root in native.glob("title68-r97-*"): paths.update(str(p) for p in root.rglob("*") if p.is_file())
    for pattern in ["title68-r97-*"]:
        paths.update(str(p) for p in Path("workplace/ida").glob(pattern) if p.is_file())
    sources={}
    for name in sorted(paths):
        p=Path(name);b=p.read_bytes();assert p.stat().st_uid==1000 and p.stat().st_gid==1000,name
        sources[name]={"bytes":len(b),"sha256":sha(b)}
    originals={}
    for p in Path("/orig").iterdir():
        if p.is_file():
            b=p.read_bytes();originals[p.name]={"bytes":len(b),"sha256":sha(b)}
    receipt={"schema":"logh3.title-player-receipts.v1","date":"2026-10-05","parent_commit":"481221ad222b02f8c88b2f1f1644fb3a8c6797af",
        "remake_worktree_dirty":True,"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","dosgolem_dirty":False,
        "gui_image":"logh3-ebiten:2","go_image":"golang:1.24-bookworm","ida_image":"ida-pro-9.4-idapython:locked-v1",
        "spec112":"CONFORMED; title player transitions and static pages","comparison":comparison,"gui_x11_captures":5,"gui_x11_pixels":1536000,
        "gui_tps":60,"player_repeated_files":29,"waiting":"hardware-spec approximation; no cycle/wall-clock claim",
        "standard_tests":{"packages":18,"pass":13,"no_tests":5},"ida":{"instructions":681,"proven":9},
        "original_inputs":originals,"sources":sources}
    with target.open("x") as f: f.write(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
elif sys.argv[1]!="--check":raise SystemExit("用法：--freeze／--check")
receipt=json.loads(target.read_text())
for name,expected in receipt["sources"].items():
    b=Path(name).read_bytes();assert len(b)==expected["bytes"] and sha(b)==expected["sha256"],name
for name,expected in receipt["original_inputs"].items():
    b=Path("/orig",name).read_bytes();assert len(b)==expected["bytes"] and sha(b)==expected["sha256"],name
print("正常標題、29檔雙跑、五張X11及",len(receipt["sources"]),"項來源符合；索引SHA-256",sha(target.read_bytes()))
