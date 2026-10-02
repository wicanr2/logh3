"""第39輪完整第三階段原版雙跑與原始指令索引；不宣稱正式 remake 通過。"""
import hashlib
import json
import os
from pathlib import Path
import struct
import shutil
import sys
from battle_spotting_plan import HEAD, PARENTS, POINTS

DS = 0x54B30
PROGRAM = "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918"


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, obj):
    assert p.parent.stat().st_uid == os.getuid()
    assert not p.exists() or p.stat().st_uid == os.getuid()
    p.write_text(json.dumps(obj, ensure_ascii=False, indent=2) + "\n")


def word(m, p):
    return struct.unpack_from("<H", m, p)[0]


def far(m, p):
    return word(m, p + 2) * 16 + word(m, p)


if sys.argv[1:] == ["prepare-coordinate"]:
    control = json.loads(Path("docs/internals/data/battle-spotting-coordinates.json").read_text())
    p = Path("workplace/round39-coordinate-plan.json")
    script = Path("workplace/round39-coordinate-probe.go")
    assert not p.exists() and not script.exists()
    write(p, control["plan"])
    assert sha(p) == control["plan_sha256"]
    source = Path("tools/scripts/battle_step_probe.go")
    assert sha(source) == control["probe_source"]["sha256"]
    shutil.copyfile(source, script)
    print("六案完整注入計畫與固定診斷工具已重生。")
    sys.exit(0)


cases = []
for scenario, points in POINTS.items():
    a, b = [Path("workplace/oracle") / f"battle79-spot-r39-{scenario}-{run}" for run in ["a", "b"]]
    parent, digest = PARENTS[scenario]
    assert sha(Path("workplace/oracle") / parent) == digest
    for name in ["screen.bin", "screen.png", "pal.bin"]:
        assert (a / name).read_bytes() == (b / name).read_bytes(), name
    frames = []
    for name, step in points:
        paths = [a / (name + ext) for ext in [".bin", ".png", "-mem.bin"]]
        for p in paths:
            assert p.read_bytes() == (b / p.name).read_bytes(), p.name
        assert len(paths[0].read_bytes()) == 640 * 480
        assert struct.unpack_from(">II", paths[1].read_bytes(), 16) == (640, 480)
        m = paths[2].read_bytes()
        assert len(m) == 1048576
        ub = far(m, DS + 0xCFB4)
        fs = m[DS + 0x9898:DS + 0x9898 + word(m, DS + 0x9894)]
        us = struct.unpack_from("<" + str(word(m, DS + 0x9896)) + "H", m, DS + 0x98BA)
        pb, pc = far(m, DS + 0xCFF8), m[DS + 0xCFFC]
        frames.append({
            "name": name, "step": step, "hashes": {p.name: sha(p) for p in paths},
            "seed": f"0x{struct.unpack_from('<I', m, 0x5E00C)[0]:08X}",
            "camera_pixels": list(struct.unpack_from("<HH", m, 0x5E724)),
            "seats": list(m[DS + 0x8F25:DS + 0x8F27]),
            "controllers": list(m[DS + 0x8F21:DS + 0x8F23]),
            "sources_hex": m[DS + 0xA776:DS + 0xA77E].hex(),
            "fleets": [{"index": i, "hex": m[DS + 0xC074 + i*122:DS + 0xC074 + (i+1)*122].hex()} for i in fs],
            "units": [{"index": i, "hex": m[ub+i*78:ub+(i+1)*78].hex()} for i in us],
            "planets": [{"index": i, "linear_hex": hex(pb+i*82), "hex": m[pb+i*82:pb+(i+1)*82].hex()} for i in range(pc)],
        })
    changes = []
    for before, after in zip(frames[0]["units"], frames[-1]["units"]):
        assert before["index"] == after["index"]
        x, y = bytes.fromhex(before["hex"]), bytes.fromhex(after["hex"])
        diff = [{"offset_hex": hex(i), "before": v, "after": y[i]} for i, v in enumerate(x) if v != y[i]]
        if diff:
            changes.append({"pool": before["index"], "fields": diff})
    assert len(changes) == (0 if scenario == "s1" else 8)
    assert frames[0]["fleets"] == frames[-1]["fleets"]
    assert frames[0]["planets"] == frames[-1]["planets"]
    assert len({f["seed"] for f in frames}) == 1
    assert len({tuple(f["camera_pixels"]) for f in frames}) == 1
    runs = []
    for run in [a, b]:
        meta = (run / "meta.txt").read_text()
        assert "dosgolem HEAD: f7523f9（未提交檔案 0 個）" in meta
        assert PROGRAM in meta
        flags = (run / "flags.txt").read_text().splitlines()
        assert flags[0] == "-load-state" and flags[1] == "/out/oracle/" + parent
        runs.append({"path": str(run), "dosgolem_head": HEAD, "clean": True,
                     "identity_method": "wrapper asserts full HEAD and clean before run; meta short HEAD checked; no executor changes",
                     "go_image": "golang:1.24-bookworm", "tick": 20000,
                     "go_image_id": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac",
                     "command_argv": ["go", "run", "./cmd/probe", "-exe", "/orig/GIN3.COM", "-root", "/orig", "-empty-env"] + flags + [
                         "-dump-screen", f"/src/workplace/oracle-tmp-{run.name}.bin",
                         "-dump-screen-png", f"/src/workplace/oracle-tmp-{run.name}.png",
                         "-dump-palette", f"/src/workplace/oracle-tmp-pal-{run.name}.bin"],
                     "meta": meta, "hashes": {n: sha(run/n) for n in ["meta.txt", "report.txt", "flags.txt", "screen.bin", "screen.png", "pal.bin"]}})
    cases.append({"scenario": scenario, "parent": parent, "parent_sha256": digest,
                  "comparison": "original determinism only; remake production not verified",
                  "runs": runs, "frames": frames, "changes": changes})

