# -*- coding: utf-8 -*-
"""費沙事件（docs/re/296）：拿事件前的傾印重播規則，與事件後的傾印逐位元組比。

    tools/py.sh tools/scripts/phezzan_event_check.py <前> <本體之後> <懲罰之後>

三份都是 `-dump-mem-at <步數>:lin:53132:53594:...` 倒出來的切片——從行星表
（段 5313 的 offset 2，線性 0x53132）一路到星域表結尾（0x5F2CC ＋ 0xFC0 ＝ 0x6028C），
中間的 DS 全域（ds:8F18 回合、ds:8F48、ds:8F21／8F25 座位、ds:328E 比例表）
都在裡面。

  前 → 本體之後：`1213:025D`（ds:8F48 ← 1、費沙八成經濟力搬到對手首都行星、
                 鄰接清單裡同陣營沒駐軍的星域整格叛離 `2703:1908(i, 1)`）
  本體之後 → 懲罰之後：`1213:03F5`（費沙擁有者是人類座位時，那一方其餘星域
                 每顆行星經濟力扣 ds:328E[回合÷100]%、民心 ← 100），
                 接著包裝的第三支 `090A:0E6D` → `2703:02E5` 每格重算
                 星域經濟力 `+0x1A`（＝ 行星加總，docs/internals/29）與未安定度
                 `+0x7D`（docs/re/43；本工具**不模擬**，那一格照實列出變了幾格、不算不符）

**只比預測會動的欄位不夠**：兩張表整張逐位元組比，預測以外的差異一律列出來。

切片比 53594 bytes 短時（`pzx296` 用的是 52058，星域表只到第 19 格的一部分），
星域表只比得到切片裡有的那幾格，會照實印出「比了幾格」；行星表 81 顆不受影響。
"""
import sys

BASE = 0x53132
DS = 0x54B30
REG = 0x5F2CC
CNT = [4, 3, 2, 3, 1, 2, 3, 3, 4, 3, 2, 2, 1, 1, 1, 2, 3, 2, 3, 3, 3, 4, 2, 2, 1, 3, 3, 2, 3, 2, 3, 5]
LISTS = {0: (31, 4, [10, 11, 12]),                       # 費沙歸帝國 → 巴拉特第 4 顆、ds:3280
         1: (0, 0, [15, 22, 23, 24, 25, 28, 29, 30])}   # 否則 → 瓦爾哈拉第 0 顆、ds:3284


def first(k):
    return sum(CNT[:k])


