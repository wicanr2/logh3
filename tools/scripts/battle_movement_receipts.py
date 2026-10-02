"""研究371的原版收據核對與摘要；在 tools/py.sh 容器內執行。"""
import hashlib
import json
import struct
from pathlib import Path

ROOT = Path("/work")
WORK = ROOT / "workplace"
HEAD = "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
STEP = 8 << 16


def sha(data):
    return hashlib.sha256(data).hexdigest()


def word(raw, off):
    return struct.unpack_from("<H", raw, off)[0]


def budget(raw, off):
    return struct.unpack_from("<i", raw, off)[0]


def refill(old, speed):
    return (old if 0 < old < STEP else 0) + (speed << 19)


def put_budget(raw, off, value):
    struct.pack_into("<i", raw, off, value)


receipts = {}
memories = {}
for mode, counter, stop in [
    ("entry", 1213409021, "181B:00FB"),
    ("init", 1213433817, "1C03:006B"),
    ("end", 1226638239, "181B:0263"),
]:
    a = WORK / f"oracle/battle79-move-{mode}-a"
    b = WORK / f"oracle/battle79-move-{mode}-b"
    flags = (a / "flags.txt").read_bytes()
    flags_b = (b / "flags.txt").read_bytes()
    # 兩次必須分目錄；只替換輸出目的地，輸入、poke與步數不正規化。
    expected_b = flags.replace(f"/out/oracle/{a.name}/".encode(), f"/out/oracle/{b.name}/".encode())
    assert expected_b == flags_b, "除輸出路徑外的完整旗標不一致"
    assert HEAD[:7] in (a / "meta.txt").read_text()
    report = (a / "report.txt").read_text()
    assert f"CS:IP ＝ {stop}" in report, "階段邊界不符"
    assert "改記憶體 lin:5E00C（5E00C）4 bytes" in report
    assert f"#{counter} 傾印記憶體 lin:0（00000）1048576 bytes" in report
    hashes = {}
    for name, length in [
        ("mem-in.bin", 1 << 20), ("mem-out.bin", 1 << 20),
        ("screen.bin", 640 * 480), ("pal.bin", 768), ("screen.png", None),
    ]:
        data = (a / name).read_bytes()
        assert data == (b / name).read_bytes(), f"{mode}/{name} 重播不一致"
        assert length is None or len(data) == length
        assert (a / name).stat().st_uid == (b / name).stat().st_uid == 1000
        hashes[name] = sha(data)
    memories[mode] = (a / "mem-out.bin").read_bytes()
    receipts[mode] = {
        "runs": [str(x.relative_to(ROOT)) for x in [a, b]],
        "counter": counter, "stop_after_one_instruction": stop,
        "flags": [flags.decode(), flags_b.decode()], "hashes": hashes,
        "machine_state_hashes": [sha((x / "machine.state").read_bytes()) for x in [a, b]],
        "grade": "proven", "scope": "原版相同局面的兩次重播；尚無remake比較",
    }

assert memories["entry"] == (WORK / "oracle/battle79-move-end-a/mem-in.bin").read_bytes()
assert sha(memories["entry"]) == "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d"
ds = 0x54B30
pool = word(memories["entry"], ds + 0xCFB6) * 16 + word(memories["entry"], ds + 0xCFB4)
fleets = []
units = []
for i in range(word(memories["entry"], ds + 0x9894)):
    index = memories["entry"][ds + 0x9898 + i]
    at = ds + 0xC074 + index * 122
    records = {k: v[at:at + 122] for k, v in memories.items()}
    assert budget(records["init"], 0x56) == budget(records["entry"], 0x52)
    row = {"fleet_index": index, "list_order": i, "linear_hex": hex(at)}
    row.update({k + "_hex": v.hex() for k, v in records.items()})
    fleets.append(row)
    indices = [word(records["entry"], 0x10 + j * 4) for j in range(records["entry"][15])]
    indices.append(word(records["entry"], 0x38))
    for j, ship in enumerate(indices):
        at = pool + ship * 78
        records_u = {k: v[at:at + 78] for k, v in memories.items()}
        scout = j == len(indices) - 1
        old = budget(records_u["entry"], 0x24)
        expected = old if scout and records["entry"][0x3A] == 255 else refill(old, records_u["entry"][0x21])
        assert budget(records_u["init"], 0x24) == expected, f"原版局面初始化不符 ship={ship}"
        row_u = {"fleet_index": index, "ship_index": ship, "scout": scout, "linear_hex": hex(at)}
        row_u.update({k + "_hex": v.hex() for k, v in records_u.items()})
        units.append(row_u)
assert len(fleets) == 7 and len(units) == 55
for data in memories.values():
    assert data[ds + 0x94DC:ds + 0x94E0].hex() == "78563412", "種子變動"

