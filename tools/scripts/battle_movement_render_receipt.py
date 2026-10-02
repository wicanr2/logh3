"""#79第二階段畫面診斷；原版來源371、正式資料驗證044、結果045。

資料與鏡頭固定，階段面板／游標相位尚未對齊；全畫面按nearby-state記錄。
保留全部差異，區域由產品screen-battle.json推導，不建立測試座標副本。
執行：LOGH3_EXTRA_RO=<專案根> tools/py.sh tools/scripts/battle_movement_render_receipt.py <HEAD> <工作樹>
"""
import hashlib
import json
import os
from pathlib import Path
import sys

from pngout import game_palette, write_png

root, work = Path("/ref"), Path("workplace")
assert work.stat().st_uid == os.getuid()
prefix = sys.argv[3] if len(sys.argv) > 3 else "battle79-round32-diagnostic"
normal_prefix = sys.argv[4] if len(sys.argv) > 4 else "battle79-round32-normal-fixed"
assert prefix.startswith("battle79-round32-") and "/" not in prefix
assert normal_prefix.startswith("battle79-round32-") and "/" not in normal_prefix


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


evidence = Path("docs/internals/data/battle-movement-phase.json")
assert sha(evidence) == "b920eb42ccc70c8c447a9a2bd53a5e40f9efae5e807929fe4972fa3cc5fbe516"
data = json.loads(evidence.read_text())
spec = json.loads(Path("docs/internals/data/screen-battle.json").read_text())
f, m, s = (spec[key] for key in ("field", "minimap", "status"))
regions = {
    "full": [0, 0, 640, 480],
    "field": [f["x"], f["y"], f["cols"] * f["tile_w"], f["rows"] * f["tile_h"]],
    "minimap": [m[key] for key in ("x", "y", "w", "h")],
    "status": [s[key] for key in ("x", "y", "w", "h")],
}
sources = ["internal/game/battlemovement_test.go", "internal/game/battlemovement.go",
           "internal/game/battlecmd.go", "internal/game/game.go", "internal/game/battle.go",
           "internal/rules/battlemovement.go", "internal/rules/battle.go",
           "internal/screen/battle.go", "docs/internals/data/screen-battle.json",
           "tools/scripts/battle_movement_render_receipt.py"]
report = {
    "schema": "logh3.movement-render-diagnostic.v1", "remake_head": sys.argv[1],
    "remake_working_tree": sys.argv[2], "remake_image": "logh3-ebiten:2",
    "command": "bash tools/gui.sh xrun sh -c 'LOGH3_MOVEMENT_OUTPUT=" + prefix + " go test -count=1 -v -run \"TestGameMovementPhaseMatchesOriginal|TestScoutRecoverAndRelaunchNormal\" ./internal/game'",
    "source_sha256": {p: sha(root / p) for p in sources},
    "original_program_sha256": data["program_sha256"], "dosgolem_head": data["dosgolem_head"],
    "original_assets_sha256": {p: sha(Path("/orig") / p) for p in ("SPRITE.SPR", "BLOCK_2.BLK", "ANI_PNL0.A98", "GRAPH.IMG", "MAP26.CMP")},
    "image_ids": {"logh3-ebiten:2": "sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8", "golang:1.24-bookworm": "sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac"},
    "seed_original_and_remake": "0x12345678，移動前固定；本移動階段不消費亂數",
    "classification": "nearby-state；完整移動資料與鏡頭相同，階段面板及游標相位未對齊；不宣稱完整same-state畫面通過",
    "camera": {"source": "原版DS:9BF4/9BF6，DS線性54B30，研究326", "world": [1056, 600], "cells": [66, 30]},
    "canvas": [640, 480], "crop": None, "scale": 1, "masks": [], "regions": regions,
    "comparisons": [], "normal_frames": []}


def save(name, px, pal):
    p = work / (prefix + "-" + name + ".png")
    if p.exists():
        assert p.stat().st_uid == os.getuid()
    write_png(str(p), 640, 480, px, pal)
    return {"path": str(p), "sha256": sha(p)}


for mode in ("entry", "end"):
    receipt = data["phase_receipts"][mode]
    original = work / ("oracle/battle79-move-" + mode + "-a")
    again = work / ("oracle/battle79-move-" + mode + "-b")
    for name in ("mem-in.bin", "mem-out.bin", "screen.bin", "screen.png", "pal.bin"):
        assert sha(original / name) == sha(again / name) == receipt["hashes"][name], (mode, name)
    a = (original / "screen.bin").read_bytes()
    actual = work / (prefix + "-phase-" + mode + ".bin")
    b = actual.read_bytes()
    pa = game_palette(str(original / "screen.png"))
    dac = actual.with_suffix(".pal")
    pb = [v << 2 | v >> 4 for v in dac.read_bytes()]
    assert len(a) == len(b) == 640 * 480
    assert len(pa) == len(pb) == 48
    differences = [i for i in range(len(a)) if pa[a[i] * 3:a[i] * 3 + 3] != pb[b[i] * 3:b[i] * 3 + 3]]
    stats = {}
    for key, (x, y, w, h) in regions.items():
        inside = [i for i in differences if x <= i % 640 < x + w and y <= i // 640 < y + h]
        stats[key] = {
            "pixels": w * h, "equal": w * h - len(inside), "different": len(inside),
            "max_channel_difference": max((abs(pa[a[i] * 3 + c] - pb[b[i] * 3 + c]) for i in inside for c in range(3)), default=0),
            "bounding_box": [min(i % 640 for i in inside), min(i // 640 for i in inside), max(i % 640 for i in inside), max(i // 640 for i in inside)] if inside else None}
    diff = bytearray(len(a))
    for i in differences:
        diff[i] = 1
    report["comparisons"].append({
        "mode": mode, "original_receipt": receipt,
        "original_metadata_sha256": sha(original / "meta.txt"),
        "remake_frame_sha256": sha(actual), "remake_palette_sha256": sha(dac),
        "effective_palette_equal": pa == pb, "region_stats": stats,
        "original": save(mode + "-original", a, pa), "remake": save(mode + "-remake", b, pb),
        "difference": save(mode + "-difference", diff, [0, 0, 0, 255, 0, 255] + [0] * 42)})
    print(mode, json.dumps(stats, ensure_ascii=False))

for label in ("scout-deployed", "scout-away", "scout-recovered", "scout-relaunched"):
    p = work / (normal_prefix + "-" + label + ".bin")
    dac = p.with_suffix(".pal")
    pixels = p.read_bytes()
    assert len(pixels) == 640 * 480
    pal = [v << 2 | v >> 4 for v in dac.read_bytes()]
    report["normal_frames"].append({"label": label, "raw_sha256": sha(p),
        "palette_sha256": sha(dac), "png": save("normal-" + label, pixels, pal),
        "scope": "正常滑鼠放出、歸隊、再放出；原池81與單位31；沒有狀態注入"})

output = work / (prefix + "-render-receipt.json")
if output.exists():
    assert output.stat().st_uid == os.getuid()
output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print("畫面診斷收據SHA-256", sha(output))
