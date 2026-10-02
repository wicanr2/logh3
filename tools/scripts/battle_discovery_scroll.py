"""第43輪固定索敵來源與逐格置鏡。入口見對應goal；禁止覆寫收據。"""
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import sys

DATA = Path("docs/internals/data")
PLAN = DATA / "battle-discovery-scroll-plan.json"
DS = 0x54B30


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, data):
    assert p.parent.stat().st_uid == os.getuid()
    assert not p.exists() or p.stat().st_uid == os.getuid()
    p.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")


def flags(name):
    m = re.fullmatch(r"battle79-scroll-r43-(fleet13|fleet3|fleet3-cache13|scout13|near-source|offscreen)-(explore(?:-v[23])?|[ab](?:-replay[1-9][0-9]*)?)", name)
    assert m
    plan = json.loads(PLAN.read_text())
    if m[2].startswith(("a", "b")):
        assert plan.get("status") == "FROZEN", "正式收據只能使用已凍結計畫"
    assert sha(Path("workplace/oracle") / plan["parent"]) == plan["parent_sha256"]
    case = next(c for c in plan["cases"] if c["name"] == m[1])
    prefix = "/src/workplace/logh3-" + name
    result = ["-load-state", "/out/oracle/" + plan["parent"], "-steps", str(case["end"] + 1),
              "-tick", "20000", "-regs-from", "1226630000", "-regs-max", "256", "-regs-at", plan["regs_at"]]
    result += case["inputs"]
    result += ["-dump-at", ";".join(f"{p['step']}:{prefix}/{p['name']}.png" for p in case["points"]),
               "-dump-mem-at", ";".join(f"{p['step']}:lin:0:1048576:{prefix}/{p['name']}-mem.bin" for p in case["points"])]
    return result, len(case["points"]) * 3


def events(text, address, end):
    block = re.search(r"\n" + address + r" 執行時[^\n]*\n((?:  [^\n]*\n|      [^\n]*\n)*)", text)
    return [int(x) for x in re.findall(r"#(\d+) ", block[1]) if int(x) < end] if block else []


def freeze():
    plan = json.loads(PLAN.read_text())
    for c in plan["cases"]:
        root = Path("workplace/explore") / ("battle79-scroll-r43-" + c["name"] + "-" + c.get("explore_suffix", "explore"))
        text = (root / "report.txt").read_text()
        end = events(text, "181B:028C", 1300000001)[0]
        ev = lambda addr: events(text, addr, end)
        waits, highlights = ev("3765:006C"), ev("25DB:0E78")
        assert len(highlights) == (8 if c["name"] == "near-source" else 1)
        assert len(waits) == len(highlights) + 1
        points = [{"name": "entry", "step": 1226638239},
                  {"name": "label", "step": ev("2526:00F8")[0]},
                  {"name": "discovered", "step": ev("2526:033B")[0] + 2},
                  {"name": "header", "step": waits[0] + 2}]
        points += [{"name": "scroll-" + str(i), "step": s} for i, s in enumerate(ev("1588:0437"))]
        points += [{"name": "highlight-start", "step": highlights[0]}]
        points += [{"name": "unit-" + str(i), "step": s + 2} for i, s in enumerate(waits[1:])]
        points += [{"name": "end", "step": end}]
        c["end"], c["points"] = end, sorted(points, key=lambda p: p["step"])
        c["exploration"] = {"report": str(root / "report.txt"), "sha256": sha(root / "report.txt")}
        c["observed_events"] = {a: ev(a) for a in ["2526:046A", "2526:04D5", "1588:01A3", "1588:0437", "3765:006C", "25DB:0E78"]}
    plan["status"] = "FROZEN"
    write(PLAN, plan)


