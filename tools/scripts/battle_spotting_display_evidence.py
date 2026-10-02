"""#79 索敵顯示最小契約。prepare 重建固定計畫；collect 核對獨立原版收據。

入口與完整命令見 docs/playtest/046-battle-spotting-phase.md。
原始指令保留原名、bytes 與三種位址，語意只從受版控的分級索引合併。
"""
import hashlib
import json
import pathlib
import shutil
import sys

ROOT = pathlib.Path(__file__).resolve().parents[2]
INDEX = ROOT / "docs/internals/data/battle-spotting-display-semantics.json"
OUTPUT = ROOT / "docs/internals/data/battle-spotting-display.json"
WORK = ROOT / "workplace/explore"


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def render_game_phase():
    """唯一入口RAM轉接，正式Game邊界結果；不改第39／40輪原版收據。"""
    from pngout import game_palette, write_png
    geometry = json.loads((ROOT / "docs/internals/data/screen-battle.json").read_text())
    f, m, status = geometry["field"], geometry["minimap"], geometry["status"]
    regions = {"full": [0, 0, 640, 480],
               "field": [f["x"], f["y"], f["cols"] * f["tile_w"], f["rows"] * f["tile_h"]],
               "minimap": [m["x"], m["y"], m["w"], m["h"] + 1],
               "status": [status["x"], status["y"], status["w"], status["h"]]}
    frames, excluded = [], []
    paths = ["docs/internals/data/battle-spotting-phase.json",
             "docs/internals/data/battle-discovery-evidence.json"]
    for source in paths:
        data = json.loads((ROOT / source).read_bytes())
        for case in data["cases"]:
            base = ROOT / case["runs"][0]["path"]
            other = ROOT / case["runs"][1]["path"]
            for frame in case["frames"]:
                name = frame["name"]
                if name in ("entry", "camera-start", "seat1"):
                    excluded.append({"scenario": case["scenario"], "frame": name,
                                     "reason": "寫階段字前／常式內部邊界；不是正式renderer停點"})
                    continue
                original = base / (name + ".bin")
                old = original.read_bytes()
                expected = frame["hashes"].get(".bin", frame["hashes"].get(name + ".bin"))
                assert digest(old) == expected
                assert old == (other / (name + ".bin")).read_bytes()
                original_png = base / (name + ".png")
                expected_png = frame["hashes"].get(".png", frame["hashes"].get(name + ".png"))
                expected_mem = frame["hashes"].get("-mem.bin", frame["hashes"].get(name + "-mem.bin"))
                assert digest(original_png.read_bytes()) == expected_png
                original_mem = base / (name + "-mem.bin")
                assert digest(original_mem.read_bytes()) == expected_mem
                assert original_mem.read_bytes() == (other / (name + "-mem.bin")).read_bytes()
                assert original_png.read_bytes() == (other / (name + ".png")).read_bytes()
                remake = ROOT / "workplace" / ("round42-spot-" + case["scenario"] + "-" + name)
                new = remake.with_suffix(".bin").read_bytes()
                dac = remake.with_suffix(".pal").read_bytes()
                assert len(old) == len(new) == 640 * 480 and len(dac) == 48
                palette = bytes((v << 2) | (v >> 4) for v in dac)
                assert palette == bytes(game_palette(str(original_png)))
                assert max(max(old), max(new)) < 16
                write_png(remake.with_suffix(".png"), 640, 480, new, palette)
                delta = remake.with_name(remake.name + "-diff").with_suffix(".png")
                changed = bytes(int(a != b) for a, b in zip(old, new))
                write_png(delta, 640, 480, changed, bytes([0, 0, 0, 255, 0, 0]))
                stats = {}
                for label, (x, y, w, h) in regions.items():
                    points = [j * 640 + i for j in range(y, y + h) for i in range(x, x + w)]
                    diff = [p for p in points if changed[p]]
                    stats[label] = {"rect": [x, y, w, h], "compared": len(points),
                                    "equal": len(points) - len(diff), "different": len(diff),
                                    "bbox_inclusive": ([min(p % 640 for p in diff), min(p // 640 for p in diff),
                                                        max(p % 640 for p in diff), max(p // 640 for p in diff)] if diff else None),
                                    "max_rgb_channel_difference": max((abs(palette[3 * old[p] + c] - palette[3 * new[p] + c])
                                                                       for p in diff for c in range(3)), default=0)}
                files = [original, original_png, base / (name + "-mem.bin"), base / "pal.bin",
                         remake.with_suffix(".bin"), remake.with_suffix(".pal"), remake.with_suffix(".png"), delta]
                frames.append({"scenario": case["scenario"], "frame": name, "seed": frame["seed"],
                               "camera": frame.get("camera_grid", frame.get("camera_pixels")),
                               "units": len(frame["units"]), "fleets": len(frame["fleets"]),
                               "stats": stats, "effective_palette_equal": True,
                               "files": {str(p.relative_to(ROOT)): digest(p.read_bytes()) for p in files}})
    assert all(frame["stats"]["full"]["different"] == 0 for frame in frames)
    normal = []
    for label in ("spotting-header", "spotting-completed"):
        prefix = ROOT / "workplace" / ("round42-normal-" + label)
        pixels, dac = prefix.with_suffix(".bin").read_bytes(), prefix.with_suffix(".pal").read_bytes()
        assert len(pixels) == 640 * 480 and len(dac) == 48
        write_png(prefix.with_suffix(".png"), 640, 480, pixels, bytes((v << 2) | (v >> 4) for v in dac))
        normal.append({"frame": label, "files": {str(prefix.with_suffix(ext).relative_to(ROOT)):
                        digest(prefix.with_suffix(ext).read_bytes()) for ext in (".bin", ".pal", ".png")}})
    sources = paths + ["internal/game/battlespotting.go", "internal/game/battlespotting_test.go",
                       "internal/game/battle.go", "internal/game/battlemarker.go", "internal/game/game.go",
                       "internal/game/intro.go", "internal/rules/battle.go", "internal/rules/spotting.go",
                       "internal/screen/battle.go", "docs/internals/data/screen-battle.json",
                       "internal/game/battledata_test.go", "tools/scripts/battle_spotting_display_evidence.py"]
    result = {"schema": "logh3.battle-spotting-game.v1", "comparison": "same-state",
              "scope": "唯一原版入口的全SHIP／FLEET／WORK／PLANET、seed、實際及已顯示鏡頭；正式Game逐觀測者UI邊界",
              "entry_method": "diagnostic RAM adapter; production parser, rule, renderer and phase continuation",
              "input_mapping": "beginBattleSpotting對應181B:0262；continue對應原版固定命令的UI呼叫邊界，等待tick另驗，不宣稱相同wall-clock",
              "display_state": "入口物件+6/+8重建最後已畫鏡頭，正式Frame畫背景後恢復RAM鏡頭；其餘幀只讀期望，不回注",
              "mouse": {"s1": [100, 150], "others": [240, 240]},
              "parent_git_head": "c749176730b4c94b6f235650db1c2a98d34acaaf", "worktree": "dirty",
              "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_worktree": "clean",
              "go_image_id": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac",
              "gui_image_id": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8",
              "canvas": [640, 480], "scaling": "none", "masking": "none", "frames": frames,
              "excluded_internal_boundaries": excluded,
              "normal_player_sample": {"comparison": "remake-normal-player", "original_parity_pass": False,
                  "path": "battleTestGame正常主選單、防拷138、S3與戰略實行至戰場，再兩次左鍵實行；第一次戰術實行前固定seed12345678一次",
                  "injection": "既有入口fixture；不注入情報、報告或指定觀測者，不使用索敵direct-entry",
                  "result": "首回合有發現標頭；完整SHIP/FLEET正式存讀與第二次實行通過",
                  "frames": normal},
              "commands": [["tools/gui.sh", "xrun", "go", "test", "./internal/game", "-run", "^TestBattleSpotting", "-count=1", "-v"],
                           ["tools/gui.sh", "xrun", "env", "LOGH3_MOVEMENT_OUTPUT=round42-normal", "go", "test", "./internal/game",
                            "-run", "^TestBattleSpottingNormalWaitAndSave$", "-count=1", "-v"],
                           ["tools/gui.sh", "xrun", "python3", "tools/scripts/battle_spotting_display_evidence.py", "phase-render"]],
              "original_commands": {case["scenario"]: case["runs"] for source in paths
                                     for case in json.loads((ROOT / source).read_bytes())["cases"]},
              "source_sha256": {p: digest((ROOT / p).read_bytes()) for p in sources},
              "limits": ["正常玩家報告與保存另驗，診斷不取代正常入口", "艦隊／偵察隊來源標頭尚未有本輪獨立原版畫面對拍",
                         "可比UI邊界不等於原版116個內部停點全覆蓋", "完整八步與發行未完成"]}
    output = ROOT / "docs/internals/data/battle-spotting-game.json"
    output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    print(len(frames), "可比邊界；SHIP", sum(f["units"] for f in frames), "FLEET", sum(f["fleets"] for f in frames),
          "全幀及逐區零差異；未比較內部停點", len(excluded), "SHA-256", digest(output.read_bytes()))


def render_receipt():
    from pngout import write_png
    geometry = json.loads((ROOT / "docs/internals/data/screen-battle.json").read_text())
    f, m = geometry["field"], geometry["minimap"]
    regions = {"full": [0, 0, 640, 480],
               "field": [f["x"], f["y"], f["cols"] * f["tile_w"], f["rows"] * f["tile_h"]],
               "minimap": [m["x"], m["y"], m["w"], m["h"] + 1]}
    frames = []
    for label in ("scout-deployed", "scout-away", "scout-recovered", "scout-relaunched"):
        before = ROOT / "workplace" / ("round41-scout-before-" + label)
        after = ROOT / "workplace" / ("round41-scout-" + label)
        old, new = before.with_suffix(".bin").read_bytes(), after.with_suffix(".bin").read_bytes()
        op, np = before.with_suffix(".pal").read_bytes(), after.with_suffix(".pal").read_bytes()
        assert len(old) == len(new) == 640 * 480
        assert op == np and len(np) % 3 == 0
        assert len(np) >= 3 * (max(max(old), max(new)) + 1)
        for path, pixels in ((before.with_suffix(".png"), old), (after.with_suffix(".png"), new)):
            write_png(path, 640, 480, pixels, np)
        diff = bytes(1 if a != b else 0 for a, b in zip(old, new))
        delta = after.with_name(after.name + "-diff").with_suffix(".png")
        write_png(delta, 640, 480, diff, bytes([0, 0, 0, 255, 0, 0]) + bytes(762))
        stats = {}
        for name, (x, y, w, h) in regions.items():
            points = [j * 640 + i for j in range(y, y + h) for i in range(x, x + w)]
            changed = [p for p in points if old[p] != new[p]]
            bounds = None
            if changed:
                xs, ys = [p % 640 for p in changed], [p // 640 for p in changed]
                bounds = [min(xs), min(ys), max(xs), max(ys)]
            stats[name] = {"rect": [x, y, w, h], "compared": len(points),
                           "equal": len(points) - len(changed), "different": len(changed),
                           "bbox_inclusive": bounds,
                           "max_rgb_channel_difference": max((abs(np[3 * old[p] + c] - np[3 * new[p] + c])
                              for p in changed for c in range(3)), default=0)}
        files = [before.with_suffix(ext) for ext in (".bin", ".pal", ".png")]
        files += [after.with_suffix(ext) for ext in (".bin", ".pal", ".png")] + [delta]
        frames.append({"label": label, "stats": stats,
                       "files": {str(p.relative_to(ROOT)): digest(p.read_bytes()) for p in files}})
    source_paths = ["internal/rules/battle.go", "internal/rules/spotting.go",
                    "internal/game/battlemovement_test.go", "docs/internals/data/screen-battle.json",
                    "workplace/round41-before-battle-adapted.go", "workplace/round41-sprite-overlay.json"]
    result = {"comparison": "remake-before-after", "original_parity_pass": False,
              "scope": "主選單既有fixture輸入至戰場，正常索敵命令／實行／歸隊／再放出；不注入回收旗標",
              "parent_git_head": "80af23933b8c31357242bdbd647fe469ae1f67ef", "worktree": "dirty",
              "seed": "0x12345678，第一次戰術實行前一次設定", "canvas": [640, 480],
              "scaling": "none", "masking": "none", "frames": frames,
              "toolchain": json.loads(INDEX.read_bytes())["toolchain"],
              "commands": [
                  ["tools/gui.sh", "xrun", "env", "LOGH3_MOVEMENT_OUTPUT=round41-scout",
                   "go", "test", "./internal/game", "-run", "^TestScoutRecoverAndRelaunchNormal$", "-count=1", "-v"],
                  ["tools/gui.sh", "xrun", "env", "LOGH3_MOVEMENT_OUTPUT=round41-scout-before",
                   "go", "test", "-overlay=/src/workplace/round41-sprite-overlay.json", "./internal/game",
                   "-run", "^TestScoutRecoverAndRelaunchNormal$", "-count=1", "-v"],
                  ["tools/gui.sh", "xrun", "python3", "tools/scripts/battle_spotting_display_evidence.py", "render"]],
              "source_sha256": {p: digest((ROOT / p).read_bytes()) for p in source_paths}}
    path = ROOT / "workplace/round41-scout-render.json"
    path.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    for frame in frames:
        print(frame["label"], frame["stats"]["full"]["different"], frame["stats"]["full"]["bbox_inclusive"])
    print("四張完整修正前後圖與差分", digest(path.read_bytes()))


def main():
    if len(sys.argv) > 1 and sys.argv[1] == "phase-render":
        render_game_phase()
        return
    if len(sys.argv) > 1 and sys.argv[1] == "render":
        render_receipt()
        return
    index_raw = INDEX.read_bytes()
    index = json.loads(index_raw)
    probe = ROOT / "tools/scripts/battle_step_probe.go"
    assert digest(probe.read_bytes()) == index["probe_sha256"]
    if len(sys.argv) > 1 and sys.argv[1] in ("prepare", "prepare-recovered"):
        recovered = sys.argv[1] == "prepare-recovered"
        prefix = "recovered_" if recovered else ""
        plan = WORK / ("battle79-spot-r41-recovered-plan.json" if recovered else "battle79-spot-r41-final-plan.json")
        script = WORK / "battle79-spot-r41-probe.go"
        assert not plan.exists(), "禁止覆寫研究輸入"
        if script.exists():
            assert digest(script.read_bytes()) == index["probe_sha256"]
        else:
            shutil.copyfile(probe, script)
        plan.write_text(json.dumps(index[prefix + "plan"], ensure_ascii=False, indent=2) + "\n")
        assert digest(plan.read_bytes()) == index[prefix + "plan_sha256"]
        print("固定計畫與既有probe已準備")
        return
    a = WORK / "battle79-spot-r41-final-a.json"
    b = WORK / "battle79-spot-r41-final-b.json"
    raw = a.read_bytes()
    assert raw == b.read_bytes(), "原版獨立雙跑不同"
    assert digest(raw) == index["receipt_sha256"]
    receipt = json.loads(raw)
    recovered_raw = (WORK / "battle79-spot-r41-recovered-a.json").read_bytes()
    assert recovered_raw == (WORK / "battle79-spot-r41-recovered-b.json").read_bytes()
    assert digest(recovered_raw) == index["recovered_receipt_sha256"]
    recovered = json.loads(recovered_raw)
    assert recovered["plan"] == index["recovered_plan"]
    assert recovered["plan_sha256"] == index["recovered_plan_sha256"]
    assert recovered["dosgolem_head"] == index["dosgolem_head"]
    assert len(recovered["rows"]) == 2
    for row in recovered["rows"]:
        assert row["seed_after_hex"] == "78563412"
        before, after = (bytes.fromhex(row[k]["scout"]) for k in ("before", "after"))
        assert len(before) == len(after) == 78
        assert before[30] == after[30] == 255
        assert {i for i, (x, y) in enumerate(zip(before, after)) if x != y} <= {25, 26, 27, 28, 29}
        assert row["before"]["fleet"] == row["after"]["fleet"]
    assert receipt["plan"] == index["plan"]
    assert receipt["plan_sha256"] == index["plan_sha256"]
    assert receipt["dosgolem_head"] == index["dosgolem_head"]
    assert receipt["input_sha256"] == index["runtime_input_sha256"]
    assert len(receipt["rows"]) == 61
    for row in receipt["rows"]:
        assert row["seed_after_hex"] == "78563412"
        for label, size in (("unit", 78), ("fleet", 122), ("object", 40),
                            ("observer", 78), ("candidate", 78)):
            assert len(bytes.fromhex(row["before"][label])) == size
            assert len(bytes.fromhex(row["after"][label])) == size
        if row["name"].startswith("cache-"):
            for label, allowed in (("unit", {47}), ("fleet", {106}), ("object", set())):
                before, after = (bytes.fromhex(row[k][label]) for k in ("before", "after"))
                assert {i for i, (x, y) in enumerate(zip(before, after)) if x != y} <= allowed
        elif row["name"].startswith("visibility-"):
            for label in ("unit", "fleet", "observer", "candidate"):
                assert row["before"][label] == row["after"][label]
        else:
            assert row["before"]["observer"] == row["after"]["observer"]
            before, after = (bytes.fromhex(row[k]["candidate"]) for k in ("before", "after"))
            assert {i for i, (x, y) in enumerate(zip(before, after)) if x != y} <= {22, 23, 24}
    ram_path = ROOT / "workplace/oracle/battle79-second-r36-a/entry-mem.bin"
    ram = ram_path.read_bytes()
    assert digest(ram) == index["ida_runtime_input_sha256"]
    entries = {(e["kind"], e["original_location"]): e for e in index["entries"]}
    functions, exports = [], []
    for name in index["exports"]:
        path = ROOT / "workplace/ida" / (name + ".json")
        data_raw = path.read_bytes()
        data = json.loads(data_raw)
        segment = int(data["functions"][0]["original_location"][:4], 16)
        source = bytes(256) + ram[segment * 16:segment * 16 + 8192]
        assert data["input_sha256"] == digest(source)
        assert data["tool"] == "IDA Pro 9.4"
        exports.append({"path": str(path.relative_to(ROOT)), "sha256": digest(data_raw),
                        "input_sha256": digest(source), "tool": data["tool"],
                        "address_space": data["address_space"], "ida_base_hex": data["ida_base_hex"],
                        "command_argv": index["export_commands"][name]})
        for function in data["functions"]:
            for kind, record in [("function", function)] + [("instruction", i) for i in function["instructions"]]:
                if kind == "instruction":
                    linear = int(record["runtime_linear_hex"], 16)
                    original = bytes.fromhex(record["original_bytes_hex"])
                    assert ram[linear:linear + len(original)] == original
                entry = entries.get((kind, record["original_location"]))
                if entry:
                    if kind == "instruction":
                        assert record["original_bytes_hex"] == entry["original_bytes_hex"]
                    for key in ("semantic", "grade", "source"):
                        record[key] = entry[key]
                    record["warning"] = "" if entry["grade"] == "proven" else "⚠ 未證實語意"
            functions.append(function)
    output = {"schema": "logh3.battle-spotting-display.v1", "index_sha256": digest(index_raw),
              "toolchain": index["toolchain"],
              "receipt_sha256": digest(raw), "receipt": receipt, "exports": exports,
              "recovered_receipt_sha256": digest(recovered_raw), "recovered_scout": recovered,
              "functions": functions, "limits": ["原版常式診斷，畫面不是正式UI證據",
              "完整第三階段Game續接、逐筆renderer、正常玩家路徑及存讀另驗"]}
    OUTPUT.write_text(json.dumps(output, ensure_ascii=False, indent=2) + "\n")
    print("61案及2案回收後完整收據雙跑相同；原始指令與分級定位重生", digest(OUTPUT.read_bytes()))


if __name__ == "__main__":
    main()