meanings = {
    "181B:0263": "第三階段取DS:3D0D索敵文字",
    "181B:026F": "兩席索敵前呼叫3617:000A印階段",
    "2526:0104": "以座位讀DS:8F25陣營映射",
    "2526:011C": "觀測艦隊數來自DS:9894",
    "2526:013E": "只觀測艦隊+01陣營等於座位陣營者",
    "2526:0149": "每支所屬艦隊依清單呼叫0198",
    "2526:0175": "行星+11擁有者必須等於座位陣營",
    "2526:018C": "行星依82-byte記錄順序觀測",
    "2526:01B8": "旗艦池索引取艦隊+10",
    "2526:01D7": "偵察隊僅艦隊+3A不等於FF才觀測",
    "2526:01DE": "偵察隊池索引取艦隊+38",
    "2526:0205": "觀測分隊種類7選250，其餘200",
    "2526:021A": "觀測點Y讀SHIP+36現在位置",
    "2526:021E": "觀測點X讀SHIP+34現在位置",
    "2526:025C": "行星觀測點Y讀+48",
    "2526:0260": "行星觀測點X讀+46",
    "2526:0295": "候選來自17A0:00AC半徑查詢",
    "2526:02A2": "以2526:02D7篩選可提升情報的候選",
    "2526:02BA": "沒有提升候選時跳過發現報告",
    "2526:02C8": "提升候選交033B發現報告",
    "2526:0303": "Seen bit0已設時不重複提升",
    "2526:0309": "未設Seen先OR bit0，保留其餘bit",
    "2526:0315": "只有Detail恰等3才不提升",
    "2526:031A": "Detail以byte加1",
    "2526:032C": "此次升到3才將Hold設2",
    "2526:0353": "艦隊報告以艦隊+1判陣營",
    "2526:036B": "行星報告以行星+11判陣營",
    "2526:0380": "3982:022D判觀測來源是否電腦",
    "2526:038B": "電腦觀測來源跳過所有UI與等待，情報已更新",
    "2526:0395": "人類來源以3826:0298顯示標頭，參數97",
    "2526:03A6": "標頭後以1588:01A3定位觀測者，旗標1",
    "2526:03C0": "發現列表以0520排序，元素2 bytes",
    "2526:03F8": "顯示物件+1E為0才可直接展示分隊",
    "2526:040D": "其他情形以分隊現在位置定位，旗標0",
    "2526:041C": "逐筆25DB:0E78展示候選資訊",
    "2526:043D": "逐筆25DB:03E9以force0重算圖示",
    "2526:0444": "逐筆25DB:0759重畫",
    "2526:0449": "逐筆展示固定等待6 ticks",
    "2526:0454": "逐筆25DB:0F93結束資訊展示",
    "2526:055E": "排序第一鍵SHIP+01艦隊索引",
    "2526:056A": "艦隊索引unsigned小者在前",
    "2526:0595": "艦隊相同時第二鍵SHIP+03分隊序號",
    "2526:05A2": "分隊序號unsigned小者在前",
    "17A0:00FC": "候選陣營直接讀SHIP+00",
    "17A0:0106": "候選只以SHIP+1E bit7排除，不直接篩Kind",
    "17A0:010F": "候選Y讀SHIP+36現在位置",
    "17A0:0115": "候選X讀SHIP+34現在位置",
    "17A0:011D": "距離由3310:0006計算",
    "17A0:0133": "距離大於半徑才排除，邊界包含",
}
source = Path("workplace/oracle/battle79-second-r36-a/entry-mem.bin")
memory = source.read_bytes()
assert sha(source) == "9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec"
entries, exports = [], []
for stem in ["181B", "2526", "17A0", "report"]:
    p = Path("workplace/ida") / f"re79-spot-r39-{stem}.json"
    export = json.loads(p.read_text())
    assert export["tool"] == "IDA Pro 9.4"
    assert sha(p.with_suffix(".com")) == export["input_sha256"]
    exports.append({"path": str(p), "sha256": sha(p), "input_sha256": export["input_sha256"]})
    for fn in export["functions"]:
        for row in fn["instructions"]:
            loc = row["original_location"]
            if loc not in meanings:
                continue
            raw = bytes.fromhex(row["original_bytes_hex"])
            linear = int(row["runtime_linear_hex"], 16)
            assert memory[linear:linear+len(raw)] == raw, loc
            row.update(semantic=meanings[loc], grade="proven", warning="", source=["docs/re/379-battle-spotting-phase.md", sha(p)])
            entries.append(dict(row))
    export["semantic_index"] = "docs/internals/data/battle-spotting-phase.json"
    write(p.with_name(p.stem + "-reviewed.json"), export)
