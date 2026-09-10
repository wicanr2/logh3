"""重算 AI 指派內政專任參謀時挑的是誰（`2D10:0213` 那條規則）。

    tools/py.sh tools/scripts/staff_pick.py <1MB傾印> <陣營> [首都星域]

提督表在 `ds:B77C`（線性 `0x602AC`），一筆 `0x1C`，82 筆（`docs/internals/08`）。

候選清單走 `28E0:0103(&清單, 首都星域, 0x65, 0x10)`，`0x65` 展開成
bit0｜bit2｜bit5｜bit6（`docs/re/208`）：同陣營、`+0x00 == 0xFF`（沒有
職務）、`+0x09` bit7 有設（沒帶艦隊），加上星域參數那道「`+0x0A` ＝ 首都」。
bit6 那一支還沒讀，第四個參數是 `0x10`。

挑的規則（`2D10:0213` 的迴圈）：

    跳過 提督[+0x06] >= 0x10
    分數 ＝ 提督[+0x0E] × 提督[+0x15]；分數 < 目前最好 → 跳過
    提督[+0x0E] < 提督[+0x15] → 跳過
    否則接任

`+0x0E` ＝ 十進位 `+14` ＝ 營運；`+0x15` ＝ 十進位 `+21`，名字未知。
選之前現任會先被寫 `+0x09 ← 0x80`（`2D10:023A`），所以現任自己也在候選裡。
"""
import pathlib
import sys

ADMIRALS = 0x602AC
REC = 0x1C
N = 82
CAPITAL = {0: 0, 1: 31}     # 3982:0329 回的首都星域（docs/re/28）

mem = pathlib.Path(sys.argv[1]).read_bytes()
faction = int(sys.argv[2])
capital = int(sys.argv[3]) if len(sys.argv) > 3 else CAPITAL[faction]
# 輸入可以是 1 MB 全記憶體，也可以是只倒提督表那 82 × 0x1C 的小檔。
if len(mem) >= 0x100000:
    base = ADMIRALS
    cur = mem[0x5DA44 + 0x3B + faction]      # ds:8F4F[陣營]，WORK +0x3B
else:
    base = 0
    cur = int(sys.argv[4])                   # 現任由呼叫端給（從 WORK 讀）

rows = []
for i in range(N):
    r = mem[base + i * REC: base + (i + 1) * REC]
    rows.append(dict(idx=i, job=r[0], fac=r[1], lvl=r[6], region=r[0x0A],
                     fleet=r[9], ops=r[0x0E], f21=r[0x15]))

print("陣營 %d，首都星域 %d，現任內政參謀 ＝ 提督 %d" % (faction, capital, cur))
rows[cur]["fleet"] = 0x80                # 2D10:023A 先把現任解成「沒帶艦隊」

cand = [a for a in rows
        if a["fac"] == faction and a["job"] == 0xFF
        and a["fleet"] & 0x80 and a["region"] == capital]
print("候選 %d 名（同陣營＋沒職務＋沒帶艦隊＋在首都）" % len(cand))

best, best_score = cur, 0
print("  idx  等級  營運  +0x15   分數  結果")
for a in sorted(cand, key=lambda a: -(a["ops"] * a["f21"])):
    score = a["ops"] * a["f21"]
    why = []
    if a["lvl"] >= 0x10:
        why.append("等級擋")
    if a["ops"] < a["f21"]:
        why.append("營運<+0x15")
    print("  %3d  %4d  %4d  %5d  %5d  %s"
          % (a["idx"], a["lvl"], a["ops"], a["f21"], score,
             "／".join(why) if why else "合格"))
for a in cand:                            # 照原版的清單順序，先中先贏
    if a["lvl"] >= 0x10 or a["ops"] < a["f21"]:
        continue
    if a["ops"] * a["f21"] > best_score:
        best, best_score = a["idx"], a["ops"] * a["f21"]
print("→ 規則選出：提督 %d（分數 %d）" % (best, best_score))

# 第五個參數給「指派之後」的提督表時，逐筆列出差異——就任那一步
# （`28E0:05CC`）會改哪些欄位是實作要照抄的。
if len(sys.argv) > 5:
    post = pathlib.Path(sys.argv[5]).read_bytes()
    print("\n== 指派前後的提督表差異")
    for i in range(N):
        a = mem[base + i * REC: base + (i + 1) * REC]
        b = post[i * REC: (i + 1) * REC]
        if a != b:
            diff = [(j, a[j], b[j]) for j in range(REC) if a[j] != b[j]]
            print("  提督 %3d: %s" % (i, "  ".join(
                "+0x%02X %02X→%02X" % d for d in diff)))
