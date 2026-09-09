"""驗 `090A:0B93`（Blend 一動就整批重算的稅率）的算式。

    tools/py.sh tools/scripts/taxrate_check.py <傾印> <等級> [陣營] [統治者編號]

拿傾印裡的星域表與統治者記錄，照反組譯讀出來的算式重算每個
「擁有者 ＝ ds:8F1D 的陣營」的稅率，和傾印裡的 `+0x26` 逐格比。
"""
import sys

DS = 0x54B30
REGION = DS + 0xA79C          # ds:A79C
ADMIRAL = DS + 0xB77C         # ds:B77C
BASE = DS + 0x147E            # ds:147E，32 位元 × 5
RULER = DS + 0x8F4F           # ds:8F4F[陣營]
FACTION = DS + 0x8F1D         # ds:8F1D

mem = open(sys.argv[1], "rb").read()
level = int(sys.argv[2])
faction = int(sys.argv[3]) if len(sys.argv) > 3 else mem[FACTION]

u16 = lambda a: mem[a] | mem[a + 1] << 8
s16 = lambda a: u16(a) - 0x10000 if u16(a) & 0x8000 else u16(a)
u32 = lambda a: u16(a) | u16(a + 2) << 16

base = [u32(BASE + 4 * i) for i in range(5)]
rulerno = int(sys.argv[4]) if len(sys.argv) > 4 else mem[RULER + faction]
ruler = ADMIRAL + rulerno * 28
r0e, r14, r15 = mem[ruler + 0x0E], mem[ruler + 0x14], mem[ruler + 0x15]
print("陣營 %d 統治者 #%d：+0x0E=%d +0x14=%d +0x15=%d；Base=%s 等級=%d"
      % (faction, rulerno, r0e, r14, r15, base, level))


def idiv(a, b):
    """x86 的 idiv：往零截斷。"""
    q = abs(a) // abs(b)
    return -q if (a < 0) != (b < 0) else q


ok = bad = 0
for i in range(32):
    rec = REGION + i * 126
    if mem[rec + 0x12] != faction:
        continue
    growth = s16(rec + 0x1E)
    if growth <= 0:
        v = idiv(base[level] * 5, 100)
    else:
        v = idiv(growth * base[level], 100)
        v = idiv(v * 100, r0e)
        v = idiv(v * r15, 100)
        loyal = mem[rec + 0x7C]
        m = idiv((loyal if faction == 0 else 200 - loyal) * r14, 100)
        if m > 150:
            v = idiv(v * 30, r14)
        elif m > 100:
            v = idiv(v * 50, r14)
        elif m > 50:
            v = idiv(v * 75, r14)
    v = 200 if v > 200 else (1 if v <= 0 else v)
    got = mem[rec + 0x26]
    flag = "" if v == got else "  ← 不合"
    if v == got:
        ok += 1
    else:
        bad += 1
    print("星域 %2d：成長率 %5d 民心 %3d → 算 %3d／傾印 %3d%s"
          % (i, growth, mem[rec + 0x7C], v, got, flag))
print("合 %d／不合 %d" % (ok, bad))
