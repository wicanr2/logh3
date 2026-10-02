"""#79第一階段正式Game.Frame收據；入口docs/playtest/042。

只驗規格事先約定的完整結果面板；全畫面差異照實保留，不冒稱整畫面一致。
tools/py.sh掛專案唯讀至/ref，argv提供remake HEAD及工作樹狀態。
"""
import hashlib
import json
import os
from pathlib import Path
import re
import sys

from pngout import game_palette, write_png


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


w = Path("workplace")
root = Path("/ref")
assert w.stat().st_uid == os.getuid(), "輸出目錄擁有權錯誤"
spec = json.loads(Path("docs/internals/data/supply-panel.json").read_text())
geometry = (root / "internal/ui/frameblk.go").read_text()
constant = lambda key: int(re.search(r"\b" + key + r"\s*=\s*(\d+)", geometry)[1])
tw, th, bias = (constant(k) for k in ("FrameTileW", "FrameTileH", "FrameCellOriginY"))
rect = [spec["cell_x"] * tw, spec["cell_y"] * th + bias, spec["cols"] * tw, spec["rows"] * th]
fixture = Path("docs/internals/data/battle-preparation-lifecycle.json")
assert sha(fixture) == "f8d4848a3e19232a66f92b18c2477647ea375a36379d2b5c792e3f5f48f1e650"
batches = {b["mode"]: b for b in json.loads(fixture.read_text())["batches"]}
sources = ["internal/game/battlepreparation.go", "internal/game/battlepreparation_test.go",
           "internal/game/battle.go", "internal/game/game.go", "internal/game/intro.go",
           "internal/game/supplypick.go", "internal/rules/battlepreparation.go",
           "internal/rules/battlepreparation_test.go", "internal/screen/supplypanel.go",
           "docs/internals/data/supply-panel.json", "internal/ui/frameblk.go",
           "tools/scripts/battle_preparation_render_receipt.py"]
report = {"schema": "logh3.preparation-render.v1", "remake_head": sys.argv[1],
          "remake_working_tree": sys.argv[2], "remake_image": "logh3-ebiten:2",
          "remake_command": "bash tools/gui.sh test ./internal/game -run TestBattlePreparation -count=1 -v",
          "source_sha256": {p: sha(root / p) for p in sources},
          "seed_original_and_remake": "0x12345678；執行前固定；資源階段不消費亂數",
          "scope": "診斷注入的S3達希利第6回合：結果面板same-state，首尾資源欄位另由Go測試驗證；不宣稱全畫面同狀態。",
          "canvas": [640, 480], "scale": 1, "region": rect, "comparisons": []}


def save(name, px, pal):
    path = w / ("battle79-round28-" + name + ".png")
    if path.exists():
        assert path.stat().st_uid == os.getuid(), "既有輸出擁有權錯誤"
    write_png(str(path), 640, 480, px, pal)
    return {"path": str(path), "sha256": sha(path)}


for mode in ("supply", "repair"):
    original = w / ("oracle/battle79-life-" + mode + "-panel-a")
    raw = original / "screen.bin"
    assert sha(raw) == batches[mode + "-panel"]["screen_sha256"]
    a = raw.read_bytes()
    actual = w / ("battle79-round28-" + mode + "-panel.bin")
    b = actual.read_bytes()
    assert len(a) == len(b) == 640 * 480
    pa = game_palette(str(original / "screen.png"))
    dac = w / ("battle79-round28-" + mode + "-panel.pal")
    pb = [v << 2 | v >> 4 for v in dac.read_bytes()]
    assert pa == pb, "兩側有效色盤不同"
    pixels = [i for i, (aa, bb) in enumerate(zip(a, b)) if aa != bb]
    x, y, width, height = rect
    inside = [i for i in pixels if x <= i % 640 < x + width and y <= i // 640 < y + height]

    def stats(items, total):
        return {"pixels": total, "equal": total - len(items), "different": len(items),
                "max_channel_difference": max((abs(pa[a[i] * 3 + c] - pb[b[i] * 3 + c]) for i in items for c in range(3)), default=0),
                "bounding_box": [min(i % 640 for i in items), min(i // 640 for i in items),
                                 max(i % 640 for i in items), max(i // 640 for i in items)] if items else None}

    diff = bytearray(len(a))
    for i in pixels:
        diff[i] = 1
    report["comparisons"].append({"mode": mode, "original_batch": batches[mode + "-panel"],
        "original_meta": {"path": str(original / "meta.txt"), "sha256": sha(original / "meta.txt")},
        "original_frame_sha256": sha(raw), "remake_frame_sha256": sha(actual),
        "remake_dac_sha256": sha(dac), "effective_palette_equal": True,
        "original": save(mode + "-original", a, pa), "remake": save(mode + "-remake", b, pb),
        "difference": save(mode + "-difference", diff, [0, 0, 0, 255, 0, 255] + [0] * 42),
        "full_canvas": stats(pixels, len(a)), "result_panel": stats(inside, width * height),
        "outside_classification": "鏡頭、單位、日期／動態狀態及其他面板未納入本輪同狀態完成範圍；不掩蓋全畫面差異。"})
    assert not inside, "完整結果面板未一致"

normal = w / "battle79-round28-normal-gui.png"
report["normal_gui"] = {"path": str(normal), "sha256": sha(normal),
    "command": "bash tools/gui.sh run ./cmd/gin3 -orig /orig -data /src/docs/internals/data -script 'move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 544 125; click 560 125; tick 2000; move 600 240; click; click; tick 4000' -shot-bin /src/workplace/battle79-round28-normal-gui.bin -shot /src/workplace/battle79-round28-normal-gui.png -shot-after 2",
    "scope": "無注入正常入口；已目視達希利第2回合，非原版全畫面對拍。"}
output = w / "battle79-round28-render-receipt.json"
if output.exists():
    assert output.stat().st_uid == os.getuid()
output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print("完整結果面板各20480像素、有效色盤相同；收據SHA-256", sha(output))