def collect(run_names=("a", "b")):
    assert len(run_names) == 2 and all(re.fullmatch(r"[ab](?:-replay[1-9][0-9]*)?", r) for r in run_names)
    assert run_names[0].startswith("a") and run_names[1].startswith("b")
    plan = json.loads(PLAN.read_text())
    cases, count = [], 0
    word = lambda m, p: struct.unpack_from("<H", m, p)[0]
    far = lambda m, p: word(m, p) + 16 * word(m, p + 2)
    for c in plan["cases"]:
        roots = [Path("workplace/oracle") / ("battle79-scroll-r43-" + c["name"] + "-" + r) for r in run_names]
        names = [p["name"] + ext for p in c["points"] for ext in [".bin", ".png", "-mem.bin"]]
        names += ["screen.bin", "screen.png", "pal.bin"]
        for n in names:
            assert (roots[0] / n).read_bytes() == (roots[1] / n).read_bytes(), (c["name"], n)
        count += len(names)
        entry = (roots[0] / "entry-mem.bin").read_bytes()
        for i in c["injections"]:
            address, value = int(i["linear_hex"], 16), bytes.fromhex(i["hex"])
            assert entry[address:address + len(value)] == value, (c["name"], i)
        frames = []
        for point in c["points"]:
            name = point["name"]
            m = (roots[0] / (name + "-mem.bin")).read_bytes()
            assert len(m) == 1048576 and len((roots[0] / (name + ".bin")).read_bytes()) == 307200
            assert struct.unpack_from(">II", (roots[0] / (name + ".png")).read_bytes(), 16) == (640, 480)
            units = struct.unpack_from("<" + str(word(m, DS + 0x9896)) + "H", m, DS + 0x98BA)
            fleets = m[DS + 0x9898:DS + 0x9898 + word(m, DS + 0x9894)]
            ub, pb, pc = far(m, DS + 0xCFB4), far(m, DS + 0xCFF8), m[DS + 0xCFFC]
            assert len(units) == 55 and len(fleets) == 7 and pc == 3
            obj = far(m, DS + 0xA78C)
            if name.startswith("unit-"):
                assert obj and m[obj + 1] == 255 and m[obj + 10:obj + 13] == bytes([16, 16, 10])
            else:
                assert obj == 0
            frames.append({**point, "hashes": {ext: sha(roots[0] / (name + ext)) for ext in [".bin", ".png", "-mem.bin"]},
                           "camera_grid": list(m[DS + 0x9BA3:DS + 0x9BA5]),
                           "camera_pixels": [word(m, DS + 0x9BF4), word(m, DS + 0x9BF6)],
                           "seed": f"0x{struct.unpack_from('<I',m,0x5E00C)[0]:08X}",
                           "units": [{"index": i, "hex": m[ub + i * 78:ub + (i + 1) * 78].hex()} for i in units],
                           "fleets": [{"index": i, "hex": m[0x60BA4 + i * 122:0x60BA4 + (i + 1) * 122].hex()} for i in fleets],
                           "planets": [{"index": i, "linear_hex": hex(pb + i * 82), "hex": m[pb + i * 82:pb + (i + 1) * 82].hex()} for i in range(pc)],
                           "candidate109_object": m[0x71406:0x71406 + 40].hex()})
        assert {f["seed"] for f in frames} == {"0x12345678"}
        assert frames[0]["planets"] == frames[-1]["planets"]
        runs = []
        for root in roots:
            argv = (root / "flags.txt").read_text().splitlines()
            assert argv == flags(root.name)[0]
            meta = (root / "meta.txt").read_text()
            assert "dosgolem HEAD: f7523f9（未提交檔案 0 個）" in meta
            assert "GIN3.COM SHA-256: 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" in meta
            report = (root / "report.txt").read_text()
            assert events(report, "181B:028C", c["end"] + 1) == [c["end"]]
            assert all(events(report, a, c["end"]) == v for a, v in c["observed_events"].items())
            source = re.search(r"\n2526:01FE 執行時[^\n]*\n((?:  [^\n]*\n|      [^\n]*\n)*)", report)
            runs.append({"path": str(root), "dosgolem_head": plan["dosgolem_head"], "clean": True,
                         "command_argv": ["go", "run", "./cmd/probe", "-exe", "/orig/GIN3.COM", "-root", "/orig", "-empty-env"] + argv +
                             ["-dump-screen", "/src/workplace/oracle-tmp-" + root.name + ".bin", "-dump-screen-png", "/src/workplace/oracle-tmp-" + root.name + ".png", "-dump-palette", "/src/workplace/oracle-tmp-pal-" + root.name + ".bin"],
                         "source_observer_trace": source[0] if source else "",
                         "hashes": {n: sha(root / n) for n in ["report.txt", "flags.txt", "meta.txt", "screen.bin", "screen.png", "pal.bin"]}})
        cases.append({"scenario": c["name"], "comparison": plan["comparison"], "injections": c["injections"], "frames": frames, "runs": runs})
    out = {"schema": "logh3.battle-discovery-scroll.v1", "dosgolem_head": plan["dosgolem_head"], "dosgolem_worktree": "clean",
           "program_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
           "bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
           "go_image_id": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac",
           "remake_parent_head": "649bef601b3ed46f976e9cfc5627a729ca41310b",
           "plan_sha256": sha(PLAN), "parent": plan["parent"], "parent_sha256": plan["parent_sha256"],
           "address_space": "dosgolem runtime real-mode linear; trace original CS:IP; no IDA rebasing",
           "identical_output_files_per_pair": count, "cases": cases,
           "limits": ["unique controlled original entry; not normal original player route", "no CPU or wall-clock equivalence", "no RNG retry or original binary/runner edits"]}
    source = DATA / "battle-discovery-evidence.json"
    static = json.loads(source.read_text())
    semantics = {"1588:0432": "單步鏡頭呼叫25DB:09C5；返回後才取完整重畫停點",
                 "1588:0437": "單步鏡頭完整重畫後的觀測邊界；不是新增等待或硬體時鐘契約"}
    reviewed = []
    for func in static["functions"]:
        for row in func["instructions"]:
            if row["original_location"] not in semantics:
                continue
            b = bytes.fromhex(row["original_bytes_hex"])
            a = int(row["runtime_linear_hex"], 16)
            assert entry[a:a + len(b)] == b
            reviewed.append({**row, "semantic": semantics[row["original_location"]], "grade": "proven", "warning": "",
                             "source": {"static_export": str(source), "sha256": sha(source),
                                        "dynamic": "five fixed cases; full redraw frame hashes and both command traces in this receipt"}})
    assert len(reviewed) == 2
    out["reviewed_semantics"] = reviewed
    out["resolution_backlinks"] = [
        {"key": "PC DOS/GIN3.COM/770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918/2526:00F8",
         "evidence": "docs/re/381-battle-spotting-display.md", "older": "docs/spec/018-battle-resolution.md", "required_marker": "索敵目前契約見087"},
        {"key": "PC DOS/GIN3.COM/770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918/2526:00F8",
         "evidence": "docs/re/381-battle-spotting-display.md", "older": "docs/spec/062-battle-turn-eight-steps.md", "required_marker": "第三階段已在戰鬥前，間隔為發現標頭與逐筆等待"},
        {"key": "PC DOS/GIN3.COM/770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918/1588:0437",
         "evidence": "docs/re/381-battle-spotting-display.md", "older": "docs/spec/087-battle-spotting-phase.md", "required_marker": "第43輪限定驗收"}]
    write(DATA / "battle-discovery-scroll.json", out)
    print("雙跑", count, "份相同；邊界", sum(len(c["frames"]) for c in cases), "SHA-256", sha(DATA / "battle-discovery-scroll.json"))


