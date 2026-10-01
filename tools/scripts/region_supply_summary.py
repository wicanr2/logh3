"""保存 #84 有限補給鏈及 #85 完整表差異；入口 docs/re/353。

tools/py.sh tools/scripts/region_supply_summary.py
原版收據必須先各重生兩輪；不從未分級散文匯入語意。
"""
import hashlib
import json
import pathlib


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


base = pathlib.Path("workplace/oracle")
replays = []
for batch in ("supply84-stationed", "supply84-gui"):
    first, second = base / batch, base / (batch + "-r2")
    inputs = json.loads((first / "inputs.json").read_text())
    repeated = json.loads((second / "inputs.json").read_text())
    assert inputs["golem_head"] == repeated["golem_head"]
    assert inputs["original_seed_hex"] == inputs["remake_seed_hex"] == "0xd3dc167e"
    assert not inputs["golem_working_tree"] and not repeated["golem_working_tree"]
    assert inputs["parent"] == repeated["parent"]
    assert inputs["memory"] == repeated["memory"]
    names = [m["stage"] + ".bin" for m in inputs["memory"]]
    names += ["screen.bin", "screen.png", "pal.bin"]
    hashes = {}
    for name in names:
        assert (first / name).read_bytes() == (second / name).read_bytes(), (batch, name)
        hashes[name] = digest(first / name)
    replays.append({"batch": batch, "repeat": batch + "-r2", "grade": "proven",
                    "scope": "指定完整檢查點、輸入及階段擷取；不宣稱所有回合",
                    "inputs": inputs, "flags": (first / "flags.txt").read_text().splitlines(),
                    "hashes": hashes})

folder = base / "supply84-stationed"
memory = (folder / "command.bin").read_bytes()
diffs = json.loads((folder / "normal-table-differences.json").read_text())
raw_differences = {}
for table, start, size in (("stellar", 0x5F2CC, 126), ("planet", 0x53132, 82)):
    remake = (folder / ("remake-" + table + ".bin")).read_bytes()
    raw_differences[table] = [{"record": i // size, "offset_hex": hex(i % size),
                               "original_linear_hex": hex(start + i), "original": memory[start + i],
                               "remake": remake[i], "grade": "unknown", "followup_issue": 85}
                              for i in diffs[table]]
    # 比較範圍由 #84 預先定義，每筆全部七種艦艇與兩種補給，不採總量代替。
    lo, hi = (0x30, 0x78) if table == "stellar" else (0x20, 0x44)
    for record in range(len(remake) // size):
        a = record * size
        assert remake[a + lo:a + hi] == memory[start + a + lo:start + a + hi]

gui = base / "supply84-gui"
stats = json.loads((gui / "stats.json").read_text())
assert stats["index_different"] == stats["regions"]["full"]["different"] == 0
report = {
    "source_exe": "GIN3.COM",
    "source_exe_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
    "bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
    "address_space": "dosgolem 執行期線性位址；程式原始定位另保留 CS:IP",
    "replays": replays,
    "callsite": {"original_location": "0654:0015", "linear_hex": "0x6555",
                 "original_bytes_hex": memory[0x6555:0x655A].hex(" "), "return_location": "0654:001A",
                 "semantic": "呼叫 2B5F:02B2；下一座位前僅回充星域池", "grade": "proven",
                 "source": "refill-before.bin／refill-after.bin、report.txt 的命中及規格 039 證據審查"},
    "stock_validation": {"grade": "proven", "scope": "S1 帝國首個實行回合，與明示同局面注入",
                         "stellar_bytes_compared": 32 * 72, "planet_bytes_compared": 81 * 36,
                         "different": 0, "test": "TestRegionSupplyNormalTurnMatchesOriginal",
                         "full_tables_equal": False},
    "full_table_differences": raw_differences,
    "remake_hashes": {name: digest(folder / name) for name in
                      ("remake-stellar.bin", "remake-planet.bin", "normal-table-differences.json")},
    "gui": {"stats": stats, "metadata": json.loads((gui / "remake-meta.json").read_text())},
    "limitations": ["完整正常回合表尚有星域 18／行星 15 bytes 差異，由 #85 追查。",
                    "保留使用者選定的負補給缺口夾零；不宣稱該情況原版一致。",
                    "固定種子及局部結果不證明整段骰序或所有回合一致。"],
}
assert report["callsite"]["original_bytes_hex"] == "9a b2 02 5f 2b"
path = pathlib.Path("docs/internals/data/region-supply-turn-timing.json")
path.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print(path, "兩輪重播一致；庫存零差異；完整表差異", {k: len(v) for k, v in raw_differences.items()})
