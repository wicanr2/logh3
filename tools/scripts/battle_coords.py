"""列戰鬥中每一支艦隊的座標與目標點（`docs/internals/34`）。

    tools/py.sh tools/scripts/battle_coords.py <dump.bin> [另一份.bin]

`+0x5C`／`+0x5E` 是現在的位置，`+0x64`／`+0x66` 是要去的點。
給兩份就並排比，看這幾個回合誰動了、往哪裡動。
"""
import pathlib
import sys

FLEET, N, SZ = 0x60BA4, 32, 122


def read(p):
    mem = pathlib.Path(p).read_bytes()
    out = {}
    for i in range(N):
        r = mem[FLEET + i*SZ: FLEET + (i+1)*SZ]
        if r[4] == 0xFF:
            continue
        w = lambda o: r[o] | r[o+1] << 8
        out[i] = (r[1], r[6], w(0x5C), w(0x5E), w(0x64), w(0x66))
    return out


a = read(sys.argv[1])
b = read(sys.argv[2]) if len(sys.argv) > 2 else None
for i, v in sorted(a.items()):
    tag = '帝' if v[0] == 0 else '同'
    line = '%2d %s 星域%2d  位置(%5d,%5d)  目標(%5d,%5d)' % (i, tag, v[1], v[2], v[3], v[4], v[5])
    if b and i in b:
        u = b[i]
        line += '   →  位置(%5d,%5d)  目標(%5d,%5d)%s' % (
            u[2], u[3], u[4], u[5], '' if (u[2], u[3]) == (v[2], v[3]) else '  ← 動了')
    print(line)
