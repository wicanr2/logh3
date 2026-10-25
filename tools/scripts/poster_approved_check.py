"""164：核准HTML與正式Linux視窗的藝術區比較。僅在Docker執行。"""
import argparse
import hashlib
import json
import math
from pathlib import Path

import numpy as np
from PIL import Image

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--work', required=True)
a = p.parse_args()
root = Path(a.work)
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()

# The X11 window is exactly 1280x720; Ebitengine's screen downscale samples
# odd pixels of the 2560x1440 logical canvas. No guessed blur/filter tolerance.
def load(path):
    im = Image.open(path).convert('RGBA')
    if im.size == (2560, 1440):
        im = im.resize((1280, 720), Image.Resampling.NEAREST)
    assert im.size == (1280, 720)
    return np.asarray(im, dtype=np.int16)

def stats(diff, mask):
    ys, xs = np.where((diff > 0) & mask)
    return dict(pixels=int(mask.sum()), different=len(xs),
                max_channel=int(diff[mask].max()) if mask.any() else 0,
                bbox=None if not len(xs) else [int(xs.min()), int(ys.min()), int(xs.max()+1), int(ys.max()+1)])

mask = np.ones((720, 1280), dtype=bool)
excluded = {
    'live_game_content': [216, 96, 1064, 597],
    'upper_short_beam': [300, 40, 380, 60],
    'lower_short_beam': [974, 672, 1045, 691],
    'observed_game_cursor': [995, 612, 1030, 646],
}
for x0, y0, x1, y1 in excluded.values():
    mask[y0:y1, x0:x1] = False
rows = []
trace = json.loads((root/'final-seams/RESULT.json').read_text())
v = next(e['observed']['poster_decoration']['viewport'] for e in trace['events'] if e['name'] == 'decor-v1')
edge = np.zeros(mask.shape, dtype=bool)
for column in [math.floor(v[0]/2), math.floor((v[0]+v[2])/2)]:
    edge[math.floor(v[1]/2):math.ceil((v[1]+v[3])/2), column] = True
for index, name in enumerate(['decor-v1', 'decor-v2-cycled', 'decor-v3-cycled']):
    actual_path = root/'final-seams'/(name+'.png')
    expected_path = root/('reference-v'+str(index+1)+'.png')
    actual, expected = load(actual_path), load(expected_path)
    diff = np.abs(actual[:, :, :3]-expected[:, :, :3]).max(2)
    alpha = load(root/'art/hulls.png')[:, :, 3]
    composite = (alpha > 0) & (alpha < 255)
    if index == 0:
        composite |= load(root/'art/labels-zh-TW.png')[:, :, 3] > 0
    exact = stats(diff, mask & ~composite & ~edge)
    blended = stats(diff, mask & composite)
    boundary = stats(diff, mask & edge & ~composite)
    print(name, 'opaque', exact, 'alpha-composited', blended, flush=True)
    assert exact['max_channel'] == 0, '不透明藝術區與核准圖不同'
    assert blended['max_channel'] <= 3, '透明合成超出觀測的整數捨入差異'
    assert boundary['max_channel'] <= 1, '遊戲小數邊緣超出一階整數捨入差異'
    path = root/('final-diff-'+name+'.png')
    Image.fromarray(np.where(mask[:, :, None], np.minimum(np.abs(actual[:, :, :3]-expected[:, :, :3])*10, 255), 0).astype('uint8')).save(path)
    rows.append(dict(set='v'+str(index+1), actual_sha256=sha(actual_path),
                     reference_sha256=sha(expected_path), diff_sha256=sha(path),
                     all_art=stats(diff, mask), opaque=exact, alpha_composited=blended, fractional_viewport_edge=boundary))

old = root/'alliance'
first, after = load(old/'decor-v1.png'), load(old/'decor-after-loop.png')
body_mask = mask.copy()
body_mask[98:597, :] = False
delta = np.abs(first[:, :, :3]-after[:, :, :3]).max(2)
assert not delta[body_mask].any(), '艦體或背景隨相位移動'
result = dict(passed=True, comparison='approved-remake-art; not DOS parity',
              proposal_sha256=sha(root/'proposal.html'), excluded=excluded,
              sampling='exact 2:1 odd-pixel nearest sampling used by Ebitengine',
              rows=rows, static_loop=stats(delta, body_mask),
              toolchain='logh3-hd-preview:20261011-r1; Chromium1194 via playwright:v1.56.1-noble')
(root/'ART-CHECK.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
