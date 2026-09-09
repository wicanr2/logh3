"""找「打到某個線性位址」的 far call，段值不必事先知道。

    farcall_to.py <傾印> <起始線性位址hex> <目標線性位址hex> [...]

`9A off16 seg16` 的目標是 `seg*16+off`，同一個線性位址有很多組
(seg, off) 寫法，所以逐一枚舉每個合法的段值去掃。
"""
import sys

dump = open(sys.argv[1], "rb").read()
base = int(sys.argv[2], 16)
for t in sys.argv[3:]:
    tgt = int(t, 16)
    print("== 目標 lin %05X ==" % tgt)
    n = 0
    lo = max(0, (tgt - 0xFFFF + 15) // 16)
    for seg in range(lo, tgt // 16 + 1):
        off = tgt - seg * 16
        pat = bytes([0x9A, off & 0xFF, off >> 8, seg & 0xFF, seg >> 8])
        i = 0
        while True:
            i = dump.find(pat, i)
            if i < 0:
                break
            print("    lin %05X → %04X:%04X" % (base + i, seg, off))
            n += 1
            i += 1
    print("    共 %d 處" % n)
