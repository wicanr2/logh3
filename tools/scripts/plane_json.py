"""把原版的全螢幕字元平面（`ds:D81E`）解成 JSON。

平面是 40 欄 × 20 列、每格 6 位元組（格式見 `docs/internals/31`）：

    +0  低 5 位＝視窗編號，0x80 髒、0x40 用快取字形、0x10 走彩色文字路徑
    +1  另一組模式旗標（bit0 半形）
    +2  字碼低位   +3  字碼高位（Big5 首位元組在這裡；0xFF?? 是內建框線圖）
    +4  屬性：高 4 位底色、低 4 位字色
    +5  樣式：0x01 底線、0x02 網點、0x04 加粗、0x08 立體上下邊

一格 16×20 像素，畫面 640×400 置中放進 640×480，所以第 r 列的螢幕
`y = 40 + 20r`、第 c 欄的 `x = 16c`。

    tools/py.sh tools/scripts/plane_json.py /work/plane.bin /work/plane.json
"""
import json
import sys

W, H, STRIDE = 40, 20, 6


def main(src, dst):
    d = open(src, "rb").read()
    if len(d) < W * H * STRIDE:
        raise SystemExit("平面只有 %d bytes，要 %d" % (len(d), W * H * STRIDE))
    rows = []
    for r in range(H):
        cells = []
        for c in range(W):
            b = d[(r * W + c) * STRIDE:(r * W + c) * STRIDE + STRIDE]
            code = b[3] << 8 | b[2]
            if b[3] >= 0xA1:
                try:
                    text = bytes([b[3], b[2]]).decode("big5")
                except UnicodeDecodeError:
                    text = None
            elif b[3] == 0xFF:
                text = None
            else:
                text = bytes([b[3], b[2]]).decode("latin1")
            cells.append({
                "col": c, "x": c * 16, "y": 40 + r * 20,
                "owner": b[0] & 0x1F, "flags": b[0], "mode": b[1],
                "code": code, "hex": "%04X" % code, "text": text,
                "ink": b[4] & 0x0F, "bg": b[4] >> 4, "style": b[5],
            })
        rows.append({"row": r, "y": 40 + r * 20, "cells": cells})
    line = lambda cs: "".join(
        c["text"] if c["text"] else ("[%04X]" % c["code"] if c["code"] else "  ")
        for c in cs)
    out = {
        "note": "原版全螢幕字元平面 ds:D81E，格式見 docs/internals/31",
        "cols": W, "rows_count": H, "cell_w": 16, "cell_h": 20, "origin_y": 40,
        "text": [line(r["cells"]) for r in rows],
        "rows": rows,
    }
    json.dump(out, open(dst, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    for i, t in enumerate(out["text"]):
        print("%02d |%s|" % (i, t))


if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2])
