"""在Docker獨立驗算固定原版清理收據；保留所有原始定位及分級。"""
import hashlib
import json
from pathlib import Path


def sha(b):
    return hashlib.sha256(b).hexdigest()


def word(b, offset):
    return int.from_bytes(b[offset:offset + 2], "little")


def source(path):
    p = Path(path)
    b = p.read_bytes()
    return {"path": path, "size": len(b), "sha256": sha(b)}


a = Path("workplace/cleanup79-r88-a.json").read_bytes()
b = Path("workplace/cleanup79-r88-b.json").read_bytes()
assert a == b and sha(a) == "80f108f6db41137c245dd52040616234779cf93349fbd7c0fae87d1aa2ef9949"
rows = json.loads(a)["rows"]
assert len(rows) == 35
cases = []
total = 0
for row in rows:
    f = bytearray.fromhex(row["before"]["fleet"])
    s = bytearray.fromhex(row["before"]["ship"])
    base = row["fleet"] * 122
    i = 0
    removed = []
    while i < f[base + 15]:
        pool = word(f, base + 16 + 4 * i)
        sh = pool * 78
        if word(s, sh + 8):
            i += 1
            continue
        assert i != 0
        removed.append(pool)
        s[sh + 1] = 255
        s[sh + 25] &= 191
        s[sh + 56:sh + 60] = bytes(4)
        f[base + 15] -= 1
        n = f[base + 15]
        off = base + 16 + 4 * i
        f[off:base + 16 + 4 * n] = f[off + 4:base + 20 + 4 * n]
    assert f == bytes.fromhex(row["after"]["fleet"]), row["name"]
    assert s == bytes.fromhex(row["after"]["ship"]), row["name"]
    for name in ("admiral", "work", "stellar", "planet"):
        assert row["before"][name] == row["after"][name]
    assert row["seed_after_hex"] == "78563412" and not row["inputs"] and not row["frames"]
    assert sum(c["location"] == "185C:020C" for c in row["calls"]) == len(removed)
    before = bytes.fromhex(row["before"]["lists"])
    after = bytes.fromhex(row["after"]["lists"])
    assert word(before, 0) == word(after, 0) == 7
    assert word(before, 2) == 55 and word(after, 2) == 55 - len(removed)
    fleet_ids = list(after[4:11])
    expected = []
    for index in fleet_ids:
        rec = f[index * 122:(index + 1) * 122]
        expected.extend(word(rec, 16 + 4 * j) for j in range(rec[15]))
        expected.append(word(rec, 56))
    # DS:98BA，與DS:9894相差38；艦隊byte清單後有兩byte間隔。
    actual = [word(after, 38 + 2 * j) for j in range(word(after, 2))]
    assert actual == expected, row["name"]
    total += len(removed)
    cases.append({"name": row["name"], "fleet": row["fleet"], "count_before": row["count"],
                  "count_after": f[base + 15], "removed_pools": removed,
                  "tables_after_sha256": {k: sha(bytes.fromhex(v)) for k, v in row["after"].items()},
                  "result": "six tables and active lists exact; seed unchanged"})
assert total == 77

ida = json.loads(Path("workplace/ida/cleanup79-r88-ida.json").read_bytes())
semantics = {
    "185C:01F6": "清理呼叫1727:0298擦小地圖點；不以函式自動名判義",
    "185C:021A": "搬移之前減FLEET+0F的byte組數",
    "185C:024F": "以新組數停止搬移；不清尾槽",
    "185C:0254": "讀後組引用的完整池word",
    "185C:025A": "寫前組引用的完整池word",
    "185C:0267": "搬引用的第三個byte",
    "185C:0275": "搬引用的第四個byte",
}
for ins in ida["instructions"]:
    if ins["original_location"] in semantics:
        ins.update(grade="proven", semantic=semantics[ins["original_location"]], warning="",
                   evidence=["workplace/ida/cleanup79-r88-ida.json", "workplace/cleanup79-r88-a.json"])
assert sum(x["grade"] == "proven" for x in ida["instructions"]) == 7
paths = ["org_game/extracted/GIN3.COM", "tools/battle_cleanup_entry_oracle.sh",
         "tools/scripts/battle_cleanup_prepare.py", "tools/battle_cleanup_native_run.sh",
         "tools/battle_cleanup_oracle.sh", "tools/scripts/battle_cleanup_review.py",
         "workplace/promotion79-r86-global.go", "workplace/cleanup79-r88-native.go",
         "workplace/cleanup79-r88-a.json", "workplace/cleanup79-r88-b.json",
         "workplace/ida/cleanup79-r88-185C.com", "workplace/ida/cleanup79-r88-ida.py",
         "workplace/ida/cleanup79-r88-ida.json", "workplace/ida/cleanup79-r88-ida.log"]
for side in ("a", "b"):
    prefix = f"workplace/oracle/cleanup79-r88-entry-{side}/"
    paths.extend(prefix + name for name in ("entry.state", "mem-in.bin", "meta.txt", "report.txt"))
assert Path("workplace/oracle/cleanup79-r88-entry-a/mem-in.bin").read_bytes() == Path("workplace/oracle/cleanup79-r88-entry-b/mem-in.bin").read_bytes()
result = {"schema": "logh3.battle-cleanup-review.v1", "issue": 79,
          "scope": "35案固定8-1入口明示死亡；純獨立模型，非正常玩家全流程",
          "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": False,
          "cases": cases, "removed_count": total, "ida": ida, "sources": [source(p) for p in paths]}
dest = Path("docs/internals/data/battle-cleanup.json")
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("原版35案完整六表及有效清單相同，77組作廢；", dest, sha(dest.read_bytes()))
