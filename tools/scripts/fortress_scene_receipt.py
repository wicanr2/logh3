"""第133輪：核對原版三案雙跑、IDA relocation 與正常 GUI 觸發。

tools/py.sh tools/scripts/fortress_scene_receipt.py --write
不加 --write 只核對已凍結的來源與結果；原版 RAM 不加入 Git。
"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DS = 0x54B30
HEAD = "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
OUT = ROOT / "docs/internals/data/fortress-scene.json"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load(path):
    return json.loads(path.read_text())


def u16(b, a):
    return struct.unpack_from("<H", b, a)[0]


def u32(b, a):
    return struct.unpack_from("<I", b, a)[0]


def base(b):
    return u16(b, DS + 0xCFB4) + 16 * u16(b, DS + 0xCFB6)


def raw(kind, name):
    return (ROOT / f"workplace/oracle/fortress48-r133-r3-{kind}-a" / name).read_bytes()


def collect():
    out = {"schema": "logh3.fortress-scene.v1", "date": "2026-10-14",
           "remake_base_commit": "b36b950f", "remake_dirty": True,
           "classification": "original injected diagnostic and independent normal remake GUI; no cross-engine same-state claim",
           "original_bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
           "original_gin3_sha256": sha(Path("/orig/GIN3.COM")) if Path("/orig/GIN3.COM").exists() else sha(ROOT / "org_game/extracted/GIN3.COM"),
           "native": {}, "sources": {}}
    for kind, expected in [("hit", (225, 138, 568)), ("empty", (8, 5, 36)), ("off", (128, 52, 224))]:
        dirs = [ROOT / f"workplace/oracle/fortress48-r133-r3-{kind}-{side}" for side in ("a", "b")]
        reports = [load(d / "report.json") for d in dirs]
        for r in reports:
            assert r["dosgolem_head"] == HEAD and not r["dosgolem_dirty"]
            assert r["seed_once"] and r["end_original_location"] == "181B:03D0"
            assert r["parent_sha256"] == "c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1"
            assert (len(r["events"]), len(r["frames"])) == expected[:2]
            assert r["uid"] == 1000 and r["irq0_every"] == 20000 and r["irq0_pinned"] and not r["cycle_clock"]
        names = sorted(p.name for p in dirs[0].iterdir() if p.suffix in (".bin", ".png") or p.name.endswith("-palette.json"))
        assert len(names) == expected[2]
        assert names == sorted(p.name for p in dirs[1].iterdir() if p.suffix in (".bin", ".png") or p.name.endswith("-palette.json"))
        for name in names:
            assert (dirs[0] / name).read_bytes() == (dirs[1] / name).read_bytes(), (kind, name)
        out["native"][kind] = {"directories": [str(p.relative_to(ROOT)) for p in dirs],
            "reports_sha256": [sha(p / "report.json") for p in dirs],
            "report_a": reports[0], "matched_files": {n: sha(dirs[0] / n) for n in names}}
    before, after = [raw("hit", n) for n in ("gun-before-mem.bin", "gun-return-181b-03bf-01-mem.bin")]
    assert len(before) == len(after) == 1048576
    hits = []
    for i in range(336):
        a, b = base(before) + i * 78, base(after) + i * 78
        lost = u16(before, a + 8) - u16(after, b + 8)
        if lost:
            assert lost > 0
            hits.append({"pool": i, "fleet": before[a + 1], "squadron": before[a + 3],
                         "before": u16(before, a + 8), "after": u16(after, b + 8), "lost": lost})
    assert len(hits) == 10 and sum(h["lost"] for h in hits) == 10637
    off = raw("off", "gun-return-181b-03bf-01-mem.bin")
    assert after[base(after):base(after) + 336 * 78] == off[base(off):base(off) + 336 * 78]
    assert after[DS + 0x8FE4:DS + 0x8FF4] == off[DS + 0x8FE4:DS + 0x8FF4]
    assert u32(after, DS + 0x94DC) == u32(off, DS + 0x94DC)
    empty0, empty1 = [raw("empty", n) for n in ("gun-before-mem.bin", "gun-return-181b-03bf-01-mem.bin")]
    assert all(u16(empty0, base(empty0) + i * 78 + 8) == u16(empty1, base(empty1) + i * 78 + 8) for i in range(336))
    assert u32(empty0, DS + 0x94DC) == u32(empty1, DS + 0x94DC)
    assert not any(e["original_location"] == "1DD6:0152" for e in out["native"]["empty"]["report_a"]["events"])
    events = out["native"]["hit"]["report_a"]["events"]
    calls = [e for e in events if "comani_record_hex" in e]
    cues = [e["stack_words"][2] for e in events if e["original_location"] == "3B74:031A"]
    assert len(calls) == 31 and cues == [27, 7, 3, 6, 3, 0, 4, 2, 3, 1, 7]
    attacker = [e for e in calls if e["comani_record_hex"] == "000a0001ffff1b"]
    assert len(attacker) == 1 and attacker[0]["stack_words"][4:6] == [0, 1]
    out["gun"] = {"hits": hits, "total_lost": 10637, "seed_before": u32(before, DS + 0x94DC),
                  "seed_after": u32(after, DS + 0x94DC), "animation_off_preserves_rules": True,
                  "empty_preserves_ships_and_seed": True, "comani_calls": len(calls), "cue_calls": cues}
    # IDA loader 的 0x1000 與 dosgolem 的 0x349 是不同 MZ 載入段。
    exe = ROOT / "workplace/ida/GIN3PS_UNPACKED.EXE"
    b = exe.read_bytes()
    ida_path = ROOT / "workplace/ida/fortress133/fortress-scene-ida.json"
    ida = load(ida_path)
    assert sha(exe) == ida["input_sha256"]
    assert ida["native_input_sha256"] == hashlib.sha256(before).hexdigest()
    relocs = set()
    for i in range(u16(b, 6)):
        a = u16(b, 24) + 4 * i
        relocs.add(u16(b, a) + 16 * u16(b, a + 2))
    normalized, relocated = 0, 0
    for f in ida["functions"]:
        for ins in f["instructions"]:
            image_offset = int(ins["ida_ea"], 16) - 0x10000
            assert int(ins["runtime_linear"], 16) == image_offset + 0x3490
            bs = bytearray.fromhex(ins["original_bytes_hex"])
            for off in relocs:
                n = off - image_offset
                if 0 <= n < len(bs):
                    assert n + 2 <= len(bs)
                    struct.pack_into("<H", bs, n, (u16(bs, n) - 0x1000 + 0x349) & 65535)
                    relocated += 1
            native = bytes.fromhex(ins["native_bytes_hex"])
            assert bs == native == before[image_offset + 0x3490:image_offset + 0x3490 + len(bs)]
            normalized += 1
    assert (normalized, relocated) == (778, 63)
    out["ida"] = {"export_sha256": sha(ida_path), "image": "ida-pro-9.4-idapython:py312-v1",
        "image_id": "sha256:4ac62de83339c215bab10e455cee3a22d9c6efed9fd0d8ed0f068327b83a06ab",
        "normalization": "IDA EA=image_offset+0x10000; runtime linear=image_offset+0x3490; each MZ relocation word=(IDA word-0x1000+0x349)&0xffff",
        "matched_instructions": normalized, "relocated_words": relocated, "export": ida}
    gui = ROOT / "workplace/round133/gui-b"
    g = load(gui / "RESULT.json")
    assert g["passed"] and g["normal_entry"] and g["ai"] == -1
    assert not any(g[k] for k in ("state_injection", "seed_injection", "script_flag", "passive_override"))
    assert g["region"] == 13 and g["planet"]["index"] == 33 and g["planet"]["gun_state"] == 2
    assert len(g["hits"]) == 2 and sum(h["lost"] for h in g["hits"]) == 4932
    assert g["binary_sha256"] == sha(ROOT / "workplace/round133/gin3")
    out["gui"] = {"image": "logh3-hd-preview:20261011-r1", "directory": str(gui.relative_to(ROOT)),
        "result_sha256": sha(gui / "RESULT.json"), "trace_sha256": sha(gui / "FRAMES.jsonl"),
        "result": g, "screenshots": {p.name: sha(p) for p in sorted(gui.glob("*.png"))}}
    source_root = Path("/ref") if Path("/ref/internal/game/herocommand.go").exists() else ROOT
    names = ["tools/fortress_scene_oracle.sh", "tools/fortress_scene_native_run.sh",
             "tools/scripts/fortress_scene_native.go", "tools/scripts/fortress_scene_ida.py",
             "tools/scripts/fortress_player_gui.py", "tools/scripts/fortress_scene_receipt.py",
             "internal/game/herocommand.go", "internal/rules/fortress_scene_oracle_test.go"]
    for name in names:
        out["sources"][name] = sha(source_root / name)
    return out


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--write", action="store_true")
    args = p.parse_args()
    result = collect()
    if args.write:
        assert not OUT.exists(), "拒絕覆寫已凍結來源索引"
        with OUT.open("x") as f:
            json.dump(result, f, ensure_ascii=False, indent=2)
            f.write("\n")
    else:
        assert result == load(OUT), "來源或收據與凍結索引不符"
    print("PASS: 原版828檔雙跑一致；778指令／63 relocation一致；10支10637艘；正常GUI 2支4932艘")
    print("來源索引 SHA-256", sha(OUT))


if __name__ == "__main__":
    main()
