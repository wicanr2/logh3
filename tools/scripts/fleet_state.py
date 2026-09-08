"""列一份 dump 裡每一支艦隊的「所在／目的地／動不動」（`docs/internals/33`）。

    tools/py.sh tools/scripts/fleet_state.py <dump.bin> [更多 dump...]

`+0x08` 是 5 就不會動——問「電腦有沒有在下命令」之前，先確認那個局面
裡敵方艦隊是不是根本被設成不動（`docs/re/77`）。
"""
import pathlib
import sys

FLEET, N, SZ = 0x60BA4, 32, 122
for p in sys.argv[1:]:
    mem = pathlib.Path(p).read_bytes()
    print('==', p)
    for i in range(N):
        r = mem[FLEET + i*SZ: FLEET + (i+1)*SZ]
        if r[4] == 0xFF:
            continue
        print('  %2d %s 星域%2d 目的%3d +08=%d%s'
              % (i, '帝' if r[1] == 0 else '同', r[6], r[7], r[8],
                 '  ← 會動' if r[8] != 5 else ''))
