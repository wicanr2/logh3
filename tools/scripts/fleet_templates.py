"""把 ds:2BC6 的艦隊編制樣板表倒成 JSON。

    tools/py.sh tools/scripts/fleet_templates.py <1MB傾印> <輸出.json>

一個類別鍵 20 bytes ＝ 10 組 × (值[2], 值[3])，填進艦隊編制槽的
`+2`／`+3`（`docs/re/212`）。
"""
import json
import sys

DS = 0x54B30
BASE = DS + 0x2BC6
KEYS = 16
SLOTS = 10

d = open(sys.argv[1], "rb").read()
out = {
    "source": "ds:2BC6（執行期資料段），一個艦隊類別鍵 20 bytes",
    "layout": "10 組 × (slot2, slot3)；0FF4:11F4 把它們寫進艦隊編制槽的 +2 與 +3",
    "keys": [],
}
for k in range(KEYS):
    a = BASE + k * SLOTS * 2
    out["keys"].append({
        "key": k,
        "slots": [{"slot2": d[a + i * 2], "slot3": d[a + i * 2 + 1]}
                  for i in range(SLOTS)],
    })
with open(sys.argv[2], "w", encoding="utf-8") as f:
    json.dump(out, f, ensure_ascii=False, indent=1)
    f.write("\n")
print("寫出 %s（%d 個類別鍵 × %d 組）" % (sys.argv[2], KEYS, SLOTS))
