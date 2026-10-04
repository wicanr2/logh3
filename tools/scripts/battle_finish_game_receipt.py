"""凍結110正式規則、Game報告與保存來源。只在Docker內執行。"""
import hashlib
import json
import struct
from pathlib import Path


sources = {}


def read(path, expected=None):
    p = Path(path)
    b = p.read_bytes()
    sha = hashlib.sha256(b).hexdigest()
    if expected is not None:
        assert sha == expected, str(p)
    sources[str(p)] = {"size": len(b), "sha256": sha}
    return b


def png_metadata(path):
    """像素由正式Go測試逐點核對；收尾另驗PNG尺寸與16色。"""
    b = read(path)
    assert b[:8] == b"\x89PNG\r\n\x1a\n"
    at, size, palette = 8, None, None
    while at < len(b):
        n = struct.unpack(">I", b[at:at+4])[0]
        kind, content = b[at+4:at+8], b[at+8:at+8+n]
        if kind == b"IHDR":
            w, h, depth, color = struct.unpack(">IIBB", content[:10])
            assert (w, h) == (640, 480) and color == 3 and depth in (4, 8)
            size = (w, h)
        elif kind == b"PLTE":
            palette = content[:48]
        at += n + 12
    assert size is not None and palette is not None and len(palette) == 48
    return palette


