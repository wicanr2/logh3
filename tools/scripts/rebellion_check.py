# -*- coding: utf-8 -*-
"""重播叛亂判定（`056A:0468` 那一圈 → `2703:1789`），把每一格的結果印出來。

    tools/py.sh tools/scripts/rebellion_check.py <前傾印.bin> [後傾印.bin]

前傾印要是 **1 MB 的執行期傾印（線性位址 0 起）**，時間點在 `056A:0469`
那一圈開始之前——它要讀星域表、艦隊表、提督表、`ds:B75C` 與亂數種子。

駐軍閘門（`+0x28 ＋ 擁有者`）用的是**照艦隊表重數**的值，不是傾印裡的
`+0x28`／`+0x29`：`056A:000C` 在 `056A:0469` 的前一步呼叫 `2703:0853`
重數，而現有的前傾印都倒在它之前，那兩格還是艦隊移動之前的舊值
（`docs/re/297`；收據 `rbx297w` 量到 20／30 兩格因此不擲亂數）。
後傾印可以是 1 MB 傾印，也可以是 4,032 bytes 的星域表切片；給了就逐格比
`+0x13` 的 bit1。

規則出處 `docs/re/292`。這支只讀傾印，不改任何東西。
"""
import sys

DS = 0x54B30                     # docs/re/247 從 `已被敵軍占領` 反推
REGION = DS + 0xA79C             # 32 × 126
FLEET = DS + 0xC074              # 32 × 122
B75C = DS + 0xB75C               # 32 bytes：載入當下的擁有者快照
ADMIRAL = DS + 0xB77C            # 82 × 28
STAFF = DS + 0x8F4D              # ds:8F4D[陣營] ＝ 情報專任參謀
SEED = DS + 0x94DC               # 32 位元亂數種子
RSIZE, FSIZE, ASIZE = 126, 122, 28


class Rng:
    """`3CD5:000C`（docs/re/87）。"""

    def __init__(self, seed):
        self.s = seed & 0xFFFFFFFF
        self.n = 0

    def next(self):
        self.s = (self.s * 0x41C64E6D + 0x3039) & 0xFFFFFFFF
        self.n += 1
        return (self.s >> 16) & 0x7FFF


def threshold_from_unrest(u):
    """`2703:1789` +0x1862 起的四段階梯。"""
    if u < 50:
        return u // 3
    if u < 100:
        return u // 2
    if u < 150:
        return (u * 3) // 4
    return (u * 4) // 3


def main():
    mem = open(sys.argv[1], "rb").read()
    if len(mem) < REGION + 32 * RSIZE:
        raise SystemExit("前傾印太小，要 1 MB 的執行期傾印")

    def region(i):
        o = REGION + i * RSIZE
        return mem[o:o + RSIZE]

    def fleet(i):
        o = FLEET + i * FSIZE
        return mem[o:o + FSIZE]

    def admiral(i):
        o = ADMIRAL + i * ASIZE
        return mem[o:o + ASIZE]

    staff = [mem[STAFF], mem[STAFF + 1]]
    seed = int.from_bytes(mem[SEED:SEED + 4], "little")
    rng = Rng(seed)

    # 2703:09DD：相對於 ds:B75C 那份快照，各陣營有幾格已經不是自己的
    lost = [0, 0]
    for i in range(32):
        own = region(i)[0x12]
        old = mem[B75C + i]
        if own == 2 or old == 2 or old == own:
            continue
        lost[old] += 1

    # 2703:0853：清掉 32 格的 +0x28／+0x29，再對每支在用的艦隊（+0x04 != FF）
    # 在 +0x28 ＋ 陣營 加一。不看 +0x00 bit0。
    garrison = [[0, 0] for _ in range(32)]
    for s in range(32):
        f = fleet(s)
        if f[0x04] == 0xFF or f[0x06] >= 32 or f[0x01] > 1:
            continue
        garrison[f[0x06]][f[0x01]] += 1

    print("ds:8F4D 情報專任參謀 = %d / %d（營運 +0x0E = %d / %d）"
          % (staff[0], staff[1],
             admiral(staff[0])[0x0E], admiral(staff[1])[0x0E]))
    print("亂數種子 ds:94DC = %08X" % seed)
    print("ds:B75C 快照 vs 現況：帝國已失 %d 格、同盟已失 %d 格" % (lost[0], lost[1]))
    print()
    hdr = ("格  擁  民心 未安  駐軍(重數)  在場艦隊 帝/同  1789?  門檻  r1+r2  "
           "bit1  舊+13")
    print(hdr)

    pred = {}
    for i in range(32):
        r = region(i)
        own = r[0x12]
        imp = ally = 0
        for s in range(32):
            f = fleet(s)
            if f[0x04] == 0xFF or f[0x06] != i:
                continue
            if f[0x00] & 1:
                continue
            if f[0x01] == 0:
                imp += 1
            else:
                ally += 1

        called = not (imp and ally) and not (imp and own != 0) and not (ally and own != 1)
        bit1, thr, roll = 0, "", ""
        if called:
            if i in (0, 31) or own == 2 or garrison[i][own] != 0:
                thr = "—"
            else:
                r1 = rng.next() % 400
                for _ in range(4):
                    rng.next()
                r2 = rng.next() % 400
                for _ in range(4):
                    rng.next()
                t = threshold_from_unrest(r[0x7D])
                op = admiral(staff[own])[0x0E]
                t = (t * 100) // op if op else 0xFFFF
                t //= lost[own] + 1
                thr, roll = str(t), str(r1 + r2)
                bit1 = 1 if r1 + r2 < t else 0
        pred[i] = bit1
        print("%2d  %2d  %4d %4d   %3d/%-3d      %2d / %-2d       %-5s %5s %6s   %d     %02X"
              % (i, own, r[0x7C], r[0x7D], garrison[i][0], garrison[i][1], imp, ally,
                 "是" if called else "否", thr, roll, bit1, r[0x13]))

    print()
    print("整圈擲了 %d 次亂數，之後的種子 = %08X" % (rng.n, rng.s))

    if len(sys.argv) > 2:
        post = open(sys.argv[2], "rb").read()
        base = 0 if len(post) == 32 * RSIZE else REGION
        print()
        bad = 0
        for i in range(32):
            got = (post[base + i * RSIZE + 0x13] >> 1) & 1
            if got != pred[i]:
                bad += 1
                print("✗ 星域 %2d：預測 bit1=%d，收據 bit1=%d" % (i, pred[i], got))
        print("逐格比對：32 格中 %d 格不符" % bad)
        if not bad:
            print("全部相同（含 %d 格叛亂）" % sum(pred.values()))


main()
