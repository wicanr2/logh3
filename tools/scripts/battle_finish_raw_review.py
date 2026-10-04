"""核對完整原池模型與逐指令寫入，凍結來源；不代替正式Game驗收。"""
from pathlib import Path
import hashlib
import json


def digest(b):
    return hashlib.sha256(b).hexdigest()


def identity(path):
    b = Path(path).read_bytes()
    return {"path": path, "size": len(b), "sha256": digest(b)}


a = Path("workplace/finish79-r91-a.json").read_bytes()
b = Path("workplace/finish79-r91-b.json").read_bytes()
assert a == b and digest(a) == "190a1aa4a1c633be6fc4bee035dd696bbb80ddea021355ffa534df4fe5bc8c52"
trace = json.loads(a)
assert trace["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
assert len(trace["rows"]) == 2
baseline = json.loads(Path("workplace/finish79-r89-a.json").read_bytes())
baseline_rows = {r["name"]: r for r in baseline["rows"]}
writes = []
for row in trace["rows"]:
    original = baseline_rows[row["name"]]
    assert row["before"] == original["before"] and row["after"] == original["after"]
    assert row["seed_after_hex"] == original["seed_after_hex"] and row["ax"] == original["ax"]
    for write in row["writes"] or []:
        assert write["instruction"] in ("297B:0915", "297B:0919")
        assert write["after_word"] in (255, 65535)
        writes.append(write)
assert len(writes) == 7

paths = ["docs/internals/data/battle-finish.json", "docs/internals/data/battle-finish-attack.json",
         "docs/internals/data/battle-formation-slots.json", "docs/internals/data/ship-kind-table.json",
         "docs/internals/data/battle-terrain.json", "internal/asset/cmp.go", "org_game/extracted/MAP26.CMP",
         "org_game/extracted/GIN3.COM", "workplace/oracle/cleanup79-r88-entry-a/entry.state",
         "workplace/oracle/cleanup79-r88-entry-a/mem-in.bin", "workplace/finish79-r89-native.go",
         "workplace/finish79-r89-a.json", "workplace/finish79-r89-b.json",
         "workplace/finish79-r90-guard-a.json", "workplace/finish79-r90-guard-b.json",
         "workplace/finish79-r91-native.go", "workplace/finish79-r91-a.json", "workplace/finish79-r91-b.json",
         "tools/scripts/battle_finish_raw_prepare.py", "tools/battle_finish_raw_oracle.sh",
         "tools/battle_finish_raw_native_run.sh", "tools/scripts/battle_finish_raw_model.go",
         "tools/scripts/battle_finish_raw_model_prepare.py", "tools/battle_finish_raw_probe.sh",
         "tools/scripts/battle_finish_raw_review.py", "tools/scripts/battle_finish_attack_probe_test.go",
         "workplace/finish79-r91-model-test-v1.go", "workplace/finish79-r91-model-test-v2.go",
         "workplace/finish79-r91-model-test-v3.go", "workplace/finish79-r91-model-test.go",
         "workplace/finish79-r91-raw-model-v1.go", "workplace/finish79-r91-raw-base-v1.json",
         "workplace/finish79-r91-raw-guard-v1.json", "go.mod", "go.sum"]
cases = []
total = 0
terrain_calls = 0
for group, original_path, count in (
        ("base", "workplace/finish79-r89-a.json", 12),
        ("guard", "workplace/finish79-r90-guard-a.json", 7)):
    native = {r["name"]: r for r in json.loads(Path(original_path).read_bytes())["rows"]}
    for suffix in ("", "-terrain"):
        path = f"workplace/finish79-r91-raw-{group}{suffix}.json"
        rows = json.loads(Path(path).read_bytes())
        paths.append(path)
        assert len(rows) == count
        for model in rows:
            original = native[model["name"]]
            assert not model["differences"] and not any(model["raw_differences"].values())
            assert set(model["raw_hex"]) == set(original["after"])
            assert all(bytes.fromhex(v) == bytes.fromhex(original["after"][k]) for k, v in model["raw_hex"].items())
            assert model["seed"] == int.from_bytes(bytes.fromhex(original["seed_after_hex"]), "little")
            compared = sum(len(bytes.fromhex(v)) for v in model["raw_hex"].values())
            total += compared
            if suffix:
                assert model["map_sha256"] == digest(Path("org_game/extracted/MAP26.CMP").read_bytes())
                assert model["blocked_calls"] == 0
                terrain_calls += model["terrain_calls"]
            cases.append({"name": model["name"], "model": path, "original": original_path,
                          "comparison": "固定入口完整六表及清單的診斷；非正常玩家收據",
                          "compared_bytes": compared, "seed": model["seed"],
                          "after_sha256": {k: digest(bytes.fromhex(v)) for k, v in model["raw_hex"].items()},
                          "terrain_calls": model.get("terrain_calls", 0), "blocked_calls": model.get("blocked_calls", 0)})
assert len(cases) == 38 and total == 1701944
paths += [str(p) for p in sorted(Path("internal/rules").glob("*.go"))]
paths = list(dict.fromkeys(paths))
result = {"schema": "logh3.battle-finish-raw-review.v1", "issue": 79,
          "parent_head": "49c6236dd6f8fd1cb57cebddb72c2036ebeab61d",
          "scope": "19案完整原池與同19案實際地形模型；兩案逐指令目標欄位診斷，production未接",
          "corrections": ["第90輪FLEET+5A末端為行星索引的判讀錯誤；索引在+07，+5A/+5B定位清FF",
                          "SHIP+28/+2A隊形目標保留避障前Station；+34/+36現在位置才經避障調整"],
          "limits": ["19案地形判定均可通行，未驗地形拒絕的新組合；既有terrain契約仍適用",
                     "65535艘是敗方守軍格式邊界，攻方使用原版正常編制",
                     "正式Game、報告續接、六檔保存與正常玩家勝負仍待驗"],
          "sources": [identity(p) for p in paths], "writes": writes, "cases": cases,
          "totals": {"unique_cases": 19, "model_runs": 38, "compared_bytes": total,
                     "native_write_cases": 2, "write_changes": 7, "terrain_queries": terrain_calls}}
dest = Path("docs/internals/data/battle-finish-raw.json")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("來源", len(paths), result["totals"], "SHA-256", digest(dest.read_bytes()))
