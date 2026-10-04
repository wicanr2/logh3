"""凍結占領核心模型與非零守軍證據；不宣稱完整原池或production通過。"""
from pathlib import Path
import hashlib
import json
import math


def sha(b):
    return hashlib.sha256(b).hexdigest()


def source(path):
    p = Path(path)
    b = p.read_bytes()
    return {"path": path, "size": len(b), "sha256": sha(b)}


def word(b, o):
    return int.from_bytes(b[o:o + 2], "little")


a = Path("workplace/finish79-r90-guard-a.json").read_bytes()
b = Path("workplace/finish79-r90-guard-b.json").read_bytes()
assert a == b and sha(a) == "85fbd698622c307051183167d2315e9977abe5cac8fc8068611b31f4287ee335"
original = json.loads(a)
assert len(original["rows"]) == 7
cases = []
queries = 0
for row in original["rows"]:
    before = {k: bytes.fromhex(v) for k, v in row["before"].items()}
    fleet_ids = before["lists"][4:11]
    assert len(fleet_ids) == 7
    # 所有敗方艦隊在8-5入口仍有bit0；近距離守軍不依這個bit過濾。
    loser = 0 if row["mask"] == 1 else 1
    assert all(before["fleet"][i*122] & 1 for i in fleet_ids if before["fleet"][i*122+1] == loser)
    guards = [c for c in row["calls"] if c["location"] == "203A:04D1"]
    scans = [c for c in row["calls"] if c["location"] == "17A0:00AC"]
    assert len(guards) == len(scans)
    observations = []
    for guard, scan in zip(guards, scans):
        faction, x, y, radius = scan["args"][:4]
        faction &= 255
        assert faction == loser
        matching = []
        total = 0
        for i in range(word(before["lists"], 2)):
            pool = word(before["lists"], 38 + i*2)
            sh = before["ship"][pool*78:(pool+1)*78]
            if sh[0] != faction or sh[30] & 128:
                continue
            dx, dy = x-word(sh, 52), y-word(sh, 54)
            if math.isqrt((dx*dx + dy*dy) << 32) > radius << 16:
                continue
            matching.append(pool)
            if sh[4] in (0, 1):
                total = min(65535, total + word(sh, 8))
        assert scan["return_ax"] == len(matching), row["name"]
        assert guard["return_ax"] == total // 100, row["name"]
        queries += 1
        observations.append({"query_args": scan["args"][:4], "matched_pools": matching,
                             "query_count": scan["return_ax"], "guard": guard["return_ax"]})
    cases.append({"name": row["name"], "variant": row["guard"], "injections": row["injections"],
                  "inputs": row["inputs"], "steps": row["steps"], "seed_after_hex": row["seed_after_hex"],
                  "guards": observations, "before_sha256": {k: sha(v) for k, v in before.items()},
                  "after_sha256": {k: sha(bytes.fromhex(v)) for k, v in row["after"].items()}})
assert queries == 16
near, hidden = original["rows"][:2]
effects = {k: sum(x != y for x, y in zip(bytes.fromhex(near["after"][k]), bytes.fromhex(hidden["after"][k])))
           for k in near["after"]}
assert effects["ship"] == 20 and effects["work"] == 2

models = []
for name, count in (("base", 12), ("guard", 7)):
    path = f"workplace/finish79-r90-probe-v2-{name}.json"
    rows = json.loads(Path(path).read_bytes())
    assert len(rows) == count
    assert all(not r["differences"] for r in rows)
    models.extend({"source": path, "name": r["name"], "differences": r["differences"],
                   "hits": r["hits"], "captured": r["captured"], "seed": r["seed"],
                   "planet_sha256": sha(bytes.fromhex(r["planet_hex"]))} for r in rows)
assert len(models) == 19
assert Path("workplace/finish79-r90-probe-v1.json").read_bytes() == Path("workplace/finish79-r90-probe-v2-base.json").read_bytes()

