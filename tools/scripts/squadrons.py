"""艦隊在戰場上散成的分隊：位置、佈陣與 78 位元組記錄的欄位。

    tools/py.sh tools/scripts/squadrons.py <記憶體dump> [艦隊索引...]

艦隊記錄 `+0x10` 起十組，第一個 word 是艦艇池索引；池的位置從
`ds:CFB4` 的 far 指標拿，一筆 78 位元組，`+0x28` 是戰場座標
（`docs/internals/38`）。不給艦隊索引就把在場的全部列出來。
"""
import pathlib
import struct
import sys

FLEET, STRIDE, REC = 0x60BA4, 122, 78
DS = FLEET - 0xC074                    # ds 段基底
POOL_PTR = 0xCFB4                      # ds:CFB4 far ptr → 艦艇池

mem = pathlib.Path(sys.argv[1]).read_bytes()
off, seg = struct.unpack_from('<HH', mem, DS + POOL_PTR)
base = seg * 16 + off
print('艦艇池 %04X:%04X → 線性 %05X' % (seg, off, base))

want = [int(v) for v in sys.argv[2:]]
for i in range(32):
    r = FLEET + i * STRIDE
    if mem[r + 4] == 0xFF or (want and i not in want):
        continue
    idxs = [struct.unpack_from('<H', mem, r + 0x10 + g * 4)[0] for g in range(10)]
    idxs = [v for v in idxs if v != 0xFFFF]
    if not idxs:
        continue
    pos = [struct.unpack_from('<hh', mem, base + v * REC + 0x28) for v in idxs]
    tgt = [struct.unpack_from('<hh', mem, base + v * REC + 0x34) for v in idxs]
    ox, oy = pos[0]
    u = (pos[1][0] - ox, pos[1][1] - oy) if len(pos) > 1 else (0, 0)
    v = (pos[2][0] - ox, pos[2][1] - oy) if len(pos) > 2 else (0, 0)
    print('艦隊 %2d 陣營 %d 編號 %2d 組數 %d 旗艦 (%d,%d) 朝向 %s'
          % (i, mem[r + 1], mem[r + 2], mem[r + 0x0F], ox, oy, u))
    for k, n in enumerate(idxs):
        a = base + n * REC
        dx, dy = pos[k][0] - ox, pos[k][1] - oy
        best = min(((abs(p * u[0] + q * v[0] - dx) + abs(p * u[1] + q * v[1] - dy), p, q)
                    for p in range(4) for q in range(4)), default=(0, 0, 0))
        print('   組%d 池%4d %05X 艦種 %d 艘數 %5d/%-5d 位置 (%5d,%5d) 目標 (%5d,%5d)'
              '  位移 (%4d,%4d) ≈ %du%+dv 誤差 %d'
              % (k, n, a, mem[a + 4],
                 struct.unpack_from('<H', mem, a + 8)[0],
                 struct.unpack_from('<H', mem, a + 0x0A)[0],
                 pos[k][0], pos[k][1], tgt[k][0], tgt[k][1],
                 dx, dy, best[1], best[2], best[0]))
