"""凍結本輪原版鏈、正式Game、保存與正常入口驗收。Docker內執行。"""
import hashlib
import json
from pathlib import Path

sources = {}


def read(path):
    p = Path(path)
    b = p.read_bytes()
    sources[str(p)] = {"size": len(b), "sha256": hashlib.sha256(b).hexdigest()}
    return b


original = json.loads(read("docs/internals/data/battle-cleanup.json"))
for item in original["sources"]:
    b = read(item["path"])
    assert len(b) == item["size"] and hashlib.sha256(b).hexdigest() == item["sha256"], item["path"]
chain_a = read("workplace/cleanup-chain79-r88-a.json")
assert chain_a == read("workplace/cleanup-chain79-r88-b.json")
assert hashlib.sha256(chain_a).hexdigest() == "0e45c28c9cfe015a6db29aa8d6a37a9e51a2dc559a981cf0f3b9a317e822b3db"
chains = json.loads(chain_a)["rows"]
assert len(chains) == 4
chain_summary = []
for row in chains:
    a = bytes.fromhex(row["before"]["admiral"])
    b = bytes.fromhex(row["after"]["admiral"])
    changed = [i for i in range(82) if a[i*28+6] != b[i*28+6]]
    assert len(changed) == 1 and not row["inputs"] and not row["frames"]
    chain_summary.append({"name": row["name"], "promoted_admiral": changed[0],
                          "seed_after_hex": row["seed_after_hex"],
                          "promotion_before_sha256": {k: hashlib.sha256(bytes.fromhex(v)).hexdigest() for k, v in row["promotion_before"].items()},
                          "after_sha256": {k: hashlib.sha256(bytes.fromhex(v)).hexdigest() for k, v in row["after"].items()}})

log = read("workplace/cleanup79-r88-final-test.log").decode()
assert "--- FAIL:" not in log and "--- SKIP:" not in log
for name in ("GameOriginal", "StaleCompositionSaveReload", "RejectsInvalidAtomically", "PromotionChainOriginal"):
    assert "--- PASS: TestBattleCleanup" + name in log
standard = read("workplace/cleanup79-r88-product-standard-final.log").decode()
assert "FAIL" not in standard
assert sum(l.startswith("ok  \t") or l.startswith("ok\t") for l in standard.splitlines()) == 13
assert standard.count("[no test files]") == 5
for name in ("before-test", "after-test", "game-test", "save-test", "standard"):
    read("workplace/cleanup79-r88-" + name + ".log")
read("workplace/cleanup79-r88-product-standard.log")

rows = json.loads(read("workplace/cleanup79-r88-a.json"))["rows"]
save_cases = 0
for row in rows:
    if row["mask"] != (1 << row["count"]) - 2:
        continue
    prefix = Path("workplace/cleanup79-r88-game") / (row["name"] + "-save")
    for name, key in {"ADMIRAL": "admiral", "FLEET": "fleet", "SHIP": "ship", "STELLAR": "stellar", "WORK": "work", "PLANET": "planet"}.items():
        want = bytes.fromhex(row["after"][key])
        if name == "PLANET":
            want = want[64:]
        assert read(prefix / (name + ".DT0")) == want, (row["name"], name)
    read(prefix / "SAVEDATA.IDX")
    save_cases += 1
assert save_cases == 7

normal = Path("workplace/cleanup79-r88-normal-a")
normal_b = Path("workplace/cleanup79-r88-normal-b")
assert sorted(p.name for p in normal.iterdir()) == sorted(p.name for p in normal_b.iterdir())
for p in sorted(normal.iterdir()):
    assert read(p) == read(normal_b / p.name), p.name
replay = json.loads(read(normal / "replay.json"))
assert replay["no_state_injection"] and replay["state"] == "StateSaveGame" and replay["battle_work_turn"] == 1
for scale in (1, 2):
    prefix = f"workplace/cleanup79-r88-normal-gui-{scale}x"
    assert read(prefix + ".bin") == read("workplace/round58-hold-player-b/last/screen.bin")
    state = read(prefix + ".txt").decode()
    assert "state=StateBattle" in state and f"window={640*scale}x{480*scale}" in state
    read(prefix + ".png")
    read(prefix + ".log")

for p in ("tools/scripts/battle_cleanup_chain_prepare.py", "tools/battle_cleanup_chain_oracle.sh",
          "workplace/cleanup-chain79-r88-native.go", "workplace/cleanup-chain79-r88-run.sh",
          "internal/rules/battlecleanup.go", "internal/rules/battlecleanup_test.go",
          "internal/game/battlecleanup.go", "internal/game/battlecleanup_test.go", "internal/game/battlefinal.go",
          "internal/game/battle_test.go",
          "internal/game/battlepreparation.go", "internal/game/promotion.go", "internal/game/battleownership.go",
          "internal/game/scoutfuel.go", "internal/rules/fleetsave.go", "internal/rules/resolve.go", "internal/rules/save.go",
          "tools/scripts/battle_cleanup_game_receipt.py", "tools/scripts/battle_promotion_player.go",
          "tools/go.sh", "tools/gui.sh", "tools/left_hold_gui.sh", "go.mod", "go.sum"):
    read(p)
result = {"schema": "logh3.battle-cleanup-game.v1", "issue": 79,
          "parent_commit": "1d15aaafc6f0d24b6742d907575540d4cb7e4305", "worktree_dirty": True,
          "build_images": ["golang:1.24-bookworm", "logh3-ebiten:2"],
          "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": False,
          "summary": {"original_cases": 35, "removed_groups": 77, "save_cases": 7, "chain_cases": 4,
                      "invalid_game_cases": 5, "invalid_rule_cases": 4, "standard_pass": 13, "standard_without_tests": 5},
          "chains": chain_summary, "normal_replay": {k: v for k, v in replay.items() if k != "inputs"},
          "limitations": "109與108 READY；死亡／足額功績明示注入。正常首回合續接與保存雙跑、X11首戰資料窗兩倍率另驗；不代表自然死亡／自然升級GUI或完整062同態。",
          "sources": [{"path": p, **s} for p, s in sorted(sources.items())]}
dest = Path("docs/internals/data/battle-cleanup-game.json")
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("清理產品來源", len(sources), hashlib.sha256(dest.read_bytes()).hexdigest())
