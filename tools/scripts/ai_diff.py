"""比兩份記憶體 dump 的艦隊表與星域記錄——用來看「電腦這幾回合做了什麼」。

    tools/py.sh tools/scripts/ai_diff.py <前.bin> <後.bin>

位址（`docs/internals/33`、`docs/re/22`）：
  艦隊表    linear 0x60BA4，32 筆 × 122，`+0x01` 陣營、`+0x04` 0xFF ＝ 沒在用、
            `+0x06` 所在星域
  星域記錄  linear 0x5F2CC，32 筆 × 126，`+0x12` 擁有者、`+0x1E` 成長、
            `+0x26` 稅、`+0x28`–`+0x2B` 艦隊數、`+0x7C` 民心、`+0x7D` 未安定度

輸出只列**變了的**：艦隊逐欄相減、星域逐欄相減。沒變的欄位是身分或
劇本資料，不是這幾回合的決策。
"""
import pathlib
import sys

FLEET, FLEET_N, FLEET_SZ = 0x60BA4, 32, 122
REGION, REGION_N, REGION_SZ = 0x5F2CC, 32, 126

a = pathlib.Path(sys.argv[1]).read_bytes()
b = pathlib.Path(sys.argv[2]).read_bytes()


def rec(buf, base, i, size):
    return buf[base + i * size: base + (i + 1) * size]


print('== 艦隊表 ==')
for i in range(FLEET_N):
    ra, rb = rec(a, FLEET, i, FLEET_SZ), rec(b, FLEET, i, FLEET_SZ)
    live_a, live_b = ra[4] != 0xFF, rb[4] != 0xFF
    if not live_a and not live_b:
        continue
    if live_a != live_b:
        print('  第 %2d 筆：%s' % (i, '新建' if live_b else '消失'))
        continue
    diff = [(o, ra[o], rb[o]) for o in range(FLEET_SZ) if ra[o] != rb[o]]
    if not diff:
        continue
    tag = '帝國' if rb[1] == 0 else '同盟'
    head = '  第 %2d 筆（%s，星域 %d→%d）' % (i, tag, ra[6], rb[6])
    print(head)
    print('    ' + ' '.join('+%02X %d→%d' % d for d in diff))

print('== 星域記錄 ==')
NAMES = {0x12: '擁有者', 0x1E: '成長', 0x26: '稅', 0x28: '帝國艦隊',
         0x29: '同盟艦隊', 0x2A: '?2A', 0x2B: '看得到的敵艦隊',
         0x7C: '民心', 0x7D: '未安定度'}
for i in range(REGION_N):
    ra, rb = rec(a, REGION, i, REGION_SZ), rec(b, REGION, i, REGION_SZ)
    diff = [(o, ra[o], rb[o]) for o in range(REGION_SZ) if ra[o] != rb[o]]
    if not diff:
        continue
    print('  星域 %2d' % i)
    print('    ' + ' '.join('%s+%02X %d→%d' % (NAMES.get(o, ''), o, x, y)
                            for o, x, y in diff))
