# FACE.AMG → 83 張 PNG（4bpp 調色盤圖）＋ 格式 JSON。
# PNG 落在 workplace/（原版素材不進 Git），JSON 只放版面與索引。
import json, os, struct, sys, zlib
sys.path.insert(0, '/work/tools/scripts')
from amg_decode import decode

def write_png(path, w, h, px, pal):
    raw = b''.join(b'\x00' + px[y*w:(y+1)*w] for y in range(h))
    def chunk(t, d):
        c = t + d
        return struct.pack('>I', len(d)) + c + struct.pack('>I', zlib.crc32(c))
    with open(path, 'wb') as fp:
        fp.write(b'\x89PNG\r\n\x1a\n')
        fp.write(chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 3, 0, 0, 0)))
        fp.write(chunk(b'PLTE', bytes(pal)))
        fp.write(chunk(b'IDAT', zlib.compress(raw, 9)))
        fp.write(chunk(b'IEND', b''))

# 調色盤取自 oracle 的畫面 PNG：色號 → 屬性暫存器 → DAC 這條鏈由執行器
# 走完了，`pal.bin` 只有最後一段（DAC），直接拿會得到錯的顏色。
png = open('/work/workplace/oracle/face-fresh/screen.png', 'rb').read()
i = 8
pal = None
while i < len(png):
    n = struct.unpack_from('>I', png, i)[0]
    t = png[i+4:i+8]
    if t == b'PLTE':
        pal = list(png[i+8:i+8+n])
        break
    i += 12 + n
pal = (pal + [0]*768)[:48]

f = open('/orig/FACE.AMG', 'rb').read()
idx = [struct.unpack_from('<HI', f, k*6) for k in range(83)]
out = '/work/workplace/assets/face'
os.makedirs(out, exist_ok=True)
meta = []
for k, (s, o) in enumerate(idx):
    img = f[o:o+s]
    W, H, px, used = decode(img)
    write_png(f'{out}/face-{k:02d}.png', W, H, px, pal)
    a, b = struct.unpack_from('<II', img, 0)
    meta.append({'index': k, 'size': s, 'offset': o, 'w': W, 'h': H,
                 'used': used, 'exact': used == s,
                 'hdr0': a, 'hdr4': b, 'flag12': img[12],
                 'colors': sorted(set(px))})
json.dump({'source': 'FACE.AMG', 'count': len(idx),
           'index_entry': {'size': 'u16 @+0', 'offset': 'u32 @+2'},
           'image_header': {'+0': 'u32 固定 1', '+4': 'u32 size-6',
                            '+8': 'u16 直行數（每行 8 像素）',
                            '+10': 'u16 每直行的 word 數（每 word 2 列）',
                            '+12': 'u8 0 ＝ 一般、0xFF ＝ 空白', '+13': '影像資料'},
           'images': meta},
          open('/work/docs/internals/data/face-amg.json', 'w'),
          ensure_ascii=False, indent=1)
print('導出', len(meta), '張 →', out)
print('全部剛好用完:', all(m['exact'] for m in meta))
print('空白張（flag12=0xFF）:', [m['index'] for m in meta if m['flag12'] == 0xFF])
