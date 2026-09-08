"""產生戰場命令選單的版面檔。

    tools/py.sh tools/scripts/battlecmd_spec.py [--emit]

字碼直接讀收據 `s3datamem` 的記憶體，不手打。版面公式與五張表的來源
是 `docs/re/117`；框內的配色與框線與戰術選單同一套
（`docs/internals/data/battle-menu.json`）。
"""
import json
import sys

DUMP = "/work/workplace/s3datamem.bin"
DS = 0x54B3
DSBASE = DS * 16
OUT = "/work/docs/internals/data/battle-command-menus.json"

d = open(DUMP, "rb").read()
word = lambda a: d[a] | d[a + 1] << 8

MENUS = [
    ("fleet",    0x4120, "旗艦（艦艇 +0x03 ＝ 0）"),
    ("squadron", 0x413C, "分隊，艦種 0–4"),
    ("supply",   0x4158, "艦種 5（運補艦）"),
    ("repair",   0x417A, "艦種 6（修護艦）"),
    ("scout",    0x4196, "艦種 7（索敵分遣隊）"),
]
# 每一張選單的內寬（16 像素為單位）。原版把這個值寫死成 2（`docs/re/117`）。
WIDTH = 2


def items(off, n):
    out = []
    for i in range(n):
        base = DSBASE + off + i * 6
        soff = word(base + 2)
        raw = d[DSBASE + soff:d.index(b"\0", DSBASE + soff)]
        codes, k = [], 0
        while k < len(raw):
            if raw[k] >= 0x80 and k + 1 < len(raw):
                codes.append(raw[k] << 8 | raw[k + 1])
                k += 2
            else:                       # 半形字：碼就是位元組
                codes.append(raw[k])
                k += 1
        out.append({"text": raw.decode("big5"), "codes": codes})
    return out


COUNT = {"fleet": 3, "squadron": 2, "supply": 3, "repair": 3, "scout": 2}

doc = {
    "_note": "戰場的命令選單（docs/re/117）。開啟常式 3667:0004 與戰術選單同一支，"
             "配色與框線也同一套，所以這裡只放每一張的項目與內寬；"
             "版面走下面的 layout 公式。",
    "width": WIDTH,
    "menus": {},
    "layout": {
        "_measured_from": "s3cmdsq3（分隊 62×78）、s3cmdfl3（艦隊 62×98）、"
                          "s3menu（戰術 126×178）三張逐一驗過",
        "cell_w": 16,
        "cell_h": 20,
        "origin_bias": 1,
        "cursor_cell_bias": -1,
        "clamp_rows": 19,
        "frame_pad_w": 30,
        "frame_pad_h": 38,
        "highlight_dx": 15,
        "highlight_dy": 19,
        "text_dy": 2,
        "col_pitch": 16,
        "row_pitch": 20,
        "ink": 7,
        "shadow": 3,
        "highlight_ink": 2,
        "disabled_ink": 8,
        "fill": 9,
        "frame": "ui.MenuFrame",
    },
}
for key, off, when in MENUS:
    doc["menus"][key] = {
        "address": "ds:%04X" % off,
        "when": when,
        "items": items(off, COUNT[key]),
    }

for key, m in doc["menus"].items():
    print("%-9s %s  %s" % (key, m["address"],
                           "／".join(it["text"] for it in m["items"])))
    for it in m["items"]:
        print("            %s → %s" % (it["text"],
                                       " ".join("%04X" % c for c in it["codes"])))

if "--emit" in sys.argv:
    with open(OUT, "w", encoding="utf-8") as fh:
        json.dump(doc, fh, ensure_ascii=False, indent=1)
        fh.write("\n")
    print("寫出", OUT)
