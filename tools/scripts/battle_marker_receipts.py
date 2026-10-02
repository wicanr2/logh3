"""#79標記原版收據、可丟棄模型與受審查索引；來源見研究374。

tools/py.sh tools/scripts/battle_marker_receipts.py
只合併下方逐項審查定位，不從散文自動擷取語意。
"""
import hashlib
import json
import os
from pathlib import Path
import struct


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, data):
    assert p.parent.is_dir() and p.parent.stat().st_uid == os.getuid()
    if p.exists():
        assert p.stat().st_uid == os.getuid()
    p.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
    print(p, sha(p))


work, output = Path("workplace"), Path("docs/internals/data")
# 新證據的定位與舊文件勘誤必須同時可回查，不由散文自動匯入語意。
for name, required in {
    "docs/re/85-units-are-sprite-tiles.md": "25DB:03E9",
    "docs/re/154-minimap-marker-not-a-trail.md": "規格082",
    "docs/spec/018-battle-resolution.md": "25DB:03E9",
    "docs/internals/38-battle-units.md": "25DB:03E9",
}.items():
    text = Path(name).read_text()
    assert "374-battle-marker-lifecycle.md" in text and required in text, name
entry = work / "oracle/battle79-move-entry-a/mem-in.bin"
assert sha(entry) == "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d"
assert entry.read_bytes()[0x26198:0x261A0].hex() == "cb55565783ec0689"
a, b = work / "battle79-marker-r33-a.json", work / "battle79-marker-r33-b.json"
assert a.read_bytes() == b.read_bytes()
assert sha(a) == "966fba906abade89332678243984ab11eb41976def7e0cc5a1d797d0d515371d"
probe = json.loads(a.read_text())
assert len(probe["rows"]) == 96
for row in probe["rows"]:
    before = bytearray.fromhex(row["before_hex"])
    if row["forced"] == 0 and row["detail"] == 0:
        if row["enabled"] and before[0x19] & 0xC0 == 0xC0:
            before[0x19] &= 0xBF
        before[0x19] &= 0x7F
    else:
        before[0x19] |= 0x80
        if row["enabled"] and not before[0x19] & 0x40:
            before[0x19] = 0xC0 | (5, 9, 4)[row["faction"]]
            before[0x1A:0x1E] = before[0x34:0x38]
    assert before.hex() == row["after_hex"], row["name"]
    assert row["seed_after_hex"] == "78563412"

na, nb = (work / ("oracle/battle79-marker-r33-fixed-" + k) for k in ("a", "b"))
hashes = {}
for name in ("mem.bin", "screen.bin", "pal.bin", "screen.png"):
    assert (na / name).read_bytes() == (nb / name).read_bytes()
    hashes[name] = sha(na / name)
assert hashes["mem.bin"] == "256a6fe846c6e75746f690be38f3b53a2c0e6af6e417c495d78e9a33ef399e03"
mem = (na / "mem.bin").read_bytes()
assert len(mem) == 1048576
ds = 0x54B30
off, seg = struct.unpack_from("<HH", mem, ds + 0xCFB4)
pool = seg * 16 + off
units = []
for fleet in mem[ds + 0x9898:ds + 0x9898 + mem[ds + 0x9894]]:
    f = mem[ds + 0xC074 + fleet * 122:ds + 0xC074 + (fleet + 1) * 122]
    slots = [(j, struct.unpack_from("<H", f, 0x10 + j * 4)[0]) for j in range(f[15])]
    slots.append((-1, struct.unpack_from("<H", f, 0x38)[0]))
    for squadron, slot in slots:
        r = mem[pool + slot * 78:pool + (slot + 1) * 78]
        flags = 0xC5 if r[0] == 0 and squadron >= 0 else 0x3F
        assert r[0x19] == flags and r[0x17] == 0
        if flags == 0xC5:
            assert r[0x1A:0x1E] == r[0x34:0x38]
        units.append({"fleet": fleet, "pool": slot, "squadron": squadron, "record_hex": r.hex()})
