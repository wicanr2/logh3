"""把戰場畫面上的單位逐一對回艦艇記錄，看塊號由哪個欄位決定。

    tools/py.sh tools/scripts/unit_tiles.py <批次> <記憶體 dump> <視野欄> <視野列>

畫面 dump 是 640×480 的 8bpp indexed，單位是 `SPRITE.SPR` 的 16×16 塊
（`docs/re/85`）。這支腳本不猜規則，只把「記錄欄位」與「畫面上實際
出現的塊號」並排列出來。
"""
import pathlib
import sys

W, H = 640, 480
DS = 0x54B30                      # ds 段基底：ds:C074 ＝ 艦隊表 0x60BA4
FLEET_TAB = DS + 0xC074
FLEET_N, FLEET_SZ = 32, 122
POOL_PTR = DS + 0xCFB4
POOL_SZ = 78

batch, memname, vcol, vrow = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
mem = pathlib.Path('/work/workplace/' + memname).read_bytes()
scr = pathlib.Path('/work/workplace/oracle/%s/screen.bin' % batch).read_bytes()
spr = pathlib.Path('/orig/SPRITE.SPR').read_bytes()

u16 = lambda a: mem[a] | (mem[a + 1] << 8)
s16 = lambda a: u16(a) - 0x10000 if u16(a) >= 0x8000 else u16(a)

pool = (u16(POOL_PTR + 2) << 4) + u16(POOL_PTR)


def sprite_block(i):
    out = []
    base = i * 128
    for y in range(16):
        row = [0] * 16
        for p in range(4):
            for b in range(2):
                v = spr[base + y * 8 + p * 2 + b]
                for k in range(8):
                    if v & (0x80 >> k):
                        row[b * 8 + k] |= 1 << p
        out.append(row)
    return out


blocks = [sprite_block(i) for i in range(len(spr) // 128)]


def screen_block(x, y):
    if x < 0 or y < 0 or x + 16 > W or y + 16 > H:
        return None
    return [[scr[(y + j) * W + x + i] for i in range(16)] for j in range(16)]


def match(tile):
    hits = [i for i, b in enumerate(blocks) if b == tile]
    return hits[0] if hits else None


def predict(vis, kind):
    if vis == 2:
        return 66 + kind
    if vis == 1:
        return 74
    return None


print('艦隊 組 序 陣營 艦種 +17 +2F 存活  世界座標      畫面      塊 預測')
bad = 0
for f in range(FLEET_N):
    rec = FLEET_TAB + f * FLEET_SZ
    if mem[rec + 4] == 0xFF:
        continue
    ngroup = mem[rec + 0x0F]
    for g in range(ngroup):
        idx = u16(rec + 0x10 + g * 4)
        p = pool + idx * POOL_SZ
        wx, wy = s16(p + 0x28), s16(p + 0x2A)
        sx = 16 + wx - vcol * 16
        sy = 60 + wy - vrow * 20
        tile = screen_block(sx, sy)
        blk = match(tile) if tile else None
        vis = mem[p + 0x17]
        pre = predict(vis, mem[p + 4])
        mark = ''
        if tile is not None and blk is not None and pre != blk:
            mark = '  ✗'
            bad += 1
        print('%4d %3d %3d %4d %4d %3d %3d %5d  (%5d,%5d) (%4d,%4d) %s %s%s'
              % (f, g, mem[p + 3], mem[p + 0], mem[p + 4], vis, mem[p + 0x2F],
                 u16(p + 8),
                 wx, wy, sx, sy, '-' if blk is None else blk,
                 '-' if pre is None else pre, mark))


print('對不上的樣本：%d' % bad)
