"""把戰場的隊形槽位表（`ds:59BA`）倒出來（`docs/re/126`）。

    tools/py.sh tools/scripts/formation_slots.py [--emit]

41 個槽位，交錯的格點：x 從 64 到 −64 每 16 一欄，偶數欄 5 個
（y ＝ −64…64 每 32）、奇數欄 4 個（y ＝ −48…48）。槽 20 是 (0,0)，
旗艦就站那裡。
"""
import json
import sys

DUMP = "/work/workplace/s3datamem.bin"
DS = 0x54B3 * 16
OUT = "/work/docs/internals/data/battle-formation-slots.json"
COUNT = 41

d = open(DUMP, "rb").read()


def s16(a):
    v = d[a] | d[a + 1] << 8
    return v - 0x10000 if v >= 0x8000 else v


slots = [{"index": i,
          "x": s16(DS + 0x59BA + i * 4),
          "y": s16(DS + 0x59BA + i * 4 + 2)} for i in range(COUNT)]

cols = {}
for s in slots:
    cols.setdefault(s["x"], []).append(s["y"])
for x in sorted(cols, reverse=True):
    print("x=%4d  y = %s" % (x, " ".join("%4d" % y for y in cols[x])))
print("共 %d 個槽位；槽 20 ＝ (%d,%d)"
      % (len(slots), slots[20]["x"], slots[20]["y"]))

if "--emit" in sys.argv:
    doc = {
        "_note": "戰場的隊形槽位表（原版 ds:59BA，一筆四個位元組 x,y）。"
                 "分隊的目的地 ＝ 艦隊位置 ＋ 這個偏移依艦隊朝向旋轉"
                 "（2BE1:0059，docs/re/126）。艦艇記錄 +0x05 就是槽索引。",
        "address": "ds:59BA",
        "count": COUNT,
        "centre": 20,
        "slots": slots,
    }
    with open(OUT, "w", encoding="utf-8") as fh:
        json.dump(doc, fh, ensure_ascii=False, indent=1)
        fh.write("\n")
    print("寫出", OUT)
