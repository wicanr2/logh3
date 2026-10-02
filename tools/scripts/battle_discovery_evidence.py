"""第40輪原版七分支完整雙跑與 IDA 指令核對。沒有正式 remake 通過聲明。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct

from battle_discovery_plan import flags

DS = 0x54B30
DATA = Path("docs/internals/data")


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def w(m, p):
    return struct.unpack_from("<H", m, p)[0]


def far(m, p):
    return w(m, p) + w(m, p + 2) * 16


def write(p, value):
    assert p.parent.stat().st_uid == os.getuid()
    assert not p.exists() or p.stat().st_uid == os.getuid()
    p.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def events(text, address, end):
    block = re.search(r"\n" + address + r" 執行時[^\n]*\n((?:  [^\n]*\n|      [^\n]*\n)*)", text)
    return [int(x) for x in re.findall(r"#(\d+) ", block[1]) if int(x) < end] if block else []


plan = json.loads((DATA / "battle-discovery-plan.json").read_text())
assert sha(Path("workplace/oracle") / plan["parent"]) == plan["parent_sha256"]
cases = []
checked = 0
for case in plan["cases"]:
    roots = [Path("workplace/oracle") / f"battle79-discovery-r40-{case['name']}-{run}" for run in ["a", "b"]]
    frames = []
    names = [point["name"] + ext for point in case["points"] for ext in [".png", ".bin", "-mem.bin"]]
    names += ["screen.png", "screen.bin", "pal.bin"]
    for name in names:
        assert (roots[0] / name).read_bytes() == (roots[1] / name).read_bytes(), (case["name"], name)
    checked += len(names)
    for point in case["points"]:
        name = point["name"]
        m = (roots[0] / (name + "-mem.bin")).read_bytes()
        assert len(m) == 1048576
        assert len((roots[0] / (name + ".bin")).read_bytes()) == 640 * 480
        assert struct.unpack_from(">II", (roots[0] / (name + ".png")).read_bytes(), 16) == (640, 480)
        units = struct.unpack_from("<" + str(w(m, DS + 0x9896)) + "H", m, DS + 0x98BA)
        fleets = m[DS + 0x9898:DS + 0x9898 + w(m, DS + 0x9894)]
        ub, pb, pc = far(m, DS + 0xCFB4), far(m, DS + 0xCFF8), m[DS + 0xCFFC]
        obj = far(m, DS + 0xA78C)
        assert len(units) == 55 and len(fleets) == 7 and pc == 3
        if name.startswith("unit-"):
            assert obj and m[obj + 1] == 255 and m[obj + 0xA:obj + 0xD] == bytes([16, 16, 10])
        if name in ["entry", "header", "end"]:
            assert obj == 0
        frames.append({**point, "hashes": {ext: sha(roots[0] / (name + ext)) for ext in [".png", ".bin", "-mem.bin"]},
                       "camera_pixels": [w(m, DS + 0x9BF4), w(m, DS + 0x9BF6)],
                       "camera_grid": list(m[DS + 0x9BA3:DS + 0x9BA5]),
                       "seed": f"0x{struct.unpack_from('<I', m, 0x5E00C)[0]:08X}",
                       "controllers": list(m[DS + 0x8F21:DS + 0x8F23]), "seats": list(m[DS + 0x8F25:DS + 0x8F27]),
                       "sources_hex": m[DS + 0xA776:DS + 0xA77E].hex(),
                       "selection_object": {"linear_hex": hex(obj), "hex": m[obj:obj + 40].hex() if obj else ""},
                       "units": [{"index": i, "hex": m[ub + i*78:ub + (i+1)*78].hex()} for i in units],
                       "fleets": [{"index": i, "hex": m[DS + 0xC074 + i*122:DS + 0xC074 + (i+1)*122].hex()} for i in fleets],
                       "planets": [{"index": i, "linear_hex": hex(pb + i*82), "hex": m[pb + i*82:pb + (i+1)*82].hex()} for i in range(pc)]})
    assert len({f["seed"] for f in frames}) == 1 and frames[0]["seed"] == "0x12345678"
    before, after = frames[0], frames[-1]
    assert before["planets"] == after["planets"]
    fleet_changes = []
    for x, y in zip(before["fleets"], after["fleets"]):
        assert x["index"] == y["index"]
        a, b = bytes.fromhex(x["hex"]), bytes.fromhex(y["hex"])
        diff = [{"offset_hex": hex(i), "before": v, "after": b[i]} for i, v in enumerate(a) if v != b[i]]
        if diff:
            fleet_changes.append({"index": x["index"], "fields": diff})
    assert fleet_changes == ([{"index": 28, "fields": [{"offset_hex": "0x6a", "before": 9, "after": 6}]}]
                             if case["name"] == "detail3" else [])
    changed = []
    for x, y in zip(before["units"], after["units"]):
        assert x["index"] == y["index"]
        a, b = bytes.fromhex(x["hex"]), bytes.fromhex(y["hex"])
        diff = [{"offset_hex": hex(i), "before": v, "after": b[i]} for i, v in enumerate(a) if v != b[i]]
        if diff:
            changed.append({"pool": x["index"], "fields": diff})
    assert [c["pool"] for c in changed] == [109, 135, 137, 138, 139, 140, 141, 147]
    if case["name"] == "detail3":
        r = bytes.fromhex(next(u["hex"] for u in after["units"] if u["index"] == 109))
        assert r[0x17:0x19] == bytes([3, 2])
    if case["name"] == "camera":
        assert before["camera_pixels"] == [0, 0] and after["camera_pixels"] == [1088, 620]
        assert after["camera_grid"] == [68, 31]
    run_data = []
    for root in roots:
        meta = (root / "meta.txt").read_text()
        assert "dosgolem HEAD: f7523f9（未提交檔案 0 個）" in meta
        argv = (root / "flags.txt").read_text().splitlines()
        assert argv == flags(root.name)[0]
        report = (root / "report.txt").read_text()
        waits = events(report, "3765:006C", case["end"])
        highlights = events(report, "25DB:0E78", case["end"])
        assert len(waits) == (0 if case["name"] == "cpu" else 9)
        assert len(highlights) == (0 if case["name"] == "cpu" else 8)
        assert case["end"] in events(report, "181B:028C", case["end"] + 1)
        run_data.append({"path": str(root), "dosgolem_head": plan["dosgolem_head"], "clean": True,
                         "identity_method": "wrapper asserts full HEAD and clean before every run; original meta short HEAD also checked",
                         "command_argv": ["go", "run", "./cmd/probe", "-exe", "/orig/GIN3.COM", "-root", "/orig", "-empty-env"] + argv + ["-dump-screen", f"/src/workplace/oracle-tmp-{root.name}.bin", "-dump-screen-png", f"/src/workplace/oracle-tmp-{root.name}.png", "-dump-palette", f"/src/workplace/oracle-tmp-pal-{root.name}.bin"],
                         "go_image": "golang:1.24-bookworm", "go_image_id": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac",
                         "tick": 20000, "wait_entries": waits, "highlight_entries": highlights,
                         "hashes": {n: sha(root / n) for n in ["report.txt", "flags.txt", "meta.txt", "screen.bin", "screen.png", "pal.bin"]}})
    cases.append({"scenario": case["name"], "comparison": plan["comparison"], "inputs": case["inputs"], "scope": case["scope"], "frames": frames, "changes": changed, "fleet_changes": fleet_changes, "runs": run_data})

# 附加語意以版控索引按原始定位合併；原名、位址、bytes與運算元保留。
index_path = DATA / "battle-discovery-semantics.json"
index = json.loads(index_path.read_text())
ram_sources = {}
for source in index["ram_sources"]:
    p = Path(source["path"])
    assert sha(p) == source["sha256"]
    for segment in source["segments"]:
        ram_sources[segment] = p.read_bytes()
annotations = {(r["kind"], r["original_location"]): r for r in index["entries"]}
exports, functions = [], []
for name in index["exports"]:
    p = Path("workplace/ida") / name
    data = json.loads(p.read_text())
    assert data["tool"] == "IDA Pro 9.4"
    com = Path(data["input"].replace("/work/", "workplace/ida/"))
    assert sha(com) == data["input_sha256"]
    segment = data["functions"][0]["original_location"].split(":")[0]
    linear = int(segment, 16) * 16
    assert com.read_bytes() == bytes(256) + ram_sources[segment][linear:linear + 8192]
    exports.append({"path": str(p), "sha256": sha(p), "input_sha256": sha(com), "address_space": data["address_space"], "ida_base_hex": data["ida_base_hex"]})
    for func in data["functions"]:
        for kind, row in [("function", func)] + [("instruction", i) for i in func["instructions"]]:
            if kind == "instruction":
                off = int(row["original_location"].split(":")[1], 16)
                b = bytes.fromhex(row["original_bytes_hex"])
                assert com.read_bytes()[256 + off:256 + off + len(b)] == b
            note = annotations.get((kind, row["original_location"]))
            if note:
                if kind == "instruction":
                    assert note["original_bytes_hex"] == row["original_bytes_hex"]
                for key in ["semantic", "grade", "source"]:
                    row[key] = note[key]
                row["warning"] = "" if row["grade"] == "proven" else "⚠ 未證實語意"
        functions.append(func)

result = {"schema": "logh3.battle-discovery-evidence.v1", "program_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
          "bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
          "plan_sha256": sha(DATA / "battle-discovery-plan.json"), "semantic_index_sha256": sha(index_path),
          "dosgolem_head": plan["dosgolem_head"], "remake_parent_head": "bc3e2b1708fdd6c9c700a47d94464b032af59846",
          "ida_image_id": "sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780",
          "exports": exports, "functions": functions, "cases": cases, "identical_output_files_per_pair": checked,
          "limits": ["original controlled execution only; not remake same-state or normal player completion", "camera injection changes RAM but leaves cached old background until original redraw", "no runtime, seed or original program modification; complete injections and both commands recorded", "full production third-phase sequencing, persistent sprite cache and visibility-dependent per-unit camera loop not yet verified"]}
write(DATA / "battle-discovery-evidence.json", result)
print(f"七組原版雙跑：{checked}份輸出相同、{sum(len(c['frames']) for c in cases)}邊界、{sum(len(f['units']) for c in cases for f in c['frames'])}筆完整分隊。")
print("JSON SHA-256", sha(DATA / "battle-discovery-evidence.json"))
