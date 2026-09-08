# -*- coding: utf-8 -*-
"""把 ds:D072 那 41 筆 × 10 bytes 印出來，並與槽位表算出來的座標對照。"""
import json, struct

BASE = 0x60000
DS = 0x54B30
tab = open('workplace/form_tab.bin', 'rb').read()
mem = open('workplace/form_mem.bin', 'rb').read()

def u16(b, o): return struct.unpack_from('<H', b, o)[0]
def s16(b, o): return struct.unpack_from('<h', b, o)[0]

slots = json.load(open('docs/internals/data/battle-formation-slots.json'))
off = slots['slots']

# 收據裡下命令的那個分隊屬於哪一支艦隊：從艦艇池找。這裡先看艦隊 28。
FT = DS + 0xC074 - BASE
rec = FT + 31 * 122
fx, fy = s16(mem, rec + 0x5C), s16(mem, rec + 0x5E)
face = mem[rec + 0x4A]
print('艦隊 31：位置 (%d,%d) 朝向 %d' % (fx, fy, face))

SIN = json.load(open('docs/internals/data/sin-table.json')) if False else None
import math
S = [round(math.sin(i / 256 * 2 * math.pi) * 1024) for i in range(256)]

def rot(x, y, a):
    c, s = S[(a + 0x40) & 0xFF], S[a]
    return (x * c - y * s + 512) >> 10, (x * s + y * c + 512) >> 10

def clamp(v, lo, hi): return lo if v < lo else hi if v > hi else v

print(' 槽   表x   表y 旗標    算x   算y  符合')
bad = 0
for i in range(41):
    o = i * 10
    tx, ty, flag = u16(tab, o), u16(tab, o + 2), tab[o + 4]
    e = off[i]
    rx, ry = rot(e['x'], e['y'], face)
    cx, cy = clamp(fx + rx, 0, 0x7FF), clamp(fy + ry, 0, 0x7F7)
    ok = (cx == tx and cy == ty)
    if not ok:
        bad += 1
    print('%3d %5d %5d %4d  %5d %5d  %s' % (i, tx, ty, flag, cx, cy, '' if ok else '✗'))
print('不符 %d／41' % bad)
print('旗標非 0 的槽：', [i for i in range(41) if tab[i * 10 + 4]])
