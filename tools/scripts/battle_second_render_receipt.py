"""S1第二戰場43邊界全畫面／逐區差分；入口注入、正常玩家抽樣另列。"""
import hashlib
import json
from pathlib import Path
import sys

from pngout import game_palette, write_png

prefix, head = sys.argv[1:3]
assert prefix in ("battle79-round36-before", "battle79-round36-after", "battle79-round37-before", "battle79-round37-after")
round_prefix = prefix.rsplit("-", 1)[0]
round37 = round_prefix.endswith("37")
work, root = Path("workplace"), Path("/ref")
evidence = Path("docs/internals/data/battle-second-frames.json")
data = json.loads(evidence.read_text())
spec = json.loads(Path("docs/internals/data/screen-battle.json").read_text())
f, m, s = (spec[k] for k in ("field", "minimap", "status"))
regions = {"full": [0, 0, 640, 480],
           "field": [f["x"], f["y"], f["cols"]*f["tile_w"], f["rows"]*f["tile_h"]],
           "minimap": [m[k] for k in ("x", "y", "w", "h")],
           "status": [s[k] for k in ("x", "y", "w", "h")]}


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def save(name, px, pal):
    p = work / (prefix + "-" + name + ".png")
    write_png(str(p), 640, 480, px, pal)
    return {"path": str(p), "sha256": sha(p)}


sources = ["internal/game/battlesecond_test.go", "internal/game/battlemovement.go",
	       "internal/game/battlepreparation_test.go", "internal/game/battlemovement_test.go",
	       "internal/game/battle_test.go", "internal/game/game.go", "internal/game/battle.go",
	       "internal/rules/battle.go", "internal/rules/resolve.go",
           "internal/rules/battlemovement.go", "internal/screen/battle.go",
           "tools/scripts/battle_second_render_receipt.py", "tools/scripts/battle_second_plan.py",
           "tools/battle_second_oracle.sh", "tools/scripts/battle_second_render_evidence.py",
           "docs/internals/data/battle-directions.json", "docs/internals/data/screen-battle.json"]
if round37:
    sources += ["tools/scripts/battle_order_evidence.py", "docs/internals/data/battle-draw-order.json"]
report = {"schema": "logh3.battle-phase-render.v1", "classification": "same-state",
          "scope": data["scope"], "remake_head": head, "remake_working_tree": "dirty",
          "seed_original_and_remake": data["movement_seed"], "seed_method": data["seed_method"],
          "camera": data["camera"], "canvas": [640, 480], "scale": 1, "crop": None, "masks": [],
          "dosgolem_head": data["dosgolem_head"], "original_evidence_sha256": sha(evidence),
          "source_sha256": {p: sha(root / p) for p in sources},
          "original_assets_sha256": {p: sha(Path("/orig") / p) for p in ("SPRITE.SPR", "BLOCK_1.BLK", "BLOCK_2.BLK", "ANI_PNL0.A98", "GRAPH.IMG", "MAP08.CMP")},
          "regions": regions, "frames": []}
if round37 and prefix.endswith("before"):
    snapshot = work / "battle79-round37-before-battle.go"
    report["source_sha256"]["internal/screen/battle.go"] = sha(snapshot)
    report["source_snapshot"] = {"internal/screen/battle.go": {"path": str(snapshot), "git_head": head, "sha256": sha(snapshot)}}
report["image_ids"] = {"logh3-ebiten:2": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8",
                       "golang:1.24-bookworm": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac",
                       "ida-pro-9.4-idapython:locked-v1": "sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780"}
report["command"] = "bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT="+prefix+" go test -count=1 -v -run \"TestMovementSecondBattleRecordsMatchOriginal|TestMovementAllBoundariesMatchOriginal|TestBattleShippedSavesNormalExecuteAndSave\" ./internal/game'"
if round37 and prefix.endswith("before"):
    report["command"] = "bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT="+prefix+" go test -count=1 -v -run TestMovementSecondBattleRecordsMatchOriginal ./internal/game'"
bundle = root / "org_game/银河英雄传说3SP.zip"
assert sha(bundle) == "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673"
report["bundle"] = {"path": "org_game/银河英雄传说3SP.zip", "sha256": sha(bundle)}
report["input_chain"] = []
for name in ("s1-path-explore", "s1-menu-explore", "s1-boundary-explore", "s1-boundary-tail-explore"):
    p = work / ("oracle/battle79-round36-" + name)
    report["input_chain"].append({"path": str(p), "scope": "exploration and parent snapshots, not promoted to a formal comparison receipt",
                                  "flags": (p / "flags.txt").read_text().splitlines(),
                                  "hashes": {q.name: sha(q) for q in p.iterdir() if q.is_file()}})
