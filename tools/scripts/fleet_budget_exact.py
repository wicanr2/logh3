# -*- coding: utf-8 -*-
"""第 28 艦隊那一回合的行程額度，逐步對到位元。

模型（`1C03:01B5` 逐行）：一次呼叫 ＝ 轉一步 ＋ 前進一步。
  轉：差 = 2BE1:021B(方位, 朝向)；|差| ≤ +0x4D 就一次轉到位且不扣，
      否則額度 < 8 就不轉，夠就轉 +0x4D 並扣 8.0。
  前進：額度 < 8 就結束；否則沿**轉完之後**的朝向走 8，扣實際距離。
"""
import os, struct, math

BASE = 0x60000
OFF = 0x54B30 + 0xC074 - BASE
REC, IDX = 122, 28
ONE = 65536

def u16(b, o): return struct.unpack_from('<H', b, o)[0]
def u32(b, o): return (u16(b, o + 2) << 16) | u16(b, o)

SIN = [round(math.sin(i / 256 * 2 * math.pi) * 1024) for i in range(256)]

def step8(a):
    """2BE1:026F，距離 8。"""
    return (SIN[(a + 0x40) & 0xFF] * 8 + 512) >> 10, (SIN[a] * 8 + 512) >> 10

def dist(dx, dy):
    """3310:0006 ＝ M × √(1 + (m/M)²)，全程 16.16。"""
    M, m = max(abs(dx), abs(dy)), min(abs(dx), abs(dy))
    if M == 0:
        return 0
    q = (m << 16) // M                      # 39B5:005C，截斷
    r = ONE + ((q * q) >> 16)               # 39B5:000C，截斷
    s = round(math.sqrt(r / ONE) * ONE)     # 3310:014B，捨去方式未讀，四捨五入與兩處都相符
    return M * s

def diff256(target, cur):
    d = (target - cur) & 0xFF
    return d - 256 if d >= 128 else d

snaps = []
for fn in sorted(f for f in os.listdir('workplace') if f.startswith('ft') and f.endswith('.bin')):
    b = open(os.path.join('workplace', fn), 'rb').read()
    r = OFF + IDX * REC
    snaps.append((fn, b[r + 0x4A], b[r + 0x4D], u32(b, r + 0x52), u32(b, r + 0x56)))

for fn, face, turn, per, left in snaps:
    print('%-8s 朝向 %3d 回轉 %2d 每回合 %8d 剩餘 %8d (%.6f)' % (fn, face, turn, per, left, left / ONE))

face0, turnspd = 190, snaps[0][2]
x, y, dx0, dy0 = 1330, 941, 1240, 840
bearing = 163                     # 2BE1:011D(目的地 − 現在)，(−90,−101) → 163
budget = 160 * ONE
face = face0
log = []
calls = 0
while True:
    d = diff256(bearing, face)
    if d:
        if abs(d) <= turnspd:
            face = (face + d) & 0xFF
            log.append('轉到位 → %d（不扣）' % face)
        elif budget >= 8 * ONE:
            face = (face + (1 if d > 0 else -1) * turnspd) & 0xFF
            budget -= 8 * ONE
            log.append('轉一步 → %d，扣 8.000000（餘 %d）' % (face, budget))
        else:
            break
    if budget < 8 * ONE:
        break
    sx, sy = step8(face)
    c = dist(sx, sy) or 1
    budget -= c
    calls += 1
    log.append('前進 %2d：朝向 %3d 位移 (%d,%d) 扣 %d（%.6f）→ 餘 %d'
               % (calls, face, sx, sy, c, c / ONE, budget))

for l in log:
    print(l)
print()
print('算出來剩 %d（%.6f）　收據 %d（%.6f）　差 %d'
      % (budget, budget / ONE, snaps[-1][4], snaps[-1][4] / ONE, budget - snaps[-1][4]))

# 中途那一張（ft03）落在哪一步
mid = snaps[3][4]
b2 = 160 * ONE - 8 * ONE - dist(*step8(175))
for k in range(0, 25):
    if b2 - k * dist(*step8(163)) == mid:
        print('ft03 的 %d ＝ 轉一步 ＋ 朝向 175 走 1 步 ＋ 朝向 163 走 %d 步' % (mid, k))