probe_raw = (WORK / "battle79-move-probe-a.json").read_bytes()
assert probe_raw == (WORK / "battle79-move-probe-b.json").read_bytes()
probe = json.loads(probe_raw)
assert len(probe["rows"]) == 20
for row in probe["rows"]:
    assert row["seed_after_hex"] == "78563412"
    if row["entry"] == "1C03:0006":
        f = bytearray.fromhex(row["fleet_before_hex"])
        put_budget(f, 0x56, budget(f, 0x52))
        assert f.hex() == row["fleet_after_hex"], "艦隊初始化出現額外欄位修改"
        for i, before in enumerate(row["units_before_hex"]):
            u = bytearray.fromhex(before)
            if i < 2 or row["scout_active"]:
                put_budget(u, 0x24, refill(budget(u, 0x24), u[0x21]))
            assert u.hex() == row["units_after_hex"][i], "分隊初始化出現額外欄位修改"
    else:
        u = bytearray.fromhex(row["units_before_hex"][0])
        arrived = word(u, 0x34) == word(u, 0x28) and word(u, 0x36) == word(u, 0x2A)
        if arrived or u[4] != 7:
            u[30] = 3
            put_budget(u, 0x24, 0)
        assert u.hex() == row["units_after_hex"][0], "自主移動停止分支出現額外修改"
        assert row["return_low_byte"] == 0

inputs = {}
entries = []
review = {
    "1C03:0006": "初始化戰場清單內艦隊額度並呼叫分隊初始化；retf=0069",
    "1C03:006A": "艦隊單次步進分派；先更新各一般分隊追隊旗標；retf=015B",
    "1C61:0005": "一般分隊／啟用索敵隊使用巡航速度與不足一步正額度累加；retf=0152",
    "1C61:0153": "第0分隊及未啟用索敵隊走0229，其餘依完整Action分派；retf=0228",
    "1C61:029C": "重算隊形集結點，再依8／12／16／24步長跟隊；retf=03A9",
    "1C61:03AA": "自主移動抵達或一般分隊額度不足時改歸隊並清額度；retf=0443",
    "1C61:0444": "以指定步長算新座標、問通行／避障，最後固定扣8.0；retf=0648",
}
specific = {
    "1C61:004E": "初始化讀巡航速度+21；沒有依戰鬥速度+22選擇",
    "1C61:009A": "0<舊額度<8.0時累加speed<<19的低word",
    "1C61:009D": "累加speed<<19的高word與進位",
    "1C61:00AC": "其餘舊額度覆蓋高word",
    "1C61:00B0": "其餘舊額度覆蓋低word",
    "1C61:03F3": "自主移動完成或一般分隊不足一步時Action設3",
    "1C61:0622": "寫入現在X+34，目的地仍為+28",
    "1C61:0629": "寫入現在Y+36，目的地仍為+2A",
    "1C61:062D": "固定扣8.0低word，未使用前面算出的距離作扣額",
    "1C61:0632": "固定扣8.0高word並扣借位",
}
exports = {}
for seg in ["1C03", "1C61"]:
    p = WORK / f"ida/re79-movement-{seg}.json"
    j = json.loads(p.read_text())
    assert j["tool"] == "IDA Pro 9.4" and j["ida_base_hex"] == "0x10200"
    inputs[seg] = j["input_sha256"]
    exports[seg] = {"path": str(p.relative_to(ROOT)), "sha256": sha(p.read_bytes()), "input_sha256": j["input_sha256"]}
    for f in j["functions"]:
        loc = f["original_location"]
        # 01D8只有既有0153內的分派區塊，不升格成函式入口。
        if loc in review:
            assert f["instructions"][0]["original_bytes_hex"] == "55"
            assert f["instructions"][-1]["original_operand"] == "retf"
            entries.append({"kind": "function", "original_location": loc,
                            "semantic": review[loc], "grade": "proven",
                            "source": ["docs/re/371-battle-movement-phase.md", exports[seg]["sha256"]]})
        for ins in f["instructions"]:
            loc = ins["original_location"]
            if loc in specific:
                entries.append({"kind": "instruction", "original_location": loc,
                                "original_bytes_hex": ins["original_bytes_hex"],
                                "semantic": specific[loc], "grade": "proven",
                                "source": ["docs/re/371-battle-movement-phase.md", exports[seg]["sha256"]]})

result = {
    "schema": "logh3.battle-movement-evidence.v1", "inputs": inputs, "entries": entries,
    "tool": "IDA Pro 9.4", "ida_base_hex": "0x10200", "exports": exports,
    "program_sha256": probe["program_sha256"], "dosgolem_head": HEAD,
    "seed": "0x12345678", "seed_linear": "lin:5E00C", "phase_receipts": receipts,
    "routine_receipt_sha256": sha(probe_raw), "routine_rows": probe["rows"],
    "fleets": fleets, "units": units,
    "scope": "原版初始化及停止條件與指定局面。完整移動仍DRAFT；沒有remake或畫面對拍通過聲明。",
}
output = ROOT / "docs/internals/data/battle-movement-phase.json"
assert output.parent.stat().st_uid == 1000
assert not output.exists() or output.stat().st_uid == 1000
output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print("三邊界雙跑相同；7艦隊／55分隊初始化、20受控案例及種子核對通過")
print("摘要", output, "SHA-256", sha(output.read_bytes()))
