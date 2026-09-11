# -*- coding: utf-8 -*-
"""叛亂的效果（`2703:1908` → `2703:1ADB`）：逐顆行星對前後兩份傾印。

    tools/py.sh tools/scripts/rebellion_effect_check.py <前.bin> <後.bin> [叛亂星域,逗號分隔]

兩個檔都是 **81 × 82 bytes 的行星表切片**（線性 `0x53132` 起，`docs/internals/29`）。
給了星域清單就一併檢查「這一格該有幾顆被翻、翻的是哪一顆」。

規則出處 `docs/re/292`。
"""
import sys

PSIZE = 82
COUNTS = [4, 3, 2, 3, 1, 2, 3, 3, 4, 3, 2, 2, 1, 1, 1, 2,
          3, 2, 3, 3, 3, 4, 2, 2, 1, 3, 3, 2, 3, 2, 3, 5]


def u16(b, o):
    return int.from_bytes(b[o:o + 2], "little")


def u32(b, o):
    return int.from_bytes(b[o:o + 4], "little")


def main():
    pre = open(sys.argv[1], "rb").read()
    post = open(sys.argv[2], "rb").read()
    want = None
    if len(sys.argv) > 3:
        want = [int(v) for v in sys.argv[3].split(",")]

    first = []
    acc = 0
    for n in COUNTS:
        first.append(acc)
        acc += n
    region_of = {}
    for r, n in enumerate(COUNTS):
        for k in range(n):
            region_of[first[r] + k] = (r, k)

    changed = {}
    other = 0
    print("行星 星域 切片序  擁有者    經濟力            防衛力/上限")
    for i in range(81):
        a, b = pre[i * PSIZE:(i + 1) * PSIZE], post[i * PSIZE:(i + 1) * PSIZE]
        if a[0x11] == b[0x11]:
            # 同一個回合裡別的常式也會動行星（防衛力回復 2703:1CD7、要塞砲
            # 歸零 2703:1D85）。只有「擁有者變了」才是叛亂翻過去的那一顆。
            other += a != b
            continue
        r, k = region_of[i]
        changed.setdefault(r, []).append(k)
        eco_a, eco_b = u32(a, 0x1A), u32(b, 0x1A)
        def_a, def_b = u16(a, 0x14), u16(b, 0x14)
        cap = u16(a, 0x16)
        want_eco = eco_a - eco_a // 3
        want_def = 0 if def_a * 3 < cap else def_a - cap // 3
        ok = ("✓" if (eco_b == want_eco and def_b == want_def and b[0x11] == 2)
              else "✗ 期望 經濟力 %d、防衛力 %d、擁有者 2" % (want_eco, want_def))
        print("%3d   %2d    %d/%d    %d→%d  %8d→%-8d  %5d→%-5d /%-5d %s"
              % (i, r, k, COUNTS[r] - 1, a[0x11], b[0x11],
                 eco_a, eco_b, def_a, def_b, cap, ok))

    print("（另有 %d 顆行星的其他欄位有變動，不是擁有者——防衛力回復與要塞砲歸零）" % other)

    if want is not None:
        print()
        bad = 0
        for r in want:
            ks = changed.get(r, [])
            last = COUNTS[r] - 1
            if ks != [last]:
                bad += 1
                print("✗ 星域 %2d：期望只翻切片最後一顆（序 %d），實際 %s" % (r, last, ks))
        extra = sorted(set(changed) - set(want))
        if extra:
            bad += 1
            print("✗ 清單外還有星域被改：%s" % (extra,))
        print("叛亂星域 %d 格：%d 格不符" % (len(want), bad))


main()