def original(stem, expected, count):
    a = read("workplace/" + stem + "-a.json", expected)
    assert a == read("workplace/" + stem + "-b.json", expected)
    d = json.loads(a)
    assert d["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
    assert not d["dosgolem_dirty"] and len(d["rows"]) == count
    return d


base = original("finish79-r89", "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477", 60)
guard = original("finish79-r90-guard", "85fbd698622c307051183167d2315e9977abe5cac8fc8068611b31f4287ee335", 7)
cursor = original("finish79-r91-cursor", "7ae71c1bea82bf045d5083b9326e41c1defdde899355164d1df2d699a82978f8", 2)
read("org_game/银河英雄传说3SP.zip", "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673")
read("org_game/extracted/GIN3.COM", "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918")
read("workplace/oracle/cleanup79-r88-entry-a/entry.state", "a62af47fb7d15a136e2e2ed51760ef0e9dd82538f68d5f7e0be4db5f2b0011b8")
read("workplace/oracle/cleanup79-r88-entry-a/mem-in.bin", "1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf")

targeted = read("workplace/finish79-r92-targeted.log").decode()
assert "--- FAIL:" not in targeted and "--- SKIP:" not in targeted
for name in ("BattleFinishMatchesOriginal", "WinnerOccupationCaptureWaitMatchesOriginal",
             "WinnerOccupationRejectsBeforeMutation", "BattleFinishGameOriginal",
             "BattleFinishGameCaptureWait", "BattleFinishSaveReload", "NormalFleetPlayerChain"):
    assert "--- PASS: Test" + name + " " in targeted, name
standard = read("workplace/finish79-r92-standard.log").decode()
assert "FAIL" not in standard
assert sum(line.startswith("ok  \t") for line in standard.splitlines()) == 13
assert standard.count("[no test files]") == 5

render = Path("workplace/finish79-r92-render")
comparisons = []
for p in sorted(render.glob("*-comparison.json")):
    d = json.loads(read(p))
    w = d["window"]
    assert d["mask"] is None and w["rect"] == [32, 180, 416, 140]
    assert w["pixels"] == w["same"] == 58240
    assert w["different"] == w["maximum_channel_difference"] == 0 and w["bbox"] is None
    prefix = str(p)[:-len("-comparison.json")]
    palette = png_metadata(prefix + "-original.png")
    assert png_metadata(prefix + "-remake.png") == palette
    assert png_metadata(prefix + "-diff.png") == palette
    comparisons.append({"path": str(p), **d})
assert len(comparisons) == 50

rows = base["rows"] + guard["rows"]
save_cases = []
for row in base["rows"]:
    if row["mask"] not in (1, 126) or row["human"] == 0 or row["round"] != 5:
        continue
    prefix = render / (row["name"] + "-save")
    for name, key in {"ADMIRAL": "admiral", "FLEET": "fleet", "SHIP": "ship",
                      "STELLAR": "stellar", "WORK": "work", "PLANET": "planet"}.items():
        want = bytes.fromhex(row["after"][key])
        if name == "PLANET":
            want = want[64:]
        assert read(prefix / (name + ".DT0")) == want, (row["name"], name)
    read(prefix / "SAVEDATA.IDX")
    save_cases.append(row["name"])
assert len(save_cases) == 6

ida_path = "workplace/ida/ownership79-r82-3826.json"
ida = json.loads(read(ida_path, "cfb1671c41cfc0ee40b52485ea325fce812b83bd42c7a3addabb0b2522b5e06f"))
assert ida["kernel_version"] == "9.4" and ida["ida_base_hex"] == "0x10200"
meanings = {"3826:040C": "讀收件座位word", "3826:045E": "FFFF座位選caller頭像；其餘選座位",
            "3826:0465": "保留原運算元[bx-70B3h]；DS:8F4D[座位]是實際情報參謀頭像",
            "3826:046D": "FFFF座位只讀caller頭像低byte", "3826:0470": "頭像高byte清零",
            "3826:0480": "將選出的實際頭像交給3958:0215"}
instructions = []
for item in ida["instructions"]:
    if item["original_location"] in meanings:
        instructions.append({**item, "grade": "proven", "warning": "",
                             "semantic": meanings[item["original_location"]],
                             "evidence": [ida_path, "workplace/finish79-r89-a.json",
                                          "workplace/finish79-r92-targeted.log"]})
assert len(instructions) == 6

# 保存全部目前產品Go來源與直接使用的資料，舊RE索引保留原雜湊。
for top in ("internal", "cmd"):
    for p in sorted(Path(top).rglob("*.go")):
        read(p)
for p in sorted(Path("docs/internals/data").glob("*.json")):
    if p.name not in ("battle-finish-game.json",):
        read(p)
for pattern in ("MAP*.CMP", "*.A98"):
    for p in sorted(Path("org_game/extracted").glob(pattern)):
        read(p)
for p in ("tools/scripts/battle_finish_formats.py", "tools/scripts/battle_finish_game_receipt.py",
          "tools/battle_finish_oracle.sh", "tools/battle_finish_guard_oracle.sh",
          "tools/battle_finish_cursor_oracle.sh", "tools/battle_finish_native_run.sh",
          "tools/battle_finish_guard_native_run.sh", "tools/battle_finish_cursor_native_run.sh",
          "tools/go.sh", "tools/gui.sh", "go.mod", "go.sum"):
    read(p)

result = {
    "schema": "logh3.battle-finish-game.v1", "issue": 79,
    "parent_commit": "8b8b2d73a9e634537ad0c4b67daab53b8d024866", "worktree_dirty": True,
    "build_images": ["golang:1.24-bookworm", "logh3-ebiten:2"],
    "dosgolem_head": base["dosgolem_head"], "dosgolem_dirty": False,
    "original_metadata": [{k: v for k, v in d.items() if k != "rows"} for d in (base, guard, cursor)],
    "summary": {"rule_cases": 67, "game_cases": 67, "callback_cases": 2, "actual_callback_waits": 1,
                "invalid_rule_cases": 6, "report_windows": 50, "window_pixels": 2912000,
                "save_cases": 6, "normal_player_factions": 2,
                "standard_pass": 13, "standard_without_tests": 5,
                "table_and_list_bytes_per_layer": sum(sum(len(bytes.fromhex(v)) for v in r["after"].values()) for r in rows)},
    "cases": [{"name": r["name"], "mask": r["mask"], "human": r["human"], "round": r["round"],
               "steps": r["steps"], "seed_after_hex": r["seed_after_hex"],
               "after_sha256": {k: hashlib.sha256(bytes.fromhex(v)).hexdigest() for k, v in r["after"].items()}}
              for r in rows],
    "save_cases": save_cases, "comparisons": comparisons,
    "instructions": instructions,
    "limits": "原版固定8-5診斷與明示注入，完整輸入／注入／亂數／時間序列在原始JSON。same-state只限報告矩形；全圖差異保存。六檔為StateSaveGame診斷。正常兩側入口／交戰回歸通過，尚非自然勝利／占領GUI或完整062同態。110 READY、062 DRAFT、#79 OPEN。",
    "sources": [{"path": p, **s} for p, s in sorted(sources.items())],
}
dest = Path("docs/internals/data/battle-finish-game.json")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
assert dest.stat().st_uid == 1000
print("正式結束來源", len(sources), "SHA-256", hashlib.sha256(dest.read_bytes()).hexdigest())
print("摘要", json.dumps(result["summary"], ensure_ascii=False))
