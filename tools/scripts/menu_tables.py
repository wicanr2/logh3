"""把主程式裡的選單表全部解出來。

    tools/py.sh tools/scripts/menu_tables.py [--emit]

輸入是收據 `s3datamem` 的 1 MB 記憶體傾印。選單由 `3667:0004` 開啟，
呼叫端把表的位址用 `mov bx, imm16 / push bx / push ss` 推進去，所以
掃 `9A 04 00 67 36` 再往回找那三條指令就能列出全部的表。

一筆六個位元組：可選旗標（byte，0 ＝ 反白不可選）＋ 補 1 byte ＋
指向名稱的遠指標（位移:段）。段一律是資料段。
"""
import json
import re
import sys

DUMP = "/work/workplace/s3datamem.bin"
DS = 0x54B3                      # 主程式的資料段（`docs/internals/03`）
DSBASE = DS * 16
OPEN = bytes([0x9A, 0x04, 0x00, 0x67, 0x36])   # call far 3667:0004

d = open(DUMP, "rb").read()

def word(lin):
    return d[lin] | d[lin + 1] << 8

def big5(lin):
    end = d.index(b"\0", lin)
    return d[lin:end].decode("big5", "replace")

def table_at(off, limit=12):
    """從資料段位移 off 解出一張選單表。"""
    items = []
    for i in range(limit):
        base = DSBASE + off + i * 6
        flag, pad = d[base], d[base + 1]
        soff, sseg = word(base + 2), word(base + 4)
        if sseg != DS or flag > 1 or pad != 0:
            break
        name = big5(DSBASE + soff)
        if not name or len(name) > 200:
            break
        items.append({"可選": flag, "名稱": name, "名稱位移": soff})
    return items

sites = []
for m in re.finditer(re.escape(OPEN), d):
    h = m.start()
    back = d[h - 60:h]
    off = None
    for i in range(len(back) - 4):
        if back[i] == 0xBB and back[i + 3] == 0x53 and back[i + 4] == 0x16:
            off = back[i + 1] | back[i + 2] << 8
    if off is None:            # 表的位址不是用立即值推進去的
        sites.append({"呼叫端": h, "表": None, "項目": []})
        continue
    sites.append({"呼叫端": h, "表": off, "項目": table_at(off)})

for s in sites:
    tbl = "ds:%04X" % s["表"] if s["表"] is not None else "（非立即值）"
    names = "／".join("%s%s" % (it["名稱"], "" if it["可選"] else "（不可選）")
                      for it in s["項目"])
    print("%05X  %-12s %s" % (s["呼叫端"], tbl, names))

if "--at" in sys.argv:
    for a in sys.argv[sys.argv.index("--at") + 1:]:
        if a.startswith("--"):
            break
        off = int(a, 16)
        items = table_at(off)
        print("ds:%04X  %s" % (off, "／".join(
            "%s%s" % (it["名稱"], "" if it["可選"] else "（不可選）") for it in items)
            or "（解不出項目）"))
        base = DSBASE + off
        print("        原始 24 bytes:", " ".join("%02x" % b for b in d[base:base + 24]))

if "--emit" in sys.argv:
    out = "/work/docs/internals/data/menus.json"
    doc = {
        "說明": "主程式的選單表；開啟常式 3667:0004，一筆六個位元組"
                "（可選旗標 byte ＋ 補 1 byte ＋ 名稱遠指標）。位移是資料段位移。",
        "資料段": "%04X" % DS,
        "選單": [{"呼叫端線性位址": "%05X" % s["呼叫端"],
                  "表": ("ds:%04X" % s["表"]) if s["表"] is not None else None,
                  "項目": s["項目"]} for s in sites],
    }
    with open(out, "w", encoding="utf-8") as fh:
        json.dump(doc, fh, ensure_ascii=False, indent=2)
        fh.write("\n")
    print("寫出", out)
