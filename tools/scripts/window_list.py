"""走訪視窗串列，把每個視窗的類型、原點、大小與命中範圍列出來。

串列頭是 far 指標，存在 `ds:D804`（位移）／`ds:D806`（段）；下一個在
視窗結構的 `+0x16`（位移）／`+0x18`（段）。命中判斷（`341B:039D`）是

    +0 == 2 且 格x ∈ [+5, +5++7) 且 格y ∈ [+6, +6++8)

回傳第一個命中的視窗——**串列順序就是 z-order**，排前面的先贏。
所以「某個熱區點不到」有兩種可能：那個視窗不在串列裡，
或它排在另一個蓋住同一塊的視窗後面。

用法：tools/py.sh tools/scripts/window_list.py <dump> [DS 段 hex]
"""
import pathlib, sys

mem = pathlib.Path(sys.argv[1]).read_bytes()
DS = int(sys.argv[2], 16) if len(sys.argv) > 2 else 0x54B3
ds = DS * 16

def w16(a):
    return mem[a] | mem[a+1] << 8

head_off, head_seg = w16(ds + 0xD804), w16(ds + 0xD806)
print('串列頭 %04X:%04X' % (head_seg, head_off))

seg, off, seen = head_seg, head_off, set()
n = 0
while (seg or off) and (seg, off) not in seen:
    seen.add((seg, off))
    a = seg * 16 + off
    if a + 32 > len(mem):
        print('  %04X:%04X 超出 dump 範圍' % (seg, off)); break
    t = mem[a]
    x, y, w, h = mem[a+5], mem[a+6], mem[a+7], mem[a+8]
    nxt_off, nxt_seg = w16(a + 0x16), w16(a + 0x18)
    px = (x * 16, (y + 2) * 20, w * 16, h * 20)
    hit = (x * 16, y * 20, (x + w) * 16, (y + h) * 20)
    print('%2d. %04X:%04X（線性 %05X）type=%02X 格(%d,%d) %dx%d'
          % (n, seg, off, a, t, x, y, w, h))
    print('     視窗像素 (%d,%d) %dx%d；命中範圍 x %d–%d, y %d–%d%s'
          % (*px, hit[0], hit[2] - 1, hit[1], hit[3] - 1,
             '' if t == 2 else '  ← type≠2，命中測試會跳過'))
    print('     next %04X:%04X' % (nxt_seg, nxt_off))
    seg, off = nxt_seg, nxt_off
    n += 1
print('共 %d 個視窗' % n)
