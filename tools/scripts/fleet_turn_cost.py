# -*- coding: utf-8 -*-
"""把 s3fltseq 十一份傾印裡第 28 艦隊的轉向相關欄位排出來。

傾印是 lin:60000:65536，艦隊表在 ds:C074（ds 基底 0x54B30）→ 檔內位移 0xBA4。
"""
import sys, os, struct

BASE = 0x60000
FLEET_TAB = 0x54B30 + 0xC074
OFF = FLEET_TAB - BASE          # 0xBA4
REC = 122

def u8(b, o):  return b[o]
def u16(b, o): return struct.unpack_from('<H', b, o)[0]
def s16(b, o): return struct.unpack_from('<h', b, o)[0]

def fx(lo, hi):
    """+0x52/+0x54 這種 16.16：低字在前一格、高字在後一格。"""
    v = (hi << 16) | lo
    if v & 0x80000000:
        v -= 1 << 32
    return v

idx = int(sys.argv[1]) if len(sys.argv) > 1 else 28
files = sorted(f for f in os.listdir('workplace') if f.startswith('ft') and f.endswith('.bin'))

print('檔      行動 朝向 目標朝向 回轉 巡航 戰鬥  每回合      剩餘        位置        目的地')
for fn in files:
    b = open(os.path.join('workplace', fn), 'rb').read()
    r = OFF + idx * REC
    act   = u8(b, r + 0x48) & 0x0F
    face  = u16(b, r + 0x4A)
    tface = u16(b, r + 0x4C)
    turn  = u8(b, r + 0x4D)
    cru   = u8(b, r + 0x4E)
    fight = u8(b, r + 0x4F)
    per   = fx(u16(b, r + 0x52), u16(b, r + 0x54))
    left  = fx(u16(b, r + 0x56), u16(b, r + 0x58))
    x, y  = s16(b, r + 0x5C), s16(b, r + 0x5E)
    dx, dy = s16(b, r + 0x64), s16(b, r + 0x66)
    print('%-6s %3d %4d %7d %5d %4d %5d %9.4f %9.4f  (%4d,%4d) (%4d,%4d)'
          % (fn, act, face, tface, turn, cru, fight,
             per / 65536.0, left / 65536.0, x, y, dx, dy))
