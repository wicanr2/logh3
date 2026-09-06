# STRMAP.RGN：32 個星域區域遮罩。
# 檔頭 32 筆 × 6 bytes ＝ {塊x, 塊y, 塊寬, 塊高, u16 位移（相對資料區）}；
# 遮罩是 1bpp、逐像素列橫跨整個區域寬度（每列 w*2 bytes）。
import struct, sys, collections
sys.path.insert(0, '/work/tools/scripts')

TW, TH, COLS = 16, 20, 20
def sheet(f):
    n = len(f)//(TH*8); W = COLS*TW
    px = bytearray(W*(n//COLS)*TH)
    for k in range(n):
        tx, ty = (k % COLS)*TW, (k//COLS)*TH
        d = f[k*TH*8:(k+1)*TH*8]
        for y in range(TH):
            w4 = struct.unpack_from('<4H', d, y*8)
            for x in range(TW):
                v = 0
                for p in range(4):
                    be = ((w4[p] & 0xFF) << 8) | (w4[p] >> 8)
                    if be & (0x8000 >> x):
                        v |= 1 << p
                px[(ty+y)*W+tx+x] = v
    return px, W, (n//COLS)*TH

base, W, H = sheet(open('/orig/STRMAP.BIT', 'rb').read())
live, _, _ = sheet(open('/work/workplace/strmap-live.bin', 'rb').read())
rgn = open('/orig/STRMAP.RGN', 'rb').read()
HDR = 192
regions = [struct.unpack_from('<4BH', rgn, i*6) for i in range(HDR//6)]
print('區域數', len(regions))

out = bytearray(base)
report = []
colours = {}
for i, (bx, by, bw, bh, off) in enumerate(regions):
    x0, y0 = bx*TW, by*TH
    w, h = bw*TW, bh*TH
    stride = bw*2
    need = stride*h
    d = rgn[HDR+off:HDR+off+need]
    if len(d) < need:
        report.append((i, '資料不足'))
        continue
    # 顏色：從 live 反推（遮罩內、底圖是 0 的地方 live 是什麼色）
    votes = collections.Counter()
    px = []
    for y in range(h):
        for x in range(w):
            # 每塊 40 bytes 連續（20 列 × 2 bytes），塊之間 row-major。
            # 「逐像素列橫跨整個區域」與這個排列在一個圖塊列上都是 200 bytes，
            # 光看 BP 的步進分不出來。
            bi = (y//TH)*bw + (x//TW)
            bit = d[bi*40 + (y % TH)*2 + (x % TW)//8] & (0x80 >> (x % 8))
            if bit:
                gx, gy = x0+x, y0+y
                if gx < W and gy < H:
                    px.append((gx, gy))
                    if base[gy*W+gx] == 0:
                        votes[live[gy*W+gx]] += 1
    colour = votes.most_common(1)[0][0] if votes else 0
    # 原版是「先把遮罩涵蓋的位元清掉，再從 EMS 的底本重畫」，
    # 所以**後畫的區域會覆蓋先畫的**，不是先到先得。
    for gx, gy in px:
        if base[gy*W+gx] == 0:
            out[gy*W+gx] = colour
    colours[i] = colour
    report.append((i, f'({bx},{by}) {bw}×{bh} 塊 色{colour} 覆蓋 {len(px)} px 純度 {votes[colour]*100//max(1,sum(votes.values()))}%'))

bad = sum(1 for i in range(W*H) if out[i] != live[i])
print(f'重建 vs 記憶體版本：{bad} / {W*H} 不同')

import json, os
json.dump({'source': 'STRMAP.RGN', 'header_bytes': HDR, 'entry': 6,
           'entry_fields': ['塊x', '塊y', '塊寬', '塊高', 'u16 位移（相對資料區）'],
           'tile': {'w': TW, 'h': TH}, 'mask': '1bpp，每塊 40 bytes，塊之間 row-major',
           'regions': [{'index': i, 'bx': r[0], 'by': r[1], 'bw': r[2], 'bh': r[3],
                        'offset': r[4], 'colour_796_01': colours[i]}
                       for i, r in enumerate(regions)]},
          open('/work/docs/internals/data/strmap-regions.json', 'w'),
          ensure_ascii=False, indent=1)
print('→ docs/internals/data/strmap-regions.json')

png = open('/work/workplace/oracle/key1/screen.png', 'rb').read()
i, pal = 8, None
while i < len(png):
    ln = struct.unpack_from('>I', png, i)[0]
    if png[i+4:i+8] == b'PLTE':
        pal = list(png[i+8:i+8+ln]); break
    i += 12 + ln
sys.path.insert(0, '/work/tools/scripts')
from amg_export import write_png
os.makedirs('/work/workplace/assets', exist_ok=True)
write_png('/work/workplace/assets/strmap-live.png', W, H, bytes(out), (pal+[0]*768)[:48])
print('→ workplace/assets/strmap-live.png')