def render():
    from pngout import game_palette, write_png
    p = DATA / "battle-discovery-scroll.json"
    data = json.loads(p.read_text())
    for row in data["resolution_backlinks"]:
        assert row["key"].split("/")[-1] in Path(row["evidence"]).read_text()
        assert row["required_marker"] in Path(row["older"]).read_text()
    geometry = json.loads((DATA / "screen-battle.json").read_text())
    f, m, s = geometry["field"], geometry["minimap"], geometry["status"]
    regions = {"full": [0, 0, 640, 480],
               "field": [f["x"], f["y"], f["cols"] * f["tile_w"], f["rows"] * f["tile_h"]],
               "minimap": [m["x"], m["y"], m["w"], m["h"] + 1],
               "status": [s["x"], s["y"], s["w"], s["h"]]}
    results, excluded = [], []
    for c in data["cases"]:
        base, other = [Path(r["path"]) for r in c["runs"]]
        for frame in c["frames"]:
            name = frame["name"]
            if name == "entry":
                excluded.append({"scenario": c["scenario"], "frame": name, "reason": "寫索敵文字之前；唯一輸入身份已核"})
                continue
            old = (base / (name + ".bin")).read_bytes()
            for ext, digest in frame["hashes"].items():
                original = base / (name + ext)
                assert sha(original) == digest and original.read_bytes() == (other / original.name).read_bytes()
            prefix = Path("workplace/round43-spot-" + c["scenario"] + "-" + name)
            new, dac = prefix.with_suffix(".bin").read_bytes(), prefix.with_suffix(".pal").read_bytes()
            palette = bytes((v << 2) | (v >> 4) for v in dac)
            assert len(old) == len(new) == 307200 and len(dac) == 48
            assert palette == bytes(game_palette(str(base / (name + ".png"))))
            write_png(prefix.with_suffix(".png"), 640, 480, new, palette)
            delta = prefix.with_name(prefix.name + "-diff").with_suffix(".png")
            write_png(delta, 640, 480, bytes(int(a != b) for a, b in zip(old, new)), bytes([0, 0, 0, 255, 0, 0]))
            stats = {}
            for label, (x, y, w, h) in regions.items():
                points = [j * 640 + i for j in range(y, y + h) for i in range(x, x + w)]
                diff = [i for i in points if old[i] != new[i]]
                stats[label] = {"rect": [x, y, w, h], "compared": len(points), "equal": len(points) - len(diff), "different": len(diff),
                                "max_rgb_channel_difference": max((abs(palette[3 * old[i] + k] - palette[3 * new[i] + k]) for i in diff for k in range(3)), default=0),
                                "bbox_inclusive": [min(i % 640 for i in diff), min(i // 640 for i in diff), max(i % 640 for i in diff), max(i // 640 for i in diff)] if diff else None}
            assert stats["full"]["different"] == 0, (c["scenario"], name)
            before = Path("workplace/round43-before-" + c["scenario"] + "-" + name)
            before_pixels = before.with_suffix(".bin").read_bytes()
            before_dac = before.with_suffix(".pal").read_bytes()
            assert before_dac == dac
            write_png(before.with_suffix(".png"), 640, 480, before_pixels, palette)
            before_delta = before.with_name(before.name + "-diff").with_suffix(".png")
            write_png(before_delta, 640, 480, bytes(int(a != b) for a, b in zip(old, before_pixels)), bytes([0, 0, 0, 255, 0, 0]))
            files = [base / (name + ext) for ext in [".bin", ".png", "-mem.bin"]]
            files += [prefix.with_suffix(ext) for ext in [".bin", ".pal", ".png"]] + [delta]
            files += [before.with_suffix(ext) for ext in [".bin", ".pal", ".png"]] + [before_delta]
            results.append({"scenario": c["scenario"], "frame": name, "camera": frame["camera_grid"],
                            "units": len(frame["units"]), "fleets": len(frame["fleets"]), "stats": stats,
                            "effective_palette_equal": True,
                            "before_difference_count": sum(a != b for a, b in zip(old, before_pixels)),
                            "files": {str(f): sha(f) for f in files}})
    normal = []
    for label in ["spotting-header", "spotting-completed"]:
        prefix = Path("workplace/round43-normal-" + label)
        pixels, dac = prefix.with_suffix(".bin").read_bytes(), prefix.with_suffix(".pal").read_bytes()
        assert len(pixels) == 307200 and len(dac) == 48
        write_png(prefix.with_suffix(".png"), 640, 480, pixels, bytes((v << 2) | (v >> 4) for v in dac))
        normal.append({"frame": label, "files": {str(prefix.with_suffix(ext)): sha(prefix.with_suffix(ext)) for ext in [".bin", ".pal", ".png"]}})
    paths = ["internal/game/battlespotting.go", "internal/game/battlespotting_test.go", "internal/game/game.go",
             "internal/game/battle.go", "internal/game/battlemarker.go", "internal/rules/spotting.go",
             "internal/rules/battleview.go", "docs/internals/data/screen-battle.json", "tools/scripts/battle_discovery_scroll.py"]
    assert all(Path(v).exists() for v in paths)
    data["remake_validation"] = {"comparison": "same-state", "entry": "unique diagnostic RAM adapter; subsequent original frames are assertions only",
        "input_mapping": "one continue/update for every original 1588:0437 completed redraw; actual Tick separately checked",
        "canvas": [640, 480], "scaling": "none", "masking": "none", "worktree": "dirty",
        "gui_image_id": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8",
        "source_sha256": {v: sha(Path(v)) for v in paths}, "frames": results, "excluded": excluded,
        "normal_player": {"comparison": "remake-normal-player", "original_parity_pass": False,
            "route": "main menu/password138/S3/strategic Execute/battle Execute; two turns; full SHIP27456/FLEET3904 Save/load", "frames": normal},
        "before_method": "parent 649bef6 spotting implementation via read-only Go overlay; same corrected fixtures; missed-frame diagnostic only",
        "commands": [["tools/gui.sh", "xrun", "env", "LOGH3_MOVEMENT_OUTPUT=round43-normal", "go", "test", "./internal/game", "-run", "^TestBattleSpotting", "-count=1", "-v"],
                     ["tools/gui.sh", "xrun", "go", "test", "./internal/...", "./cmd/...", "-count=1"],
                     ["tools/gui.sh", "xrun", "python3", "tools/scripts/battle_discovery_scroll.py", "render"]],
        "test_logs": {str(v): sha(v) for v in [Path("workplace/round43-final-tests.log"), Path("workplace/round43-corrected-fixture-before-tests.log"), Path("workplace/round43-standard-tests.log")]}}
    write(p, data)
    print("正式UI", len(results), "SHIP", sum(v["units"] for v in results), "FLEET", sum(v["fleets"] for v in results), "零差異；SHA-256", sha(p))


if __name__ == "__main__":
    if sys.argv[1] == "freeze":
        freeze()
    elif sys.argv[1] == "collect":
        collect(tuple(sys.argv[2:]) if len(sys.argv) > 2 else ("a", "b"))
    elif sys.argv[1] == "render":
        render()
    elif sys.argv[1] == "flags":
        result, count = flags(sys.argv[2])
        p = Path("workplace") / (sys.argv[2] + "-flags.txt")
        assert not p.exists()
        p.write_text("\n".join(result) + "\n")
        print(count)
