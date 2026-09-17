# 提督頭像 ＋ 能力值對照表 → docs/lookup/admiral-portraits.md
#
# 給 #16（`+20`／`+22` 的欄名）人工查紙本「登場人物初期資料表」用：
# 一列一位提督，頭像、姓名、陣營、已命名的十欄與四個沒有名字的欄位並排。
# 數值取 `ADMIRAL.DTX`（新遊戲樣板）——紙本附的是「初期」資料，對的是樣板，
# 不是三格存檔（`+20` 在存檔裡會變，`docs/internals/08` 四個版本的差別）。
#
# 輸入：docs/internals/data/admiral.json、keireki.json（缺檔就停，不自製代用品）
# 頭像：assets/original/face/face-NN.png（`amg_export.py` 產生後複製進 repo）
import json, sys

D = '/work/docs/internals/data'
try:
    adm = json.load(open(f'{D}/admiral.json'))['variants']['DTX']
    bio = {r['index']: r for r in json.load(open(f'{D}/keireki.json'))['records']}
    face = {m['index']: m for m in json.load(open(f'{D}/face-amg.json'))['images']}
except FileNotFoundError as e:
    sys.exit(f'缺 {e.filename}')

# abilities[i] ＝ +(11 + i)；名字見 docs/internals/08「能力欄位的名字」
COLS = [(11, '統率力'), (12, '移動'), (13, '回轉'), (14, '營運'), (15, '光線砲'),
        (16, '飛彈'), (17, '戰鬥機'), (18, '防禦'), (19, '占領'),
        (20, '**+20**'), (21, '**+21**'), (22, '**+22**'), (23, '運用'), (24, '**+24**')]
SIDE = {0: '帝國', 1: '同盟'}

out = []
w = out.append
w('# 提督頭像與能力值對照表（人工查詢用）')
w('')
w('這份表給 [#16](https://github.com/wicanr2/logh3/issues/16) 拿紙本'
  '「登場人物初期資料表」或其他來源逐人比對用：找出 `+20`／`+21`／`+22`／`+24` '
  '在紙本上叫什麼。粗體欄是還沒有名字的欄位，其餘十欄已由提督詳細資料那張畫面'
  '標名（[`docs/internals/08`](../internals/08-admiral-dt.md)）。')
w('')
w('- 數值來源：`ADMIRAL.DTX`（新遊戲樣板），82 筆 × 28 bytes，欄位是記錄內位移。'
  '`+20` 在三格存檔裡會變，`+21`／`+22`／`+24` 四個版本一致。')
w('- 頭像：`FACE.AMG` 第 n 張 ＝ 提督索引 n（confirmed，'
  '[`docs/internals/09`](../internals/09-face-amg.md)），檔案在 '
  '[`assets/original/face/`](../../assets/original/face/)，像素與 PNG 雜湊在 '
  '`docs/internals/data/face-amg.json`。')
w('- 姓名與陣營：`KEIREKI.DAT`（`docs/internals/data/keireki.json`）。')
w('- 比對時的已知事實：公開的 Win 版能力表只有那十個已命名欄位，數值逐欄吻合，'
  '沒有粗體那四欄（[`docs/re/287`](../re/287-admiral-remaining-fields.md)）。'
  '紙本若有第十一欄以上，先拿萊因哈特（第 0 列）對 `+20`＝'
  f'{adm[0]["abilities"][9]}、`+21`＝{adm[0]["abilities"][10]}、'
  f'`+22`＝{adm[0]["abilities"][11]}、`+24`＝{adm[0]["abilities"][13]}。')
w('')
w('由 `tools/py.sh tools/scripts/admiral_lookup.py` 產生，不要手改。')
w('')
w('| # | 頭像 | 姓名 | 陣營 | ' + ' | '.join(n for _, n in COLS) + ' |')
w('|---:|---|---|---|' + '---:|' * len(COLS))
for r in adm:
    k = r['index']
    b = bio.get(k, {})
    assert face[k].get('admiral') == k, k
    img = f'<img src="../../assets/original/face/face-{k:02d}.png" width="48" alt="{k}">'
    vals = ' | '.join(str(r['abilities'][off - 11]) for off, _ in COLS)
    w(f'| {k} | {img} | {b.get("name", "")} | {SIDE.get(b.get("side"), "")} | {vals} |')
w('')
w('第 82 張頭像是沒有人物的單色格（不對應任何提督），不列入。')
w('')

import os
os.makedirs('/work/docs/lookup', exist_ok=True)
open('/work/docs/lookup/admiral-portraits.md', 'w').write('\n'.join(out))
print('寫出 docs/lookup/admiral-portraits.md，%d 列' % len(adm))