if round37:
    for suffix in ("s1-strip-explore", "s1-strip-sites", "s1-initial-strip", "s1-strip-early"):
        p = work / ("oracle/battle79-round37-"+suffix)
        report["input_chain"].append({"path": str(p), "scope": "VRAM source exploration only; no new parity claim",
                                      "flags": (p / "flags.txt").read_text().splitlines(),
                                      "hashes": {q.name: sha(q) for q in p.iterdir() if q.is_file()}})
    report["ida_exports"] = {"re79-render-r37-"+seg+ext: sha(work/"ida"/("re79-render-r37-"+seg+ext))
                             for seg in ("35D1", "2376") for ext in (".com", ".json", ".index.json")}
if prefix.endswith("after"):
    logs = [round_prefix+("-target.log" if round37 else "-final-target.log"), round_prefix+"-standard.log", round_prefix+"-gui.log", round_prefix+"-before.log"]
    if not round37:
        logs += ["battle79-round36-after-unresolved.log"]
    else:
        logs += ["battle79-round37-order-probe.log"]
    report["log_sha256"] = {n: sha(work / n) for n in logs}
    target = (work / logs[0]).read_text()
    assert target.count("--- PASS:") == 6 and "--- SKIP:" not in target and "--- FAIL:" not in target
    standard = (work / logs[1]).read_text().splitlines()
    assert len(standard) == 18 and all(s.startswith("ok") or s.startswith("?") for s in standard)
    report["normal_gui"] = {"scope": "normal product main menu / S1 / strategic Execute / tactical Execute / movement; no diagnostic state injection; not same-state oracle comparison",
                            "hashes": {round_prefix+n: sha(work / (round_prefix+n)) for n in ("-s1-gui-movement.png", "-s1-gui-movement.bin", "-s1-gui-movement.state")}}
for frame in data["frames"]:
    name = frame["name"]
    op = work / ("oracle/battle79-second-r36-a/" + name + ".bin")
    assert sha(op) == frame["hashes"][".bin"]
    a = op.read_bytes()
    p = work / (prefix + "-second-" + name + ".bin")
    b = p.read_bytes()
    pa = game_palette(str(op.with_suffix(".png")))
    pb = [v << 2 | v >> 4 for v in p.with_suffix(".pal").read_bytes()]
    assert len(a) == len(b) == 640*480 and len(pa) == len(pb) == 48
    diffs = [i for i in range(len(a)) if pa[a[i]*3:a[i]*3+3] != pb[b[i]*3:b[i]*3+3]]
    stats = {}
    for k, (x, y, w, h) in regions.items():
        ds = [i for i in diffs if x <= i % 640 < x+w and y <= i // 640 < y+h]
        stats[k] = {"pixels": w*h, "equal": w*h-len(ds), "different": len(ds),
                    "max_channel_difference": max((abs(pa[a[i]*3+c]-pb[b[i]*3+c]) for i in ds for c in range(3)), default=0),
                    "bounding_box": [min(i % 640 for i in ds), min(i // 640 for i in ds), max(i % 640 for i in ds), max(i // 640 for i in ds)] if ds else None}
    diff = bytearray(len(a))
    for i in diffs:
        diff[i] = 1
    report["frames"].append({"name": name, "original_hashes": frame["hashes"],
                             "remake_frame_sha256": sha(p), "remake_palette_sha256": sha(p.with_suffix(".pal")),
                             "effective_palette_equal": pa == pb, "region_stats": stats,
                             "original": save(name+"-original", a, pa), "remake": save(name+"-remake", b, pb),
                             "difference": save(name+"-difference", diff, [0, 0, 0, 255, 0, 255]+[0]*42)})
out = work / (prefix + "-render-receipt.json")
report["full_frame_passed"] = all(f["effective_palette_equal"] and f["region_stats"]["full"]["different"] == 0 for f in report["frames"])
report["field_passed"] = all(f["effective_palette_equal"] and f["region_stats"]["field"]["different"] == 0 for f in report["frames"])
report["remaining_differences"] = {"row_203": "unknown; S3 fixed residue mask does not match S1"}
if not (round37 and prefix.endswith("after")):
    report["remaining_differences"]["field_overlap"] = "unknown; original display object order differs from current unit list order"
else:
    assert all(f["region_stats"]["field"]["different"] == 0 for f in report["frames"])
    assert all(f["region_stats"]["full"]["different"] == 38 for f in report["frames"])
out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print("43邊界全畫面差異", [f["region_stats"]["full"]["different"] for f in report["frames"]], "收據", sha(out))
if "--require-exact" in sys.argv:
    assert report["full_frame_passed"]
