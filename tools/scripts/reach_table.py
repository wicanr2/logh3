# -*- coding: utf-8 -*-
"""從 1 MB 執行期傾印算出原版 AI 的「從首都走得到嗎」表（`2B99:0392`）。

    tools/py.sh tools/scripts/reach_table.py <1MB傾印.bin> [...]

`2B99:0392(首都, 陣營, &旗標[32])` 對每一格填
`2B99:028F(首都, i, 陣營) < 0x800`（`docs/re/185`）；`2B99:028F` 查的是
EMS 分頁框裡那張 32×32 word 距離表，表由 `2B99:007D` ＋ `2B99:01B1`
從 `ds:533A` 邊權與星域擁有者 `+0x12` 重建（`docs/re/243`）。

對每一份傾印：

1. 讀星域擁有者與邊權表，照 `2B99:007D` ＋ `2B99:01B1` 重建四張表；
2. 看分頁框 `E0000`–`E17FF` 那一刻是不是這四張表（**逐格比**，
   中點矩陣與距離矩陣都要相同才算「頁框裡就是這一頁」）；
3. 印兩個陣營從自家首都出發的距離與可達表。

第 2 步對上時，印出來的距離就是**執行期的值**，不只是重算的值。
這支只讀傾印，不改任何東西。
"""
import sys

DS = 0x54B30                     # docs/re/247
REGION = DS + 0xA79C             # 32 × 126，擁有者 +0x12
ADJ = DS + 0x533A                # 32 × 32 邊權
SEED = DS + 0x94DC
FRAME = 0xE0000                  # EMS 分頁框（docs/re/241）
INF, PEN = 0x7FFF, 0x400
CAPITAL = (0, 31)


def build(weight, owner, faction):
    d = [[INF if weight[i][j] == 0xFF else weight[i][j] +
          (PEN if owner[i] != faction or owner[j] != faction else 0)
          for j in range(32)] for i in range(32)]
    mid = [[0xFF] * 32 for _ in range(32)]
    for i in range(32):
        d[i][i] = 0
    for k in range(32):
        for i in range(32):
            for j in range(32):
                v = (d[i][k] + d[k][j]) & 0xFFFF
                if v < d[i][j]:
                    d[i][j] = v
                    mid[i][j] = k
    return d, mid


def main():
    for path in sys.argv[1:]:
        mem = open(path, "rb").read()
        if len(mem) < FRAME + 0x1800:
            print(path, "太小，要 1 MB 傾印")
            continue
        owner = [mem[REGION + i * 126 + 0x12] for i in range(32)]
        weight = [[mem[ADJ + i * 32 + j] for j in range(32)] for i in range(32)]
        seed = int.from_bytes(mem[SEED:SEED + 4], "little")
        print("==", path)
        print("  擁有者", "".join(str(o) for o in owner), " 種子 %08X" % seed)
        for faction, dbase, mbase in ((0, 0x0000, 0x1000), (1, 0x0800, 0x1400)):
            d, mid = build(weight, owner, faction)
            fd = [[int.from_bytes(mem[FRAME + dbase + (r * 32 + c) * 2:][:2], "little")
                   for c in range(32)] for r in range(32)]
            fm = [list(mem[FRAME + mbase + r * 32:FRAME + mbase + r * 32 + 32])
                  for r in range(32)]
            same_d = sum(fd[r][c] == d[r][c] for r in range(32) for c in range(32))
            same_m = sum(fm[r][c] == mid[r][c] for r in range(32) for c in range(32))
            cap = CAPITAL[faction]
            src = "頁框（執行期）" if same_d == 1024 and same_m == 1024 else "重算"
            row = fd[cap] if src.startswith("頁框") else d[cap]
            reach = [1 if v < 0x800 else 0 for v in row]
            print("  陣營 %d：頁框距離 %4d/1024、中點 %4d/1024 與重算相同 → 下列取自%s"
                  % (faction, same_d, same_m, src))
            print("    距離(首都 %d, i)：%s" % (cap, " ".join(str(v) for v in row)))
            print("    可達：%s  （0 的格：%s）" % ("".join(map(str, reach)),
                  " ".join(str(i) for i in range(32) if not reach[i])))


if __name__ == "__main__":
    main()
