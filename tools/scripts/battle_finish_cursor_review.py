"""凍結占領報告前後資料與原指令；補110可暫停續接契約。"""
from pathlib import Path
import hashlib
import json


def sha(b):
    return hashlib.sha256(b).hexdigest()


a = Path("workplace/finish79-r91-cursor-a.json").read_bytes()
b = Path("workplace/finish79-r91-cursor-b.json").read_bytes()
assert a == b
original = json.loads(a)
assert original["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
baseline = {r["name"]: r for r in json.loads(Path("workplace/finish79-r89-a.json").read_bytes())["rows"]}
cases = []
for row in original["rows"]:
    old = baseline[row["name"]]
    assert row["before"] == old["before"] and row["after"] == old["after"]
    assert row["seed_after_hex"] == old["seed_after_hex"] and row["ax"] == old["ax"]
    for c in row["calls"]:
        if c["location"] != "185C:0B18":
            continue
        assert c["tables"] == c["return_tables"] and c["seed_hex"] == c["return_seed_hex"]
        fleet = (c["args"][0] - 0xC074)//122
        assert fleet == 31
        rec = bytes.fromhex(c["tables"]["fleet"])[fleet*122:(fleet+1)*122]
        adm, deputy = rec[4], rec[5]
        before = bytes.fromhex(c["tables"]["admiral"])
        after = bytes.fromhex(row["after"]["admiral"])
        increases = {}
        for index in (adm, deputy):
            o = index*28+2
            increases[str(index)] = (int.from_bytes(after[o:o+4], "little") - int.from_bytes(before[o:o+4], "little")) & 0xFFFFFFFF
        assert increases[str(adm)] == 3300 and increases[str(deputy)] == 1650
        cases.append({"name": row["name"], "fleet": fleet, "args": c["args"],
                      "entry_step": c["step"], "return_location": c["return_location"],
                      "seed_hex": c["seed_hex"], "post_callback_merit": increases,
                      "entry_sha256": {k: sha(bytes.fromhex(v)) for k, v in c["tables"].items()}})
assert len(cases) == 1
ida_path = "workplace/ida/finish79-r90-20B4.json"
ida = json.loads(Path(ida_path).read_bytes())
semantics = {"20B4:17A4": "攻擊量功績在攻占判斷前提交",
             "20B4:1878": "行星換手與占領計分後呼叫報告callback；此處須可暫停",
             "20B4:188C": "callback返回後才加行星上限功績，副官加半量"}
instructions = []
for ins in ida["instructions"]:
    if ins["original_location"] in semantics:
        ins.update(grade="proven", semantic=semantics[ins["original_location"]], warning="",
                   evidence=[ida_path, "workplace/finish79-r91-cursor-a.json"])
        instructions.append(ins)
assert len(instructions) == 3
paths = ["docs/internals/data/battle-finish-raw.json", "workplace/finish79-r91-native.go",
         "workplace/finish79-r91-cursor-native.go", "workplace/finish79-r91-cursor-a.json",
         "workplace/finish79-r91-cursor-b.json", "workplace/finish79-r89-a.json", ida_path,
         "tools/scripts/battle_finish_cursor_prepare.py", "tools/battle_finish_cursor_oracle.sh",
         "tools/battle_finish_cursor_native_run.sh", "tools/scripts/battle_finish_cursor_review.py",
         "workplace/finish79-r91-standard.log"]
result = {"schema": "logh3.battle-finish-cursor-review.v1", "issue": 79,
          "parent_head": "49c6236dd6f8fd1cb57cebddb72c2036ebeab61d",
          "scope": "固定兩案原版雙跑；一個真正占領callback前後七張表與seed不變，返回後功績已核",
          "limits": ["尚未實作正式cursor／Game，不稱正常玩家或renderer通過"],
          "sources": [{"path": p, "size": Path(p).stat().st_size, "sha256": sha(Path(p).read_bytes())} for p in paths],
          "cases": cases, "instructions": instructions}
dest = Path("docs/internals/data/battle-finish-cursor.json")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("callback unchanged; merit follows return; SHA-256", sha(dest.read_bytes()))