class Snap:
    def __init__(self, path):
        self.m = bytearray(open(path, 'rb').read())

    def b(self, lin):
        return self.m[lin - BASE]

    def setb(self, lin, v):
        self.m[lin - BASE] = v & 0xFF

    def w(self, lin):
        return self.m[lin - BASE] | self.m[lin - BASE + 1] << 8

    def d(self, lin):
        o = lin - BASE
        return int.from_bytes(self.m[o:o + 4], 'little')

    def setd(self, lin, v):
        o = lin - BASE
        self.m[o:o + 4] = (v & 0xFFFFFFFF).to_bytes(4, 'little')

    def setw(self, lin, v):
        o = lin - BASE
        self.m[o:o + 2] = (v & 0xFFFF).to_bytes(2, 'little')

    def has(self, lin):
        return 0 <= lin - BASE < len(self.m)

    def tables(self):
        pla = bytes(self.m[0:81 * 82])
        n = max(0, min(32, (len(self.m) - (REG - BASE)) // 126))
        reg = bytes(self.m[REG - BASE:REG - BASE + n * 126])
        return pla, reg


def P(i):
    return BASE + i * 82


def R(i):
    return REG + i * 126


def rebel_all(s, r):
    """2703:1908(r, 1)：整格行星照切片順序全部跑 2703:1ADB。"""
    s.setb(R(r) + 0x12, 2)
    for p in range(first(r), first(r) + CNT[r]):
        s.setb(P(p) + 0x11, 2)
        e = s.d(P(p) + 0x1A)
        s.setd(P(p) + 0x1A, e - e // 3)
        dfn, mx = s.w(P(p) + 0x14), s.w(P(p) + 0x16)
        s.setw(P(p) + 0x14, 0 if dfn * 3 < mx else dfn - mx // 3)


def body(s):
    """1213:025D"""
    s.setb(DS + 0x8F48, 1)
    own = s.b(R(14) + 0x12)
    tr, tp, lst = LISTS[0] if own == 0 else LISTS[1]
    e = s.d(P(first(14)) + 0x1A)
    v = e * 80 // 100
    s.setd(P(first(14)) + 0x1A, e - v)
    t = P(first(tr) + tp)
    s.setd(t + 0x1A, s.d(t + 0x1A) + v)
    hit = []
    for r in lst:
        if not s.has(R(r) + 0x7D):
            raise SystemExit('星域 %d 不在切片裡，清單那一段預測不了' % r)
        ro = s.b(R(r) + 0x12)
        if ro != own:
            continue
        if s.b(R(r) + 0x28 + ro) != 0:
            continue
        rebel_all(s, r)
        hit.append(r)
    return own, v, hit


def human(s, f):
    """3982:022D：0 ＝ 人類座位。"""
    if f == 2:
        return 0xFF
    seat = 0 if s.b(DS + 0x8F25) == f else 1
    return s.b(DS + 0x8F21 + seat)


def penalty(s):
    """1213:03F5"""
    own = s.b(R(14) + 0x12)
    if human(s, own) != 0:
        return own, None, []
    t = s.w(DS + 0x8F18)
    pct = 0 if t >= 600 else s.b(DS + 0x328E + t // 100)
    hit = []
    for i in range(32):
        if not s.has(R(i) + 0x7D):
            print('   星域 %d 不在切片裡：擁有者讀不到，這一格不預測（它的行星仍在行星表裡照比）' % i)
            continue
        if i == 14 or s.b(R(i) + 0x12) != own:
            continue
        for p in range(first(i), first(i) + CNT[i]):
            e = s.d(P(p) + 0x1A)
            s.setd(P(p) + 0x1A, e - e * pct // 100)
        s.setb(R(i) + 0x7C, 100)
        hit.append(i)
    return own, pct, hit


def recompute_sums(s):
    """090A:0E6D → 2703:02E5 → 2703:034E：星域經濟力 ＝ 行星加總（32 位元）。"""
    for i in range(32):
        if not s.has(R(i) + 0x7D):
            continue
        tot = sum(s.d(P(p) + 0x1A) for p in range(first(i), first(i) + CNT[i]))
        s.setd(R(i) + 0x1A, tot)


def compare(tag, pred, real, before, skip=()):
    pp, pr = pred.tables()
    rp, rr = real.tables()
    bp, br = before.tables()
    moved = sum(1 for i in range(len(pp)) if pp[i] != bp[i]) + \
        sum(1 for i in range(len(pr)) if pr[i] != br[i])
    bad = []
    skipped = []
    for name, a, b, c, size in (('行星', pp, rp, bp, 82), ('星域', pr, rr, br, 126)):
        for i in range(len(a)):
            if name == '星域' and i % size in skip:
                if b[i] != c[i]:
                    skipped.append(i // size)
                continue
            if a[i] != b[i]:
                bad.append('%s %d +0x%02X：前 %02X 預測 %02X 實際 %02X'
                           % (name, i // size, i % size, c[i], a[i], b[i]))
    g = [(DS + 0x8F48, 'ds:8F48')]
    for lin, nm in g:
        if pred.b(lin) != real.b(lin):
            bad.append('%s：預測 %d 實際 %d' % (nm, pred.b(lin), real.b(lin)))
    print('%s：行星 81 × 82 ＋ 星域 %d 格 × 126 ＝ %d bytes 逐位元組比，預測會變的 %d bytes，不符 %d 處'
          % (tag, len(pr) // 126, len(pp) + len(pr), moved, len(bad)))
    for x in bad[:40]:
        print('   ', x)
    if skip:
        print('    （不模擬的欄位 %s：%d 格有變 %s）'
              % (' '.join('+0x%02X' % k for k in skip), len(skipped), skipped))
    return len(bad)


def main():
    pre, mid, post = (Snap(p) for p in sys.argv[1:4])
    print('前：回合 %d、ds:8F48 %d、ds:988E %d、費沙擁有者 %d、費沙行星經濟力 %d'
          % (pre.w(DS + 0x8F18), pre.b(DS + 0x8F48), pre.b(DS + 0x988E),
             pre.b(R(14) + 0x12), pre.d(P(first(14)) + 0x1A)))
    s = Snap(sys.argv[1])
    own, v, hit = body(s)
    print('本體預測：擁有者 %d → 搬走 %d；整格叛離 %s' % (own, v, hit or '（無）'))
    n = compare('前 → 本體之後', s, mid, pre)
    s2 = Snap(sys.argv[2])
    own, pct, hit = penalty(s2)
    if pct is None:
        print('懲罰預測：擁有者 %d 不是人類座位 → 不動' % own)
    else:
        print('懲罰預測：擁有者 %d 是人類座位、扣 %d%%、星域 %s' % (own, pct, hit))
    recompute_sums(s2)
    n += compare('本體之後 → 懲罰之後（含 090A:0E6D 的重算）', s2, post, mid, skip=(0x7D,))
    sys.exit(1 if n else 0)


if __name__ == '__main__':
    main()
