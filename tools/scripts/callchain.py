"""從一個位址往上追呼叫鏈：函式起點 → 誰呼叫它 → 再往上。

    tools/py.sh tools/scripts/callchain.py <dump> <線性hex> [層數]

段值在真實模式下有很多寫法，所以比對一律用**線性位址**
（同 `xref_linear.py`）。函式起點的判法是往前找最近的 `retf`／`ret`，
其後第一個 `push bp` 家族——**這是啟發式，不是證據**，印出來的位元組
要自己看一眼確認是序幕。
"""
import pathlib
import sys

mem = pathlib.Path(sys.argv[1]).read_bytes()
addr = int(sys.argv[2], 16)
depth = int(sys.argv[3]) if len(sys.argv) > 3 else 6

PROLOG = (b'\x55\x83', b'\x55\x89', b'\x55\x8b', b'\x55\x56', b'\x55\x57')


def entry(a):
    for i in range(a, max(0, a - 0x1000), -1):
        if mem[i] in (0xCB, 0xC3):
            j = i + 1
            if mem[j:j+2] in PROLOG:
                return j
    return None


def callers(target):
    out = []
    for i in range(len(mem) - 5):
        if mem[i] != 0x9A:
            continue
        off = int.from_bytes(mem[i+1:i+3], 'little')
        seg = int.from_bytes(mem[i+3:i+5], 'little')
        if seg * 16 + off == target:
            out.append((i, seg, off))
    return out


cur = addr
for level in range(depth):
    e = entry(cur)
    if e is None:
        print('%d. %05X → 找不到函式起點' % (level, cur))
        break
    print('%d. %05X 在函式 %05X（%s …）' % (
        level, cur, e, ' '.join('%02X' % b for b in mem[e:e+8])))
    cs = callers(e)
    if not cs:
        print('   沒有 call far 到 %05X——可能是近程呼叫或間接跳' % e)
        break
    for i, seg, off in cs[:4]:
        print('   ← %05X  9A %04X:%04X' % (i, seg, off))
    cur = cs[0][0]
