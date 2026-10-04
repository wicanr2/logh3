"""第82輪：凍結規則、報告矩形、玩家存檔選單與正常GUI的不同驗證範圍。"""
import hashlib
import json
from pathlib import Path

sources = {}


def read(name):
    p = Path(name)
    b = p.read_bytes()
    sources[str(p)] = {"sha256": hashlib.sha256(b).hexdigest(), "size": len(b)}
    return b


original = json.loads(read("docs/internals/data/battle-ownership-v2.json"))
assert sources["docs/internals/data/battle-ownership-v2.json"]["sha256"] == (
    "306f7612cb536cdb441bba2e245baecb9786fbc84a9210b066b06a09934ac262")
for name, expected in original["sources"].items():
    assert read(name) is not None and sources[name] == expected, name

target = read("workplace/round82-ownership-target-v8.log").decode()
assert "FAIL" not in target and "SKIP" not in target
assert "156案完整126-byte" in target and target.count("報告58240像素零差") == 6
assert target.count("正式8-1／8-2／等待／回合續接及六檔存讀通過") == 2
standard = read("workplace/round82-standard-v2.log").decode()
assert "FAIL" not in standard
assert len([s for s in standard.splitlines() if s.startswith("ok ")]) == 13
assert standard.count("[no test files]") == 5
assert "正常S3首戰：倍率1" in (gui := read("workplace/round82-normal-gui-v3.log").decode())
assert "正常S3首戰：倍率2" in gui

windows = []
out = Path("workplace/ownership79-r82-ui-v8")
for p in sorted(out.glob("*-comparison.json")):
    row = json.loads(read(p))
    assert row["window"]["pixels"] == 58240 and row["window"]["different"] == 0
    assert row["window"]["maximum_channel_difference"] == 0 and row["window"]["bbox"] is None
    assert row["full"]["different"] == 248681 and row["mask"] is None
    windows.append(row)
assert len(windows) == 6
for p in sorted(out.rglob("*")):
    if p.is_file():
        read(p)
for owner in (0, 1):
    for name in ("STELLAR", "PLANET", "ADMIRAL", "FLEET", "SHIP", "WORK"):
        read(out / f"save-owner{owner}" / f"{name}.DT0")

entry = "workplace/ownership79-r82-player-entry-v2/entry/screen.bin"
baseline = read(entry)
assert len(baseline) == 307200
read(entry.replace(".bin", ".png"))
for scale in (1, 2):
    base = f"workplace/round82-normal-gui-v3-{scale}x"
    assert read(base + ".bin") == baseline
    state = read(base + ".txt").decode()
    assert "state=StateBattle" in state and f"window={640*scale}x{480*scale}" in state
    for ext in (".png", ".log"):
        read(base + ext)

for name in (
    "internal/game/battle.go", "internal/game/battlefinal.go", "internal/game/game.go",
    "internal/game/battleownership.go", "internal/game/battleownership_test.go",
    "internal/rules/battleplanet.go", "docs/internals/data/turn-window-panel.json",
    "tools/scripts/battle_ownership_player_entry.go", "tools/turn_counter_gui.sh",
    "tools/go.sh", "tools/gui.sh", "tools/scripts/battle_ownership_game_receipt.py",
    "workplace/round82-save-menu-before.log", "workplace/round82-player-entry-v2.log",
):
    read(name)
for p in sorted(Path("org_game/extracted").glob("*.DT2")):
    read(p)
read("org_game/extracted/PLADAT.DAT")

result = {
    "schema": "logh3.battle-ownership-game.v1",
    "parent_commit": "58291ecad91f4efbfdfce26a3988707200044e51",
    "working_tree_dirty": True,
    "dosgolem_head": original["dosgolem_head"],
    "gui_image": "logh3-ebiten:2",
    "gui_image_id": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8",
    "scope": "156 controlled original cases; six report windows only; injected ownership with public save-menu; normal S3 entry GUI regression",
    "seed": "0x12345678 for controlled ownership and save tests; normal GUI does not fix or compare RNG",
    "windows": windows,
    "normal_gui": {"scales": [1, 2], "logical_pixels_each": 307200,
                   "same_state_original": False, "intro": False,
                   "inputs": ["click 300,270", "keys 138 Return", "click 300,322",
                              "click 318,208", "click 544,125", "click 560,125", "move 240,240"],
                   "source": "tools/turn_counter_gui.sh; actual X11 input; independent Game entry above"},
    "standard": {"passed_packages": 13, "no_test_packages": 5},
    "limits": ["full original/remake battle not compared", "normal GUI ownership change not verified",
               "save-menu tested through public Game clicks, not X11", "LoadSlot returns command phase; battle resume unknown",
               "062 remains DRAFT; 106 READY; 8-3 through 8-5 absent"],
    "sources": sources,
}
dest = Path("docs/internals/data/battle-ownership-game.json")
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("產品來源", len(sources), "SHA-256", hashlib.sha256(dest.read_bytes()).hexdigest())
