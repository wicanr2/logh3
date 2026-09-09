"""把 EMS 分頁框裡的兩張 32×32 距離矩陣倒成 JSON。

    tools/py.sh tools/scripts/distance_table.py <8KB傾印> <輸出json>

傾印是 `-dump-mem-at '<步數>:lin:E0000:8192:…'` 取的分頁框內容
（`docs/re/241`）。版面：

    E0000  帝國 32×32 word（ds:D062）
    E0800  同盟 32×32 word（ds:D066）
    E1000  帝國中點矩陣 32×32 byte（ds:D06A）
    E1400  同盟中點矩陣（ds:D06E）

**先用兩張中點矩陣核對**：與 `star-graph.json` 逐位元組相同才算數，
否則那一刻頁框裡放的是別的頁。
"""
import json
import sys
import collections

dump = open(sys.argv[1], "rb").read()
out = sys.argv[2]
graph = json.load(open("docs/internals/data/star-graph.json", encoding="utf-8"))

mid = {"empire": dump[0x1000:0x1400], "alliance": dump[0x1400:0x1800]}
for side, key in (("empire", "midpoint_empire"), ("alliance", "midpoint_alliance")):
    want = bytes(b for row in graph[key] for b in row)
    if mid[side] != want:
        raise SystemExit("%s 的中點矩陣對不上——那一刻頁框裡不是這幾張表" % side)

adj = graph["adjacency"]
hops = [[-1] * 32 for _ in range(32)]
for s in range(32):
    hops[s][s] = 0
    q = collections.deque([s])
    while q:
        u = q.popleft()
        for v in adj[u]:
            if hops[s][v] < 0:
                hops[s][v] = hops[s][u] + 1
                q.append(v)

tables = {}
for side, base in (("empire", 0x0000), ("alliance", 0x0800)):
    m = [[int.from_bytes(dump[base + (r * 32 + c) * 2:base + (r * 32 + c) * 2 + 2],
                         "little") for c in range(32)] for r in range(32)]
    tables[side] = m

# 形狀：對角線、對稱性、高位元組與步數的關係。
note = {}
for side, m in tables.items():
    diag = all(m[i][i] == 0 for i in range(32))
    sym = all(m[r][c] == m[c][r] for r in range(32) for c in range(32))
    hi = sorted({v >> 8 for row in m for v in row})
    pairs = sorted({(v >> 8, hops[r][c]) for r, row in enumerate(m)
                    for c, v in enumerate(row)})
    note[side] = {"diagonal_zero": diag, "symmetric": sym,
                  "high_bytes": hi,
                  "max": max(max(r) for r in m),
                  "cells_ge_0x800": sum(1 for r in m for v in r if v >= 0x800),
                  "high_byte_vs_hops": ["%d/%d" % p for p in pairs][:24]}

e, a = tables["empire"], tables["alliance"]
note["between_tables"] = {
    "same_low_byte_cells": sum(1 for r in range(32) for c in range(32)
                               if (e[r][c] & 0xFF) == (a[r][c] & 0xFF)),
    "high_byte_diff": sorted({(a[r][c] >> 8) - (e[r][c] >> 8)
                              for r in range(32) for c in range(32)}),
}

json.dump({
    "source": "EMS 分頁框 E0000／E0800（ds:D062／ds:D066），收據見 docs/re/241",
    "layout": "32×32 word，索引 [起點 × 32 + 終點]；2B99:028F 查它",
    "note": "載入的當下整張是 0，路徑計算跑過才填（docs/re/43）",
    "shape": note,
    "empire": tables["empire"],
    "alliance": tables["alliance"],
}, open(out, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
open(out, "a", encoding="utf-8").write("\n")
print("寫出", out)
for side in ("empire", "alliance"):
    print(side, "對角線 0" if note[side]["diagonal_zero"] else "對角線非 0",
          "對稱" if note[side]["symmetric"] else "不對稱",
          "高位元組", note[side]["high_bytes"],
          "最大", note[side]["max"], ">=0x800", note[side]["cells_ge_0x800"])
print("兩表低位元組相同的格數", note["between_tables"]["same_low_byte_cells"],
      "高位元組差", note["between_tables"]["high_byte_diff"])