assert len(entries) == len(meanings)
strings = []
for off in [0x3D0D, 0x4A00, 0x4A0B, 0x4A16]:
    raw = memory[DS+off:].split(bytes(1), 1)[0]
    strings.append({"original_location": f"DS:{off:04X}", "hex": raw.hex(), "text_big5": raw.decode("big5"), "grade": "proven"})
write(Path("docs/internals/data/battle-spotting-phase.json"), {
    "schema": "logh3.battle-spotting-phase.v1", "program_sha256": PROGRAM,
    "bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
    "dosgolem_head": HEAD, "remake_parent_head": "c89c22e4ebf0b1f7b38e180e577e6c26041c6876",
    "address_space": "CS:IP, dosgolem linear, IDA COM offset0=0x10200 separately labelled; DS linear=0x54B30",
    "tool": "IDA Pro 9.4", "ida_image_id": "sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780",
    "ram_sha256": sha(source), "entries": entries, "exports": exports,
    "strings": strings, "canvas": [640, 480], "cases": cases,
    "limits": ["原版雙跑完整第三階段；正式remake尚未對拍", "Detail3資訊、鏡頭移動、滑鼠提前結束及雙人分支仍待取證"]})
print(f"18邊界54檔及6份終態輸出雙跑相同；{len(entries)}筆指令；190筆分隊初態、8筆變更。")

# 座標分離案例用既有battle_step_probe通用常式診斷，不當作正常GUI證據。
a = Path("workplace/logh3-round39-coordinate-a.json")
b = Path("workplace/logh3-round39-coordinate-b.json")
assert a.read_bytes() == b.read_bytes()
assert sha(a) == "24fec495ea3ab7884e61564d81b00093c1c28e706583dc4c62c07fdb3b3d1ae4"
control = json.loads(a.read_text())
assert control["dosgolem_head"] == HEAD and control["program_sha256"] == PROGRAM
assert sha(Path("workplace/round39-coordinate-plan.json")) == control["plan_sha256"]
assert len(control["rows"]) == 6
for row in control["rows"]:
    assert row["before"]["observer"] == row["after"]["observer"]
    assert row["seed_after_hex"] == "78563412"
    before, after = [bytes.fromhex(row[k]["candidate"]) for k in ["before", "after"]]
    assert all(before[i] == after[i] for i in range(78) if i not in [22, 23])
control["grade"] = "proven"
control["runs"] = [{"path": str(p), "sha256": sha(p)} for p in [a, b]]
control["command"] = "DOSGOLEM_ORIG=<project>/org_game/extracted DOSGOLEM_EXTRA_MOUNT=<project>/workplace:/out:ro <dosgolem>/tools/go.sh run /out/round39-coordinate-probe.go -plan /out/round39-coordinate-plan.json -output /src/workplace/logh3-round39-coordinate-<a|b>.json"
control["probe_source"] = {"path": "tools/scripts/battle_step_probe.go", "sha256": sha(Path("tools/scripts/battle_step_probe.go"))}
write(Path("docs/internals/data/battle-spotting-coordinates.json"), control)
print("6案座標分離／半徑邊界完整原版雙跑相同。")
