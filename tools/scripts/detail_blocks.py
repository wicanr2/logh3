# -*- coding: utf-8 -*-
"""數戰場上有幾個顯示層級 2 的方塊（`docs/re/159`）。

    tools/py.sh tools/scripts/detail_blocks.py <批次> [批次...]

顯示層級 2 的分隊畫成 `SPRITE.SPR` 第 38 塊：16×16、外圈一整圈色 6、
內圈一圈色 1、內部 12×12 印艦種代號（`docs/re/113`）。找那個外框圖樣。
"""
import pathlib
import sys

W = 640
X0, Y0, X1, Y1 = 16, 60, 463, 419      # 戰場
root = pathlib.Path('/work')
for batch in sys.argv[1:] or ['s3data0', 's3detail_back']:
    scr = (root / ('workplace/oracle/%s/screen.bin' % batch)).read_bytes()
    hits = []
    for y in range(Y0, Y1 - 16):
        for x in range(X0, X1 - 16):
            if (all(scr[y * W + x + i] == 6 for i in range(16))
                    and all(scr[(y + j) * W + x] == 6 for j in range(16))
                    and all(scr[(y + 1) * W + x + 1 + i] == 1 for i in range(14))):
                hits.append((x, y))
    print('%-16s 方塊 %d 個  %s' % (batch, len(hits), hits))
