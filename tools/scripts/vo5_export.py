#!/usr/bin/env python3
"""`.VO5` 結局圖 → PNG。

解碼常式來自 `ENDING.EXE`（IDA：`sub_102EC` 拆檔頭、`sub_10364` 解一個
plane）。版面：

    +0  u8   寬（bytes，×8 ＝ 像素寬）
    +1  u16  高（列）
    +3  …    壓縮資料，四個 plane 依序解，共用同一個來源指標

壓縮與 `.AMG` 同源，但是**逐 byte** 而不是逐 word：

    literal            寫下去，記成「上一個值」
    v v n              兩個相同的 byte ＝ 重複組，n ＝ 總共幾個 v
                       （n 是 1 byte；等於 0 時改讀後面的 word）

填法是**直行**：寫一個 byte 之後位址加「一列的 bytes 數」往下走，
走完一整行的高度就回到頂端、往右一格（原版是 `sub di, w*h-1`）。
"""
import glob, json, os, struct, sys
sys.path.insert(0, '/work/tools/scripts')
from pngout import write_png, game_palette
from vo5_decode import decode

OUT = '/work/workplace/assets/vo5'


os.makedirs(OUT, exist_ok=True)
# 結局有自己的一組調色盤，不能拿主畫面那組。
pal = game_palette('/work/workplace/explore/end1.png')
index = []
for f in sorted(glob.glob('/orig/*.VO5')):
    d = open(f, 'rb').read()
    name = os.path.splitext(os.path.basename(f))[0]
    try:
        w, h, px, used = decode(d)
    except (IndexError, struct.error) as e:
        print('%-10s 解不開：%s' % (name, e))
        index.append({'file': os.path.basename(f), 'ok': False})
        continue
    write_png(f'{OUT}/{name}.png', w, h, px, pal)
    index.append({'file': os.path.basename(f), 'width': w, 'height': h,
                  'bytes': len(d), 'consumed': used, 'ok': used == len(d)})
    print('%-10s %d×%d  用掉 %d / %d bytes %s'
          % (name, w, h, used, len(d), '' if used == len(d) else '← 有剩'))
json.dump({'note': '.VO5 結局圖；PNG 由 tools/scripts/vo5_export.py 從自備原版重建，不進 Git',
           'images': index},
          open('/work/docs/internals/data/vo5.json', 'w', encoding='utf-8'),
          ensure_ascii=False, indent=1)
