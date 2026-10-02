"""083證據匯出：核兩次原版35邊界，保存完整原池記錄與分級定位。"""
import hashlib
import json
from pathlib import Path


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, d):
    p.write_text(json.dumps(d, ensure_ascii=False, indent=2) + "\n")


d = json.loads(Path("docs/internals/data/battle-movement-phase.json").read_text())
assert sha(Path("docs/internals/data/battle-movement-phase.json")) == "b920eb42ccc70c8c447a9a2bd53a5e40f9efae5e807929fe4972fa3cc5fbe516"
a = Path("workplace/oracle/battle79-phase-r34-a")
b = Path("workplace/oracle/battle79-phase-r34-b")
names = ["entry"] + [f"loop-{i:02d}" for i in range(33)] + ["end"]
rows = []
for n in names:
    for ext in [".bin", ".png", "-mem.bin"]:
        assert (a / (n + ext)).read_bytes() == (b / (n + ext)).read_bytes(), n + ext
    mem = (a / (n + "-mem.bin")).read_bytes()
    assert len(mem) == 1048576
    assert mem[0x5E00C:0x5E010].hex() == "78563412"
    assert mem[0x5E724:0x5E728].hex() == "20045802"
    rows.append({"name": n, "phase": "repair" if n == "entry" else "movement",
                 "hashes": {ext: sha(a / (n + ext)) for ext in [".bin", ".png", "-mem.bin"]},
                 "fleets": [{"index": r["fleet_index"], "hex": mem[int(r["linear_hex"], 16):int(r["linear_hex"], 16)+122].hex()} for r in d["fleets"]],
                 "units": [{"index": r["ship_index"], "hex": mem[int(r["linear_hex"], 16):int(r["linear_hex"], 16)+78].hex()} for r in d["units"]]})
assert rows[0]["hashes"]["-mem.bin"] == d["phase_receipts"]["entry"]["hashes"]["mem-in.bin"]
assert rows[-1]["hashes"]["-mem.bin"] == d["phase_receipts"]["end"]["hashes"]["mem-out.bin"]
out = {"schema": "logh3.battle-phase-frames.v1",
       "scope": "same-state fixed Turn6 / 7 fleets / 55 units / camera(66,30) / 35 boundaries; diagnostic state injection in remake; normal path separate",
       "program_sha256": d["program_sha256"], "dosgolem_head": d["dosgolem_head"],
       "parent_evidence_sha256": sha(Path("docs/internals/data/battle-movement-phase.json")),
       "parent_phase_receipts": d["phase_receipts"]["entry"],
       "snapshot_sha256": sha(Path("workplace/oracle/battle79-move-entry-a/machine.state")),
       "seed": "0x12345678", "camera": [66, 30], "canvas": [640, 480],
       "runs": [{"path": str(p), "meta": (p / "meta.txt").read_text(),
                 "meta_sha256": sha(p / "meta.txt"), "report_sha256": sha(p / "report.txt")} for p in [a, b]],
       "frames": rows}
write(Path("docs/internals/data/battle-phase-frames.json"), out)

