"""在記憶體 dump 上找「呼叫某個線性位址」的地方，**不管它用哪一組段:位移**。

    tools/py.sh tools/scripts/xref_linear.py <dump> <線性hex> [容差]

真實模式裡同一個線性位址有很多種段:位移寫法，所以照單一段值去比
`call far` 的簽章會漏掉（`callers.py xref` 就是這樣）。這一支把每一個
`9A off seg` 的目標算成線性位址再比，近程 `E8 rel16` 也一起掃
（同段內的呼叫，段值取「呼叫點所在的 64K 對齊段」是不夠的，
所以近程只在目標與呼叫點相距 64K 以內時才算數）。
"""
import pathlib
import sys

mem = pathlib.Path(sys.argv[1]).read_bytes()
target = int(sys.argv[2], 16)
tol = int(sys.argv[3]) if len(sys.argv) > 3 else 0

far, near = [], []
for i in range(len(mem) - 5):
    if mem[i] == 0x9A:
        off = int.from_bytes(mem[i+1:i+3], 'little')
        seg = int.from_bytes(mem[i+3:i+5], 'little')
        lin = seg * 16 + off
        if abs(lin - target) <= tol:
            far.append((i, seg, off, lin))
    elif mem[i] == 0xE8:
        rel = int.from_bytes(mem[i+1:i+3], 'little', signed=True)
        lin = i + 3 + rel
        if abs(lin - target) <= tol and abs(rel) < 0x8000:
            near.append((i, lin))

print('call far 到 %05X（容差 %d）：%d 處' % (target, tol, len(far)))
for i, seg, off, lin in far[:25]:
    print('  線性 %05X  ← 9A %04X:%04X（%05X）' % (i, seg, off, lin))
print('call near：%d 處' % len(near))
for i, lin in near[:25]:
    print('  線性 %05X → %05X' % (i, lin))
