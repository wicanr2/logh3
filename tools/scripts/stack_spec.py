"""把 `overlay_spec.py` 量出來的那一層併到底層規格上，而不是蓋掉它。

    tools/py.sh tools/scripts/stack_spec.py <底層規格> <這一份規格> <新視窗名>

`overlay_spec.py` 產出的視窗一律叫 `overlay`。底層如果**也**叫
`overlay`（疊在別的疊層上的第三層就是這種情況），直接用會把底層那個
視窗整個蓋掉——畫出來少一整層框，差幾萬個像素。

這支把量到的那一層改名，接在底層後面（`order` ＝ 底層…→ 新名字），
`mouse` 與 `hotspots` 沿用量到的那一份。可以重跑：已經併過就拿新名字
那個視窗再併一次。

`bio_spec.py` 是這支的前身（提督經歷那一層），差別只在它寫死了檔名。
"""
import copy
import json
import sys

base_name, name, win = sys.argv[1], sys.argv[2], sys.argv[3]
BASE = f'/work/docs/internals/data/{base_name}.json'
PATH = f'/work/docs/internals/data/{name}.json'

spec = json.load(open(PATH))
base = json.load(open(BASE))

w = spec['windows'].get(win) or spec['windows']['overlay']
out = copy.deepcopy(base)
out['windows'][win] = w
order = [n for n in (base.get('order') or ['overlay']) if n in out['windows']]
out['order'] = order + [win]
for k in ('mouse', 'hotspots'):
    if k in spec:
        out[k] = spec[k]
json.dump(out, open(PATH, 'w'), ensure_ascii=False, indent=1)
print(f'{name}：視窗 {list(out["windows"])} order {out["order"]}；'
      f'{win} ＝ ({w["x"]},{w["y"]}) {w["w"]}×{w["h"]}、{len(w.get("lines", []))} 行')
