"""列出小地圖對拍剩下的像素，看它們落在哪、是什麼顏色。

    tools/py.sh tools/scripts/minimap_residual.py <批次>

remake 端的小地圖由 `internal/screen` 畫，這裡不重畫，只把原版畫面上
小地圖那一塊與「地形取樣」重建出來的底圖相減，把差異列出來。
"""
import json
import pathlib
import sys

W = 640
batch = sys.argv[1]
root = pathlib.Path('/work')
spec = json.loads((root / 'docs/internals/data/screen-battle.json').read_text(encoding='utf-8'))
scr = (root / ('workplace/oracle/%s/screen.bin' % batch)).read_bytes()
m = spec['minimap']
x0, y0, w, h = m['x'], m['y'], m['w'], m['h']
print('小地圖 (%d,%d) %d×%d' % (x0, y0, w, h))
hist = {}
for j in range(h):
    for i in range(w):
        v = scr[(y0 + j) * W + x0 + i]
        hist[v] = hist.get(v, 0) + 1
print('顏色分布：', sorted(hist.items(), key=lambda kv: -kv[1]))
