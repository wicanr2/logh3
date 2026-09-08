"""看戰力算式（`2E54:1876`，`docs/re/79`）用到的那三個欄位與那張小表。

    tools/py.sh tools/scripts/ship_strength.py <SHIP.DT?> <dump.bin>

艦艇池一筆 78 bytes（`docs/re/42`）。算式是

    ds:6032[艦艇 +0x04] × 艦艇 +0x08 × 1000 ÷ 艦艇 +0x0A
"""
import collections
import pathlib
import sys

ship = pathlib.Path(sys.argv[1]).read_bytes()
mem = pathlib.Path(sys.argv[2]).read_bytes()
DS = 0x54B3 * 16
SZ = 78

kinds = collections.Counter()
for i in range(len(ship) // SZ):
    r = ship[i*SZ:(i+1)*SZ]
    if all(b == 0 for b in r[:0x38]):
        continue
    kinds[(r[4], int.from_bytes(r[8:10], 'little'), int.from_bytes(r[0x0A:0x0C], 'little'))] += 1

print('（+0x04, +0x08, +0x0A）→ 幾筆')
for k, n in sorted(kinds.items()):
    print('  %-18s %4d' % (str(k), n))

print('\nds:6032 前 32 個 byte')
print('  ' + ' '.join('%02X' % b for b in mem[DS+0x6032: DS+0x6032+32]))
