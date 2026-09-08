"""產生戰場右側分隊面板的版面檔（`docs/re/118`）。

    tools/py.sh tools/scripts/unitpanel_spec.py [--emit]

字碼與各張名稱表都直接讀收據 `s3datamem` 的記憶體，不手打。
"""
import json
import sys

DUMP = "/work/workplace/s3datamem.bin"
DS = 0x54B3
DSBASE = DS * 16
OUT = "/work/docs/internals/data/unit-panel.json"

d = open(DUMP, "rb").read()
word = lambda a: d[a] | d[a + 1] << 8


def codes(s):
    """Big5 位元組串 → 字碼；全形一個碼，半形一個碼。"""
    out, k = [], 0
    while k < len(s):
        if s[k] >= 0x80 and k + 1 < len(s):
            out.append(s[k] << 8 | s[k + 1])
            k += 2
        else:
            out.append(s[k])
            k += 1
    return out


def text(off):
    """一條字串；codes 只留前面那段靜態字（碰到 `%`、顏色跳脫或換行就停）。"""
    raw = d[DSBASE + off:d.index(b"\0", DSBASE + off)]
    cut = len(raw)
    for k, b in enumerate(raw):
        if b in (0x25, 0x03, 0x0A):     # '%'、\x03、\n
            cut = k
            break
    return {"text": raw.decode("big5", "replace"),
            "codes": codes(raw[:cut].rstrip(b" "))}


def table(off, n, strip=False):
    """一張「四個位元組一筆、前兩個是字串位移」的名稱表。"""
    out = []
    for i in range(n):
        so = word(DSBASE + off + i * 4)
        raw = d[DSBASE + so:d.index(b"\0", DSBASE + so)]
        if strip:                      # 陣營名前後包著顏色跳脫
            raw = raw.split(b",", 1)[1].rsplit(b"\x03", 1)[0]
        out.append({"text": raw.decode("big5", "replace"), "codes": codes(raw)})
    return out


spec = {
    "_note": "戰場右側的分隊面板（docs/re/118）。點自己陣營的分隊時換上來，"
             "與狀態欄同一個框、同一組字元格。八列，列距 20。",
    "cell": {"x": 480, "y": 242, "col": 16, "row": 20, "half": 8, "rows": 8},
    "ink": 7,
    "morale_low_ink": 2,
    "morale_low_below": 50,
    "faction_ink": [b for b in d[DSBASE + 0x6096:DSBASE + 0x6096 + 4]],
    "labels": {
        "di": text(0x3C70),            # 第
        "fleet": text(0x3C54),         # 艦隊
        "unit": text(0x3C73),          # 部隊（後面補空白）
        "scout_unit": text(0x3C5F),    # 索敵部隊
        "fleet_morale": text(0x41B0),  # 艦隊士氣值
        "unit_morale": text(0x41BD),   # 部隊士氣值
        "cannot_fight": text(0x41CF),  # 無法戰鬥
        "fleet_action": text(0x41E0),  # 艦隊行動（含顏色跳脫）
        "unit_action": text(0x41F5),   # 部隊行動
        "beam": text(0x4231),          # 光線砲
        "missile": text(0x4240),       # 飛彈
        "fighter": text(0x4250),       # 戰鬥機
        "supply": text(0x4227),        # 補給物資
        "repair_max": text(0x4205),    # 最高修理數
        "fuel_left": text(0x4216),     # 剩餘燃料
    },
    "factions": table(0x60F4, 2, strip=True),
    "fleet_actions": table(0x691C, 5),
    "fleet_action_ink": [b & 0xF for b in d[DSBASE + 0x6930:DSBASE + 0x6935]],
    "squadron_actions": table(0x694A, 4),
    "weapon_ink": {"full": 7, "part": 5, "low": 2},
    # 每一段落在第幾個半形格上（x ＝ cell.x ＋ 半形格 × 8）。走原版的
    # 格式字串算出來的：全形一個字兩格、半形一格。
    "halves": {
        "faction": 0, "di": 8, "fleet_no": 10, "fleet_word": 12,
        "unit_di": 0, "unit_no": 2, "unit_word": 4,
        "admiral": 0,
        "kind_name": 0, "ships": 12, "ships_width": 4,
        "morale_label": 0, "morale_value": 12, "morale_width": 4,
        "action_label": 0, "action_value": 12,
        "weapon_label": 0, "weapon_value": 10, "weapon_width": 6,
    },
}

for k, v in spec["labels"].items():
    print("%-14s %r" % (k, v["text"]))
print("陣營色", spec["faction_ink"])
print("陣營名", [f["text"] for f in spec["factions"]])
print("艦隊行動", [a["text"] for a in spec["fleet_actions"]], spec["fleet_action_ink"])
print("分隊行動", [a["text"] for a in spec["squadron_actions"]])

if "--emit" in sys.argv:
    with open(OUT, "w", encoding="utf-8") as fh:
        json.dump(spec, fh, ensure_ascii=False, indent=1)
        fh.write("\n")
    print("寫出", OUT)
