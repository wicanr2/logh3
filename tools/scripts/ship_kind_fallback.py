"""把 ds:295C 的艦種替代表倒成 JSON。

    tools/py.sh tools/scripts/ship_kind_fallback.py <1MB傾印> <輸出.json>

`0FF4:1350(艦隊, 艦種)` 拿 `ds:295C + 艦種 × 6` 當起點，逐個往下試，
第一個「在那一格編得出至少一組」的就是替代艦種；`0xFF` 結束
（`docs/re/216`）。
"""
import json
import sys

DS = 0x54B30
BASE = DS + 0x295C
KINDS = 8
STRIDE = 6

d = open(sys.argv[1], "rb").read()
out = {
    "source": "ds:295C（執行期資料段），一個艦種 6 bytes 的起點",
    "reader": "0FF4:1350：逐個試，第一個編得出至少一組的就用；0xFF 結束",
    "note": "清單以 0xFF 結尾，長度不固定；艦種 7 的六個 byte 裡沒有 0xFF，"
            "結尾落在下一格（那一格的第一個 byte 就是 0xFF）",
    "kinds": [],
}
for k in range(KINDS):
    a = BASE + k * STRIDE
    raw = list(d[a:a + STRIDE])
    seq, i = [], a
    while i < a + 32 and d[i] != 0xFF:
        seq.append(d[i])
        i += 1
    out["kinds"].append({"kind": k, "raw": raw, "order": seq})
with open(sys.argv[2], "w", encoding="utf-8") as f:
    json.dump(out, f, ensure_ascii=False, indent=1)
    f.write("\n")
for e in out["kinds"]:
    print("艦種 %d  raw %s  → %s" % (e["kind"],
          " ".join("%02X" % b for b in e["raw"]),
          " ".join("%02X" % b for b in e["order"])))
