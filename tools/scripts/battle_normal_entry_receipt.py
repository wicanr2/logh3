"""#79正常玩家入口比較；入口不同時只保存nearby-state診斷，不升格。

tools/gui.sh xrun python3 tools/scripts/battle_normal_entry_receipt.py
"""
import hashlib
import json
import os
import struct
from collections import Counter
from pathlib import Path

from pngout import game_palette, write_png


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, text):
    assert p.parent.is_dir() and p.parent.stat().st_uid == os.getuid()
    if p.exists():
        assert p.stat().st_uid == os.getuid()
    p.write_text(text)


original = [Path("workplace/oracle/battle79-normal-r44-" + r) for r in "ab"]
remake = [Path("workplace/round44-normal-fixed-" + r) for r in "ab"]
original_names = ["loaded-mem.bin", "entry-mem.bin", "screen.bin", "screen.png", "pal.bin"]
for n in original_names:
    assert (original[0] / n).read_bytes() == (original[1] / n).read_bytes(), n
remake_names = ["probe.json"] + [s + "/" + n for s in ("loaded", "entry") for n in
    ("STELLAR.DT0", "PLANET.DT0", "ADMIRAL.DT0", "FLEET.DT0", "SHIP.DT0", "WORK.DT0", "SAVEDATA.IDX", "screen.bin")]
for n in remake_names:
    assert (remake[0] / n).read_bytes() == (remake[1] / n).read_bytes(), n
nosave = Path("workplace/round44-normal-nosave")
nosave_names = ["entry/" + n for n in ("STELLAR.DT0", "PLANET.DT0", "ADMIRAL.DT0", "FLEET.DT0", "SHIP.DT0", "WORK.DT0", "screen.bin")]
for n in nosave_names:
    assert (remake[0] / n).read_bytes() == (nosave / n).read_bytes(), n
meta = json.loads((remake[0] / "probe.json").read_text())
assert meta["strategy_seed_observed_before"] == meta["strategy_seed_before"] == "D3DC167E"
report = (original[0] / "report.txt").read_text()
assert "#300000000 改記憶體 lin:5E00C（5E00C）4 bytes" in report
flags = (original[0] / "flags.txt").read_text().splitlines()
assert flags[flags.index("-poke") + 1] == "lin:5E00C@300000000=7E 16 DC D3"
mems = {s: (original[0] / (s + "-mem.bin")).read_bytes() for s in ("loaded", "entry")}
assert all(len(b) == 1048576 for b in mems.values())
tables = [("FLEET", 0x60BA4, 122, 32), ("SHIP", 0x668B2, 78, 352),
          ("STELLAR", 0x5F2CC, 126, 32), ("WORK", 0x5DA44, 230, 1)]
stages = {}
for s, m in mems.items():
    rows = {}
    for name, base, stride, count in tables:
        r = (remake[0] / s / (name + ".DT0")).read_bytes()
        o = m[base:base + count * stride]
        assert len(o) == len(r) == count * stride
        diffs = [i for i, (a, b) in enumerate(zip(o, r)) if a != b]
        records = []
        for i in range(count):
            a, b = o[i * stride:(i + 1) * stride], r[i * stride:(i + 1) * stride]
            records.append({"index": i, "original_hex": a.hex(), "remake_hex": b.hex(),
                            "different_offsets": [j for j in range(stride) if a[j] != b[j]]})
        rows[name] = {"bytes": len(o), "different_bytes": len(diffs),
                      "offset_counts": dict(sorted(Counter(i % stride for i in diffs).items())), "records": records}
    stages[s] = rows
assert all(stages["loaded"][n]["different_bytes"] == 0 for n in ("FLEET", "SHIP", "STELLAR"))
m = mems["entry"]
ds = 0x54B30
count = struct.unpack_from("<H", m, ds + 0x9894)[0]
fleets = list(m[ds + 0x9898:ds + 0x9898 + count])
raw = (remake[0] / "entry/FLEET.DT0").read_bytes()
remake_fleets = [i for i in range(32) if raw[i * 122 + 4] != 255 and raw[i * 122 + 6] == 26]
assert fleets == [12, 16, 17, 28, 29, 30, 31] and remake_fleets == [12, 28, 31]
assert stages["loaded"]["WORK"]["records"][0]["different_offsets"] == [12, 14, 16]
original_bins = (original[0] / "screen.bin").read_bytes()
remake_bins = (remake[0] / "entry/screen.bin").read_bytes()
assert len(original_bins) == len(remake_bins) == 640 * 480
assert len((original[0] / "pal.bin").read_bytes()) == 768
# 讀原版PNG的有效色盤，包含屬性暫存器映射；不把DAC直接當色號色盤。
pal = bytes(game_palette(str(original[0] / "screen.png")))
assert len(pal) == 48


write_png("workplace/round44-normal-remake-original-palette.png", 640, 480, remake_bins, pal)
write_png("workplace/round44-normal-entry-diff.png", 640, 480,
          bytes(int(a != b) for a, b in zip(original_bins, remake_bins)), bytes([0, 0, 0, 255, 0, 0]))
