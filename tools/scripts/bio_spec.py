"""把「提督經歷」那一層併到提督詳細資料的規格上。

    tools/py.sh tools/scripts/bio_spec.py

`overlay_spec.py` 產出的視窗一律叫 `overlay`，而底層（提督詳細資料）
的視窗也叫 `overlay`——直接拿它當底層會**把底層那個視窗蓋掉**，
畫出來就少了上半的框與頭像。

這支把 `screen-admiral-bio.json` 裡量到的那一層改名成 `bio`，接在
詳細資料的規格後面（`order` ＝ overlay → bio），其餘欄位沿用量到的那份。
經歷的視窗是 (80,200)–(559,359)，只蓋住下半（`docs/spec/012`）。
"""
import copy
import json

BIO = '/work/docs/internals/data/screen-admiral-bio.json'
DETAIL = '/work/docs/internals/data/screen-admiral-detail.json'

bio = json.load(open(BIO))
detail = json.load(open(DETAIL))

w = bio['windows'].get('bio') or bio['windows']['overlay']
out = copy.deepcopy(detail)
out['windows']['bio'] = w
out['order'] = ['overlay', 'bio']
for k in ('mouse', 'hotspots'):
    if k in bio:
        out[k] = bio[k]
json.dump(out, open(BIO, 'w'), ensure_ascii=False, indent=1)
print('視窗', list(out['windows'].keys()), 'order', out['order'],
      '經歷那層', w['x'], w['y'], w['w'], w['h'], len(w.get('lines', [])), '行')