semantics = {
    "17A0:00D9": "單位查詢讀DS:9896 word筆數、DS:98BA池word清單",
    "17A0:00FF": "查詢依SHIP+00陣營byte篩選，不讀艦隊士氣bit0",
    "17A0:0106": "SHIP+1E bit7排除查詢",
    "17A0:010F": "查詢距離讀SHIP+36現在Y",
    "17A0:0115": "查詢距離讀SHIP+34現在X",
    "17A0:0133": "距離大於半徑才排除；等於仍包含",
    "203A:0544": "守軍消費端只加艦種0／1艘數",
    "203A:0577": "守軍艘數加總飽和65535",
    "203A:0589": "守軍加總最後除100",
    "20B4:0EA2": "反擊呼叫守軍消費端203A:04D1",
    "20B4:0EAC": "守軍加防衛力以word回繞，後除要塞20／一般10",
    "203A:06BC": "實扣艘數提交SHIP+08",
    "203A:06C3": "修理上限減實扣平方除滿編的低word",
    "203A:06DE": "每次損失後呼叫1261:0067，即時重算艦隊",
    "297B:0915": "定位時FLEET+5A先清FF；其後終態來源未完全審查",
    "297B:09D2": "定位時索敵池SHIP+1E寫FF，不能只改一般分隊位置",
    "297B:0BA7": "定位朝向寫FLEET+4A，接著寫+4C及+4B",
    "297B:0BCC": "定位目的地Y寫FLEET+66；+64同樣同步X",
    "1261:008C": "損失即時重算呼叫297B:10DE全陣營維持費",
    "1261:00A8": "全陣營維持費高word回寫DS:8F72加陣營索引",
}
ida = []
ida_paths = [f"workplace/ida/finish79-r90-{seg}.json" for seg in ("0DEA", "1261", "17A0", "203A", "203A-full", "20B4", "297B")]
for path in ida_paths:
    export = json.loads(Path(path).read_bytes())
    assert export["kernel_version"] == "9.4" and export["uid"] == 1000
    for ins in export["instructions"]:
        if ins["original_location"] in semantics:
            ins.update(grade="proven", semantic=semantics[ins["original_location"]], warning="",
                       evidence=[path, "workplace/finish79-r90-guard-a.json"])
    ida.append(export)
assert len(set(ins["original_location"] for export in ida for ins in export["instructions"] if ins["grade"] == "proven")) == len(semantics)
paths = ["org_game/extracted/GIN3.COM", "workplace/oracle/cleanup79-r88-entry-a/entry.state",
         "workplace/oracle/cleanup79-r88-entry-a/mem-in.bin", "docs/internals/data/battle-finish.json",
         "workplace/finish79-r89-native.go", "workplace/finish79-r89-a.json", "workplace/finish79-r89-b.json",
         "workplace/finish79-r90-guard-native.go", "workplace/finish79-r90-guard-a.json", "workplace/finish79-r90-guard-b.json",
         "workplace/finish79-r90-probe-v1-test.go", "workplace/finish79-r90-probe-v1.json",
         "workplace/finish79-r90-probe-v2-base.json", "workplace/finish79-r90-probe-v2-guard.json",
         "tools/scripts/battle_finish_guard_prepare.py", "tools/battle_finish_guard_oracle.sh",
         "tools/battle_finish_guard_native_run.sh", "tools/scripts/battle_finish_attack_ida_prepare.py",
         "tools/scripts/battle_finish_attack_probe_test.go", "tools/battle_finish_attack_probe.sh",
         "tools/scripts/battle_finish_attack_review.py", "internal/rules/strategicassault.go",
         "internal/rules/fleet.go", "internal/rules/save.go", "internal/rules/planetassault.go",
         "docs/internals/data/battle-formation-slots.json", "docs/internals/data/ship-kind-table.json",
         "workplace/ida/finish79-r90-ida.py", "workplace/finish79-r90-standard.log"]
paths += ida_paths + [f"workplace/ida/finish79-r90-{seg}.com" for seg in ("0DEA", "1261", "17A0", "203A", "20B4", "297B")]
paths += [str(p) for p in sorted(Path("internal/rules").glob("*.go"))]
paths += ["go.mod", "go.sum"]
paths = list(dict.fromkeys(paths))
result = {"schema": "logh3.battle-finish-attack-review.v1", "issue": 79,
          "parent_head": "55815840a6d5342d41d595df61c7450030c693ea",
          "scope": "19案暫存規則模型：完整行星、艘數、位置、艦隊士氣／旗標、計分、功績及seed；7案原版守軍16查詢。非完整原池、production或正常玩家驗收。",
          "limits": ["FLEET／SHIP窄回寫、軍事費與完整六表尚未模型對拍", "定位後FLEET+5A終態寫入來源尚未完全審查",
                     "模型地形Passable為nil，本批位置符合不推廣所有擋路情形", "守軍飽和案65535艘是外部格式邊界診斷，不稱一般編成",
                     "110及完整062保持DRAFT；正式報告、保存與正常玩家勝負仍缺"],
          "native": {k: v for k, v in original.items() if k != "rows"}, "sources": [source(p) for p in paths],
          "models": models, "guard_cases": cases, "guard_effect_bytes": effects, "ida": ida,
          "totals": {"model_cases": 19, "guard_native_cases": 7, "guard_queries": queries,
                     "ida_instruction_records": sum(len(x["instructions"]) for x in ida), "proven_semantic_locations": len(semantics)}}
dest = Path("docs/internals/data/battle-finish-attack.json")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("19案核心模型、7案非零守軍、16查詢；來源", len(paths), "SHA-256", sha(dest.read_bytes()))