# 所有語意逐筆與IDA原始bytes核對；未審查指令仍unknown，不批次推論。
meanings = {
    "181B:003F": "補給、修理及防衛力恢復入口；003F..0082，保留原名",
    "181B:0045": "載入DS:3CD4補給階段文字",
    "181B:0051": "印補給階段到DS:9890/9892狀態窗",
    "181B:0059": "文字後執行1930:019D補給",
    "181B:005F": "載入DS:3CE2修理階段文字",
    "181B:006B": "印修理階段到同一狀態窗",
    "181B:0073": "文字後執行1AAB:006D修理",
    "181B:00FA": "移動階段入口；00FA..021B",
    "181B:0112": "載入DS:3CF0移動階段文字",
    "181B:011E": "初始化額度前印移動階段",
    "181B:0126": "文字後呼叫1C03:0006初始化額度",
    "3617:000A": "視窗格式文字入口；000A..0039",
    "3617:001F": "呼叫3617:0059填寫視窗文字緩衝",
    "3617:002E": "呼叫341B:02E5重畫視窗",
}
entries = []
exports = {}
for seg, filename in [("181B", "re79-phase-r34-181B.json"), ("3617", "re79-phase-r34-3617-short.json")]:
    p = Path("workplace/ida") / filename
    e = json.loads(p.read_text())
    exports[seg] = {"path": str(p), "sha256": sha(p), "input_sha256": e["input_sha256"]}
    for f in e["functions"]:
        for row, kind in [(f, "function")] + [(i, "instruction") for i in f["instructions"]]:
            loc = row["original_location"]
            if loc not in meanings:
                continue
            r = {"kind": kind, "original_location": loc, "semantic": meanings[loc],
                 "grade": "proven", "source": ["docs/re/375-battle-phase-panel.md", sha(p)]}
            if kind == "instruction":
                raw = bytes.fromhex(row["original_bytes_hex"])
                linear = int(seg, 16)*16 + int(loc.split(":")[1], 16)
                assert mem[linear:linear+len(raw)] == raw, loc
                r["original_bytes_hex"] = row["original_bytes_hex"]
            entries.append(r)
    # 維持定位／原始name／operand。以受版控索引自動附加語意。
    byloc = {r["original_location"]: r for r in entries if r["kind"] == "instruction"}
    funcs = {r["original_location"]: r for r in entries if r["kind"] == "function"}
    for f in e["functions"]:
        for row, mapping in [(f, funcs)] + [(i, byloc) for i in f["instructions"]]:
            if row["original_location"] in mapping:
                r = mapping[row["original_location"]]
                row.update({k: r[k] for k in ["semantic", "grade", "source"]})
                row.pop("warning", None)
    e["semantic_index"] = "docs/internals/data/battle-phase-semantics.json"
    write(Path("workplace/ida") / (filename.removesuffix(".json") + "-reviewed.json"), e)
strings = []
for key, off in [("supply", 0x3CD4), ("repair", 0x3CE2), ("movement", 0x3CF0)]:
    raw = mem[0x54B30+off:0x54B30+off+14]
    assert raw[0] == 12 and raw[-1] == 0 and raw[9:13] == b"    "
    strings.append({"phase": key, "original_location": f"DS:{off:04X}", "raw_hex": raw.hex(),
                    "text": raw[1:9].decode("big5"),
                    "codes": [int.from_bytes(raw[i:i+2], "big") for i in range(1, 9, 2)],
                    "grade": "proven", "source_ram_sha256": rows[-1]["hashes"]["-mem.bin"]})
write(Path("docs/internals/data/battle-phase-semantics.json"), {
    "schema": "logh3.semantic-index.v1", "tool": "IDA Pro 9.4", "ida_base_hex": "0x10200",
    "address_space": "COM offset0=IDA 0x10200; original CS:IP and dosgolem linear separate; DS base 0x54B30",
    "program_sha256": d["program_sha256"], "exports": exports, "entries": entries, "strings": strings,
    "resolutions": [{"immutable_key": "PC DOS GIN3.COM 181B:011E/0126",
                     "semantic": "先印移動階段，再初始化額度", "grade": "proven",
                     "evidence": "docs/re/375-battle-phase-panel.md",
                     "older_spec": "docs/spec/080-battle-movement-phase.md",
                     "required_correction": "181B:00FA在011E先印移動文字、0126才初始化"}]})
assert "181B:0112" in Path("docs/re/375-battle-phase-panel.md").read_text()
assert "181B:00FA在011E先印移動文字、0126才初始化" in Path("docs/spec/080-battle-movement-phase.md").read_text()
assert "375-battle-phase-panel.md" in Path("docs/spec/083-battle-phase-panel.md").read_text()
print(f"35邊界／105檔雙跑相同；{len(entries)}筆分級定位、3字串已核。")
