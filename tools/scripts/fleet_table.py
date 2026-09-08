"""把一份記憶體 dump 裡的執行期艦隊表列出來（`docs/internals/33`）。

    tools/py.sh tools/scripts/fleet_table.py <dump.bin> [星域記錄也印]

`ds:C074` ＝ linear 0x60BA4，32 筆 × 122。`+0x04` 是 0xFF 就是沒在用。
"""
import pathlib
import sys

FLEET, N, SZ = 0x60BA4, 32, 122
REGION, RN, RSZ = 0x5F2CC, 32, 126
mem = pathlib.Path(sys.argv[1]).read_bytes()

print('筆 陣營 星域  +00 +02 +07 +0A +48 +4A +4B +4C')
live = 0
for i in range(N):
    r = mem[FLEET + i*SZ: FLEET + (i+1)*SZ]
    if r[4] == 0xFF:
        continue
    live += 1
    print('%2d  %s  %2d   %3d %3d %3d %3d %3d %3d %3d %3d' % (
        i, '帝國' if r[1] == 0 else '同盟', r[6],
        r[0], r[2], r[7], r[0x0A], r[0x48], r[0x4A], r[0x4B], r[0x4C]))
print('在用 %d 筆' % live)

if len(sys.argv) > 2:
    print('\n星域 擁有者 帝國艦隊 同盟艦隊 +2A 看得到的敵艦隊 民心 未安定度')
    for i in range(RN):
        r = mem[REGION + i*RSZ: REGION + (i+1)*RSZ]
        print('%3d %5d %6d %8d %5d %8d %6d %6d' % (
            i, r[0x12], r[0x28], r[0x29], r[0x2A], r[0x2B], r[0x7C], r[0x7D]))
