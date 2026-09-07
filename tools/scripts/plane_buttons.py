"""從字元平面的 JSON 找出按鈕，印成可以貼進畫面規格的 JSON。

按鈕在原版裡不是一個物件，是樣式旗標拼出來的（`docs/internals/31`）：
左端一格 `0x10`、中間文字格 `0x08`、右端一格 `0x20`。矩形換算成像素是
「左端格的最後兩個像素」到「右端格的前兩個像素」。

    tools/py.sh tools/scripts/plane_buttons.py /work/workplace/x.json
"""
import json
import sys

STYLE_LEFT, STYLE_TEXT, STYLE_RIGHT = 0x10, 0x08, 0x20


def main(path):
    d = json.load(open(path, encoding="utf-8"))
    items = []
    for row in d["rows"]:
        cells = row["cells"]
        i = 0
        while i < len(cells):
            if cells[i]["style"] & STYLE_LEFT == 0:
                i += 1
                continue
            j = i + 1
            text, codes = "", []
            while j < len(cells) and cells[j]["style"] & STYLE_TEXT:
                c = cells[j]
                if c["text"] and c["text"].strip():
                    text += c["text"]
                    codes.append(c["code"])
                j += 1
            if j < len(cells) and cells[j]["style"] & STYLE_RIGHT:
                x0 = cells[i]["x"] + 14
                x1 = cells[j]["x"] + 1
                first = next((c for c in cells[i + 1:j]
                              if c["text"] and c["text"].strip()), cells[i + 1])
                items.append({
                    "x": x0, "y": row["y"], "w": x1 - x0 + 1, "h": 20,
                    "text_x": first["x"], "text": text, "codes": codes,
                })
                i = j
                continue
            i += 1
    print(json.dumps(items, ensure_ascii=False, indent=1))


if __name__ == "__main__":
    main(sys.argv[1])