pair = b"".join(original_bins[y * 640:(y + 1) * 640] + remake_bins[y * 640:(y + 1) * 640] for y in range(480))
write_png("workplace/round44-normal-entry-pair.png", 1280, 480, pair, pal)
regions = {}
for name, rect in {"full": (0, 0, 640, 480), "field": (0, 0, 448, 360)}.items():
    x0, y0, x1, y1 = rect
    offsets = [y * 640 + x for y in range(y0, y1) for x in range(x0, x1)]
    different = [i for i in offsets if original_bins[i] != remake_bins[i]]
    n = len(different)
    rgb_different = [i for i in different if pal[original_bins[i] * 3:original_bins[i] * 3 + 3] != pal[remake_bins[i] * 3:remake_bins[i] * 3 + 3]]
    regions[name] = {"rectangle": rect, "compared_pixels": len(offsets), "same_pixels": len(offsets) - n,
                     "different_index_pixels": n, "different_rgb_pixels": len(rgb_different),
                     "maximum_channel_difference": max((abs(pal[original_bins[i] * 3 + k] - pal[remake_bins[i] * 3 + k]) for i in different for k in range(3)), default=0),
                     "difference_bbox_inclusive": [min(i % 640 for i in different), min(i // 640 for i in different), max(i % 640 for i in different), max(i // 640 for i in different)] if different else None}
source_names = ["tools/battle_normal_entry_oracle.sh", "tools/scripts/battle_normal_probe.go",
                "tools/scripts/battle_normal_probe_seed.py", "tools/scripts/battle_normal_entry_receipt.py",
                "internal/game/game.go", "internal/game/originalbrain.go", "internal/game/rebellion.go",
                "internal/bootstrap/load.go"]
runs = []
for p in original:
    names = original_names + ["entry.state", "flags.txt", "meta.txt", "report.txt"]
    runs.append({"path": str(p), "hashes": {n: sha(p / n) for n in names},
                 "command_argv": (p / "flags.txt").read_text().splitlines()})
data = {"schema": "logh3.normal-battle-entry.v1", "classification": "nearby-state",
        "grade": "proven for measured entry differences; upstream cause unknown",
        "parent_head": "5751b76bce200b3b229fafcda51fd39a2bbd3907",
        "remake_worktree": "dirty documentation and probes; production sources unchanged",
        "toolchain": {"go_image": "golang:1.24-bookworm", "go_image_id": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac",
                      "gui_image": "logh3-ebiten:2", "gui_image_id": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8"},
        "dosgolem_head": "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c", "dosgolem_dirty": False,
        "program_sha256": sha(Path("org_game/extracted/GIN3.COM")),
        "bundle_sha256": "de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673",
        "original_runs": runs, "remake_runs": [{"path": str(p), "hashes": {n: sha(p / n) for n in remake_names}} for p in remake],
        "original_reproducible_outputs": original_names, "remake_reproducible_outputs": remake_names,
        "save_side_effect_control": {"path": str(nosave), "equal_outputs": nosave_names, "hashes": {n: sha(nosave / n) for n in nosave_names}},
        "snapshot_files_equal": (original[0] / "entry.state").read_bytes() == (original[1] / "entry.state").read_bytes(),
        "snapshot_limit": "snapshot檔不相同，不以全檔宣稱決定性；1MiB RAM、畫面與色盤獨立雙跑相同",
        "remake_seed_and_inputs": meta, "original_seed_after": f"{struct.unpack_from('<I', m, 0x5E00C)[0]:08X}",
        "original_turn_at_entry": struct.unpack_from("<H", m, 0x5DA48)[0],
        "remake_turn_at_entry": struct.unpack_from("<H", (remake[0] / "entry/WORK.DT0").read_bytes(), 4)[0],
        "original_fleets": fleets, "remake_fleets": remake_fleets,
        "original_pool_count_including_unreleased_scouts": struct.unpack_from("<H", m, ds + 0x9896)[0],
        "remake_pool_count_including_unreleased_scouts": sum(raw[i * 122 + 15] + 1 for i in remake_fleets),
        "original_camera_grid": list(m[ds + 0x9BA3:ds + 0x9BA5]),
        "tables": stages, "regions": regions,
        "images": {n: sha(Path(n)) for n in ("workplace/round44-normal-remake-original-palette.png",
                   "workplace/round44-normal-entry-diff.png", "workplace/round44-normal-entry-pair.png")},
        "source_hashes": {n: sha(Path(n)) for n in source_names},
        "limits": ["未通過正常玩家same-state；087保持READY", "WORK座位差異與日期差異不得遮罩",
                   "原版與remake初始seed相同，後續seed不同；未要求跨完整流程骰序一致",
                   "完整table包含未參戰紀錄與執行器指標，不將每個byte差異當作玩法bug",
                   "PNG比較使用同一原版色盤，屬nearby-state診斷；不是正式產品RGB parity",
                   "未修改產品；未重擲seed；未以原版畫面或RAM回注Game"]}
target = Path("docs/internals/data/battle-normal-entry.json")
write(target, json.dumps(data, ensure_ascii=False, indent=2) + "\n")
print("原版5份輸出／remake17份輸出雙跑相同；正常入口7艦隊對3艦隊，未通過same-state")
print(target, sha(target)); print(json.dumps(regions, ensure_ascii=False))
