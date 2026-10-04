"""108正式資料／五表對拍／報告／存讀與正常入口來源索引。只在Docker執行。"""
import hashlib
import json
import os
from pathlib import Path

sources = {}


def read(path):
    p = Path(path)
    b = p.read_bytes()
    sources[str(p)] = {"bytes": len(b), "sha256": hashlib.sha256(b).hexdigest()}
    return b


for name in ("battle-promotion-global.json", "battle-promotion-cache.json"):
    manifest = json.loads(read("docs/internals/data/" + name))
    for item in manifest["sources"]:
        b = read(item["path"])
        assert len(b) == item["bytes"] and hashlib.sha256(b).hexdigest() == item["sha256"], item["path"]

for kind, names in {
    "rules": ("OriginalTables", "InvalidAtomic"),
    "game": ("GameOriginal", "WaitAndRanks", "SaveReload", "CachedPaths", "CurrentMilitary"),
}.items():
    log = read(f"workplace/round87-promotion-{kind}.log").decode()
    assert "--- FAIL:" not in log and "--- SKIP:" not in log
    for name in names:
        assert "--- PASS: TestBattlePromotion" + name in log, name

standard = read("workplace/round87-standard.log").decode()
assert "FAIL" not in standard
assert len([s for s in standard.splitlines() if s.startswith("ok  \t") or s.startswith("ok\t")]) == 13
assert standard.count("[no test files]") == 5

comparisons = []
out = Path("workplace/promotion79-r87-game-final")
for p in sorted(out.glob("*-comparison.json")):
    d = json.loads(read(p))
    assert d["window"]["different"] == 0 and d["window"]["pixels"] == 58240
    assert d["window"]["maximum_channel_difference"] == 0 and d["mask"] is None
    assert d["full"]["different"] == 248681
    for suffix in ("original", "remake", "diff"):
        read(str(p).replace("comparison.json", suffix + ".png"))
    comparisons.append({"path": str(p), **d})
assert len(comparisons) == 11
saves = sorted(out.glob("*-save"))
assert len(saves) == 9
for save in saves:
    for p in sorted(save.iterdir()):
        if p.is_file():
            read(p)

entry = read("workplace/promotion79-r87-player-a/entry/screen.bin")
read("workplace/promotion79-r87-player-a/entry/screen.png")
for scale in (1, 2):
    base = f"workplace/promotion79-r87-gui-{scale}x"
    assert read(base + ".bin") == entry
    state = read(base + ".txt").decode()
    assert "state=StateBattle" in state and f"window={640*scale}x{480*scale}" in state
    read(base + ".png")
    read(base + ".log")

normal = Path("workplace/promotion79-r87-normal-b")
second = Path("workplace/promotion79-r87-normal-c")
for p in sorted(normal.iterdir()):
    assert read(p) == read(second / p.name), p.name
replay = json.loads(read(normal / "replay.json"))
assert replay["no_state_injection"] and replay["battle_work_turn"] == 1
assert replay["state"] == "StateSaveGame" and replay["strategic_turn"] == 380

production = [
    "internal/bootstrap/load.go", "internal/game/promotion.go", "internal/game/promotion_test.go",
    "internal/rules/promotion.go", "internal/rules/promotion_test.go", "internal/rules/admiral.go",
    "internal/rules/admiral_test.go", "internal/game/game.go", "internal/game/battle.go",
    "internal/game/battle_test.go", "internal/game/game_test.go", "internal/game/turnreport.go",
    "internal/game/annihilation.go", "internal/game/fleetadj.go", "internal/game/fleetcmd.go",
    "internal/game/fleetdetail.go", "internal/game/listtabs.go", "internal/game/regioninfo.go",
    "internal/game/scoutfuel.go", "internal/game/staffpick.go", "internal/game/turnwindow.go",
    "internal/game/battlefinal.go", "internal/screen/list.go", "tools/scripts/battle_promotion_spec.py",
    "tools/scripts/battle_promotion_player.go", "tools/scripts/battle_promotion_game_receipt.py",
    "tools/scripts/battle_ownership_player_entry.go", "tools/turn_counter_gui.sh",
    "tools/go.sh", "tools/gui.sh", "docs/internals/data/admiral-promotion-spec.json",
    "docs/internals/data/ship-kind-table.json", "docs/internals/data/fleet-editor.json",
    "docs/internals/data/star-graph.json", "docs/internals/data/fleet-window.json",
    "docs/internals/data/turn-window-panel.json", "docs/internals/data/battle-menu.json",
]
for p in production:
    read(p)
result = {
    "schema": "logh3.battle-promotion-game.v1",
    "parent_commit": "b501420f46eb55314b1d9430fc931ce700ae1a79",
    "oracle_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c",
    "build_images": ["golang:1.24-bookworm", "logh3-ebiten:2"],
    "scope": "66控制案／121嘗試／53次五表／473RND；11則報告矩形與有效16色；9案六檔診斷存讀；2快取控制；正常S3公開實行及保存雙跑、首戰X11兩倍率。",
    "limitations": "108 READY；自然累積功績升級、正常升級GUI、8-1清空分隊到8-4輸入鏈、8-5及完整062原版同態仍未驗。診斷空白底圖全圖差異不豁免。",
    "summary": {"cases": 66, "attempts": 121, "upgrades": 53, "rnd": 473,
                "reports": 11, "save_cases": 9, "cache_cases": 2,
                "standard_pass": 13, "standard_without_tests": 5},
    "comparisons": comparisons,
    "normal_replay": {k: v for k, v in replay.items() if k != "inputs"},
    "sources": [{"path": p, **s} for p, s in sorted(sources.items())],
}
destination = Path("docs/internals/data/battle-promotion-game.json")
assert destination.parent.stat().st_uid == os.getuid()
with destination.open("x", encoding="utf-8") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("升級正式來源", len(sources), hashlib.sha256(destination.read_bytes()).hexdigest())
