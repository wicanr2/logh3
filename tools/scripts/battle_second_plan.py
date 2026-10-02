"""#79 S1第二戰場固定移動邊界；契約見第36輪及驗證045。"""
import hashlib
import json
from pathlib import Path
import struct
import sys

PARENT = Path("workplace/oracle/battle79-round36-s1-menu-explore/menu.state")
PARENT_HASH = "ee5a09cf20429062cc1d07351268937f30280a81f5d09e6ee31546e920ef67c2"
HEAD = "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
COUNTS = [1239412351, 1240687009, 1241758089, 1242775932, 1243721921,
          1244684619, 1245667923, 1246680676, 1247708880, 1248717209,
          1249722632, 1250731511, 1251733933, 1252737547, 1253713185,
          1254665384, 1255593415, 1256513911, 1257428005, 1258340500,
          1259230813, 1259926902, 1260634344, 1261333878, 1261830988,
          1262222722, 1262519246, 1262816502, 1263113076, 1263401119,
          1263689390, 1263985298, 1264178495, 1264370613, 1264562408,
          1264754377, 1264946766, 1265138937, 1265315893, 1265493493,
          1265670654]
POINTS = [("entry", 1239379060)] + [(f"loop-{i:02d}", n) for i, n in enumerate(COUNTS)] + [("end", 1265700574)]


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


assert sha(PARENT) == PARENT_HASH
if sys.argv[1] == "plan":
    name = sys.argv[2]
    assert name in ("battle79-second-r36-a", "battle79-second-r36-b")
    prefix = f"/src/workplace/logh3-{name}"
    flags = ["-dump-at", ";".join(f"{n}:{prefix}/{k}.png" for k, n in POINTS),
             "-dump-mem-at", ";".join(f"{n}:lin:0:1048576:{prefix}/{k}-mem.bin" for k, n in POINTS)]
    out = Path("workplace") / (name + "-flags.txt")
    assert not out.exists()
    out.write_text("\n".join(flags) + "\n")
elif sys.argv[1] == "export":
    a, b = [Path("workplace/oracle") / ("battle79-second-r36-" + x) for x in "ab"]
    frames = []
    for name, step in POINTS:
        for ext in (".bin", ".png", "-mem.bin"):
            assert (a / (name + ext)).read_bytes() == (b / (name + ext)).read_bytes(), name + ext
        m = (a / (name + "-mem.bin")).read_bytes()
        assert len(m) == 1048576
        u = lambda addr: struct.unpack_from("<H", m, addr)[0]
        assert u(0x5e3c4) == 18 and u(0x5e3c6) == 135
        assert u(0x5da48) == 93 and u(0x5da4a) == 0
        assert m[0x5e00c:0x5e010].hex() == "87cdc6ec"
        assert (u(0x5e724), u(0x5e726)) == (928, 20)
        fleets = list(m[0x5e3c8:0x5e3c8+18])
        units = [u(0x5e3ea+2*i) for i in range(135)]
        assert len(set(fleets)) == 18 and len(set(units)) == 135
        frames.append({"name": name, "step": step,
                       "hashes": {ext: sha(a / (name + ext)) for ext in (".bin", ".png", "-mem.bin")},
                       "fleets": [{"index": i, "hex": m[0x60ba4+i*122:0x60ba4+(i+1)*122].hex()} for i in fleets],
                       "units": [{"index": i, "hex": m[0x668b2+i*78:0x668b2+(i+1)*78].hex()} for i in units]})
    assert (a / "pal.bin").read_bytes() == (b / "pal.bin").read_bytes()
    data = {"schema": "logh3.battle-phase-frames.v1", "grade": "same-state",
            "scope": "S1 region8 / Turn1 / 18 fleets / 135 units / 41 movement loops; remake entry injection only; normal player path separate",
            "program_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
            "dosgolem_head": HEAD, "snapshot_sha256": PARENT_HASH,
            "input_seed": "0x12345678", "movement_seed": "0xECC6CD87",
            "seed_method": "lin:5E00C@1220999999=78 56 34 12, before first Execute; preparation consumes RNG; movement entry seed derived and fixed in remake",
            "camera": [58, 1], "canvas": [640, 480],
            "runs": [{"path": str(p), "meta": (p / "meta.txt").read_text(),
                      "hashes": {n: sha(p / n) for n in ("meta.txt", "flags.txt", "report.txt", "pal.bin")}} for p in (a, b)],
            "frames": frames}
    out = Path("docs/internals/data/battle-second-frames.json")
    out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
    print(f"43邊界／129檔雙跑相同；18艦隊／135分隊，seed ECC6CD87，鏡頭58,1。JSON {sha(out)}")
else:
    raise ValueError("plan <name> 或 export")