assert len(units) == 55 and sum(bytes.fromhex(u["record_hex"])[0x19] == 0xC5 for u in units) == 8
for p in (na, nb):
    report = (p / "report.txt").read_text()
    assert "26E5:00AC 執行時的暫存器（前 1 次）" in report
    assert "1727:02E2 執行時的暫存器（前 55 次）" in report

inputs = {seg: sha(work / ("ida/re79-marker-r33-" + seg + ".com")) for seg in ("1727", "26E5", "297B")}
review = {
    "1727:0207": "bit7有／bit6無／地圖啟用才點上，記顏色與現在位置快照",
    "1727:0298": "依C0狀態及啟用條件擦標記，清bit6",
    "1727:02E2": "force或Detail決定bit7開關，96案完整記錄互證",
    "1727:0324": "世界座標除16加小地圖實體原點，呼叫XOR點",
    "26E5:00AC": "逐戰場一般分隊與索敵隊，依viewer及全域force條件更新圖示和標記",
    "297B:08D8": "佈陣時重設一般及索敵隊Detail與標記旗標；其餘控制流保持原定位",
}
instruction_review = {
    "297B:097F": ("26 c6 47 19 3f", "一般分隊標記旗標重設3F"),
    "297B:09E1": ("26 c6 44 19 3f", "未放出索敵隊標記旗標重設3F"),
    "26E5:0149": ("f6 06 1c 8f 01", "全域8F1C bit0允許force；其他位元未知"),
    "26E5:0188": ("26 80 7f 3a ff", "ScoutMode=FF使索敵force為0"),
    "1727:026C": ("0c c0", "點上後記bit7／bit6"),
    "1727:0271": ("26 88 47 19", "寫入標記旗標與低4位顏色"),
    "1727:027E": ("26 89 44 1a", "寫入現在X快照"),
    "1727:028E": ("26 89 47 1c", "寫入現在Y快照"),
    "1727:02D8": ("26 80 27 bf", "擦除後只清bit6"),
    "1727:031C": ("26 80 67 19 7f", "force=0且Detail=0清bit7"),
}
entries, exports = [], {}
for seg in inputs:
    p = work / ("ida/re79-marker-r33-" + seg + ".json")
    report = json.loads(p.read_text())
    assert report["input_sha256"] == inputs[seg] and report["tool"] == "IDA Pro 9.4"
    exports[str(p)] = sha(p)
    for f in report["functions"]:
        loc = f["original_location"]
        if loc in review:
            entries.append({"kind": "function", "original_location": loc, "semantic": review[loc], "grade": "proven", "source": "docs/re/374-battle-marker-lifecycle.md"})
        for i in f["instructions"]:
            loc = i["original_location"]
            if loc in instruction_review:
                raw, meaning = instruction_review[loc]
                assert i["original_bytes_hex"] == raw
                entries.append({"kind": "instruction", "original_location": loc, "original_bytes_hex": raw, "semantic": meaning, "grade": "proven", "source": "docs/re/374-battle-marker-lifecycle.md"})
assert len(entries) == 16
write(output / "battle-marker-semantics.json", {"schema": "logh3.semantic-index.v1", "inputs": inputs, "entries": entries})
write(output / "battle-markers.json", {"schema": "logh3.battle-markers.v1", "program_sha256": probe["program_sha256"], "dosgolem_head": probe["dosgolem_head"], "primitive_receipt_sha256": sha(a), "primitive_scope": probe["scope"], "primitive_rows": probe["rows"], "normal": {"classification": "原版正常玩家路徑，未宣稱remake比較", "hashes": hashes, "flags": [(p / "flags.txt").read_text() for p in (na, nb)], "metadata_sha256": [sha(p / "meta.txt") for p in (na, nb)], "seed_before": "0x12345678", "seed_after": hex(struct.unpack_from("<I", mem, ds + 0x94DC)[0]), "units": units}, "ida_exports": exports})
print("96案及正常55筆雙跑核對通過；6函式／10指令逐項分級")
