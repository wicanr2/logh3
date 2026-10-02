"""083正式renderer的35邊界收據；只比較固定局面，不升格32/33輪。"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

from pngout import game_palette, write_png

root, work = Path("/ref"), Path("workplace")
assert work.stat().st_uid == os.getuid()
prefix = "battle79-round34-after"
evidence = Path("docs/internals/data/battle-phase-frames.json")
data = json.loads(evidence.read_text())
spec = json.loads(Path("docs/internals/data/screen-battle.json").read_text())
f, m, s = (spec[k] for k in ["field", "minimap", "status"])
regions = {"full": [0, 0, 640, 480],
           "field": [f["x"], f["y"], f["cols"]*f["tile_w"], f["rows"]*f["tile_h"]],
           "minimap": [m[k] for k in ["x", "y", "w", "h"]],
           "status": [s[k] for k in ["x", "y", "w", "h"]]}


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


sources = ["internal/game/game.go", "internal/game/battle.go", "internal/game/battlepreparation.go",
           "internal/game/battlemovement.go", "internal/game/battlemovement_test.go",
           "internal/game/battlephase_test.go", "internal/game/battlepreparation_test.go",
           "internal/rules/battlemovement.go", "internal/rules/battlemarker.go",
           "internal/screen/battle.go", "docs/internals/data/screen-battle.json",
           "tools/scripts/battle_phase_render_receipt.py"]
report = {"schema": "logh3.battle-phase-render.v1", "classification": "same-state",
          "scope": data["scope"], "remake_head": sys.argv[1], "remake_working_tree": "dirty",
          "seed_original_and_remake": "0x12345678, before phase; unchanged at all 35 boundaries",
          "camera": [66, 30], "canvas": [640, 480], "scale": 1, "crop": None, "masks": [],
          "dosgolem_head": data["dosgolem_head"], "original_evidence_sha256": sha(evidence),
          "source_sha256": {p: sha(root / p) for p in sources},
          "original_assets_sha256": {p: sha(Path("/orig") / p) for p in ["SPRITE.SPR", "BLOCK_2.BLK", "ANI_PNL0.A98", "GRAPH.IMG", "MAP26.CMP"]},
          "image_ids": {"logh3-ebiten:2": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8",
                        "golang:1.24-bookworm": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac"},
          "regions": regions, "frames": [], "normal_frames": []}
report["command"] = "bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT=battle79-round34-after LOGH3_PREPARATION_OUTPUT=battle79-round34-after go test -count=1 -v -run \"TestMovementAllBoundariesMatchOriginal|TestMovementCommandsExecuteAndSave|TestGameMovementPhaseMatchesOriginal|TestBattlePreparationWholePhaseMatchesOriginal|TestBattlePreparationNormalPlayerExecute|TestScoutRecoverAndRelaunchNormal\" ./internal/game'"
report["logs"] = {n: sha(work / n) for n in ["battle79-round34-final-target.log", "battle79-round34-standard.log", "battle79-round34-gui.log"]}
counts = re.findall(r"正常資源結果幀：補給(\d+)、修理(\d+)", (work / "battle79-round34-final-target.log").read_text())
assert len(counts) == 1
report["normal_resource_results"] = {"supply_frames": int(counts[0][0]), "repair_frames": int(counts[0][1]),
                                     "limitation": "正常S3本回合沒有資源結果；結果底圖只由診斷fixture抽驗，不宣稱正常資源結果畫面已驗"}


def save(name, px, pal):
    p = work / (prefix+"-"+name+".png")
    if p.exists():
        assert p.stat().st_uid == os.getuid()
    write_png(str(p), 640, 480, px, pal)
    return {"path": str(p), "sha256": sha(p)}


for frame in data["frames"]:
    name = frame["name"]
    op = work / ("oracle/battle79-phase-r34-a/"+name+".bin")
    assert sha(op) == frame["hashes"][".bin"]
    a = op.read_bytes()
    p = work / (prefix+"-boundary-"+name+".bin")
    b = p.read_bytes()
    pa = game_palette(str(op.with_suffix(".png")))
    pb = [v << 2 | v >> 4 for v in p.with_suffix(".pal").read_bytes()]
    assert len(a) == len(b) == 640*480 and len(pa) == len(pb) == 48
    differences = [i for i in range(len(a)) if pa[a[i]*3:a[i]*3+3] != pb[b[i]*3:b[i]*3+3]]
    stats = {}
    for k, (x, y, w, h) in regions.items():
        ds = [i for i in differences if x <= i % 640 < x+w and y <= i // 640 < y+h]
        stats[k] = {"pixels": w*h, "equal": w*h-len(ds), "different": len(ds),
                    "max_channel_difference": max((abs(pa[a[i]*3+c]-pb[b[i]*3+c]) for i in ds for c in range(3)), default=0),
                    "bounding_box": [min(i % 640 for i in ds), min(i // 640 for i in ds), max(i % 640 for i in ds), max(i // 640 for i in ds)] if ds else None}
    diff = bytearray(len(a))
    for i in differences:
        diff[i] = 1
    report["frames"].append({"name": name, "original_hashes": frame["hashes"],
                             "remake_frame_sha256": sha(p), "remake_palette_sha256": sha(p.with_suffix(".pal")),
                             "effective_palette_equal": pa == pb, "region_stats": stats,
                             "original": save(name+"-original", a, pa), "remake": save(name+"-remake", b, pb),
                             "difference": save(name+"-difference", diff, [0, 0, 0, 255, 0, 255]+[0]*42)})
for name in ["long-moving", "formation-moving", "scout-moving", "unit-moving", "long-after", "scout-deployed", "scout-away", "scout-recovered", "scout-relaunched", "supply-panel", "repair-panel", "normal-execute"]:
    p = work / (prefix+"-"+name+".bin")
    pal = [v << 2 | v >> 4 for v in p.with_suffix(".pal").read_bytes()]
    report["normal_frames"].append({"name": name, "frame_sha256": sha(p),
                                   "scope": "diagnostic resource fixture" if name in ["supply-panel", "repair-panel"] else "normal player path; no position or flag injection",
                                   "png": save("sample-"+name, p.read_bytes(), pal)})
out = work / (prefix+"-render-receipt.json")
out.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n")
assert all(f["effective_palette_equal"] and f["region_stats"]["full"]["different"] == 0 for f in report["frames"])
print("35全畫面及有效RGB零差異；收據SHA-256", sha(out))
