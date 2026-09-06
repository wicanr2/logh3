#!/usr/bin/env python3
"""ENDING.EXE 的文字：結局台詞、職員表、演出名單。

`ENDING.EXE` 是未打包的 MZ，字串以 `\\0` 分隔連續放在 0x63FE 起的資料區，
一路排到 0x765C 的 `.VO5` 檔名清單為止。編碼是 Big5（倚天），大量使用
全形英文字母排版——職員表整段都是靠全形空白對齊的。

中文版由亨和企業．微波軟體代理發行，訊息翻譯邵程衍、程式 Andrew Liao；
這段職員表是原版檔案裡唯一記下中文化工作者姓名的地方。
"""
import json, os, re

SRC = '/orig/ENDING.EXE'
START, END = 0x63FE, 0x7660
# 0x659B 起不再是結局旁白，而是兩段尾聲對白（萊因哈特／吉爾菲艾斯、
# 楊威利／尤里安）。原版沒有標記，用位址切。
EPILOGUE = 0x659B
OUT_JSON = '/work/docs/internals/data/ending-text.json'
OUT_MD = '/work/docs/lore/ending-and-credits.md'

d = open(SRC, 'rb').read()
items, p = [], START
while p < min(END, len(d)):
    e = d.find(b'\x00', p)
    if e < 0:
        break
    s = d[p:e]
    if s:
        try:
            items.append({'offset': p, 'text': s.decode('big5hkscs')})
        except UnicodeDecodeError:
            pass
    p = e + 1

# 分段：兩個結局各自從【…】開始；職員表從「銀河英雄傳說ⅢＳＰ」那一行起；
# 演出名單是一整批全形英文；最後是 .VO5 檔名。
endings, credits, cast, files = [], [], [], []
cur = None
mode = 'ending'
for it in items:
    t = it['text']
    if re.fullmatch(r'[a-z0-9]+', t):
        files.append(t)
        continue
    if it['offset'] >= EPILOGUE and mode == 'ending':
        if not endings or endings[-1]['title'] != '尾聲':
            cur = {'title': '尾聲', 'lines': []}
            endings.append(cur)
    if t.startswith('【'):
        cur = {'title': t.strip('【】'), 'lines': []}
        endings.append(cur)
        continue
    if '銀河英雄傳說' in t and 'ＳＰ' in t:
        mode = 'credits'
    stripped = t.strip().strip('　')
    if mode == 'credits':
        # 全形英文大寫且沒有中文 ⇒ 演出名單
        if stripped and not re.search(r'[一-鿿]', stripped) and \
           re.fullmatch(r'[Ａ-Ｚ　\s－]+', stripped):
            cast.append(stripped)
        else:
            credits.append(t.rstrip())
    elif cur is not None:
        cur['lines'].append(stripped)

os.makedirs(os.path.dirname(OUT_JSON), exist_ok=True)
json.dump({'source': 'ENDING.EXE', 'range': [START, END],
           'endings': endings, 'credits': credits, 'cast': cast,
           'vo5_files': files, 'raw': items},
          open(OUT_JSON, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)

def squash(s):
    return re.sub(r'[　 ]+', ' ', s).strip()

os.makedirs(os.path.dirname(OUT_MD), exist_ok=True)
with open(OUT_MD, 'w', encoding='utf-8') as f:
    f.write("""# 結局與職員表

《銀河英雄傳說III SP》PC DOS 中文版的兩個結局台詞與完整職員表，逐字取自
原版 `ENDING.EXE`。原版用全形空白排版，這裡把連續空白收成一格。

中文版由亨和企業．微波軟體代理發行。職員表是原版檔案裡唯一記下中文化
工作者姓名的地方。

| 來源 | 值 |
|---|---|
| 檔案 | `ENDING.EXE`，64,677 bytes（未打包 MZ）|
| 字串區 | 0x63FE–0x765C，`\\0` 分隔 |
| 編碼 | Big5（倚天） |
| 資料 | `docs/internals/data/ending-text.json` |

## 結局

""")
    for e in endings:
        f.write('### %s\n\n' % e['title'])
        for ln in e['lines']:
            f.write('%s\n\n' % squash(ln) if squash(ln) else '')
        f.write('\n')
    f.write('## 職員表\n\n```\n')
    for c in credits:
        f.write(squash(c) + '\n' if squash(c) else '\n')
    f.write('''```

「ミラㄧジュ」的長音符寫成注音符號「ㄧ」——Big5 沒有片假名長音符，
中文化時拿字形一樣的注音符號頂替。這裡照原檔保留，不改成「ー」。

## 演出

原版在職員表之後逐一打出提督的羅馬拼音，兩兩一組（名／姓），
順序與 `ADMIRAL.DT*` 的提督索引一致（帝國在前、同盟在後）。

```
''')
    for c in cast:
        f.write(squash(c) + '\n')
    f.write('```\n\n## 結局腳本檔\n\n`ENDING.EXE` 依結局載入的 `.VO5`：\n\n')
    f.write(', '.join('`%s.VO5`' % x for x in files) + '\n')

print('結局 %d 段、職員表 %d 行、演出 %d 行、腳本檔 %d 個 → %s'
      % (len(endings), len(credits), len(cast), len(files), OUT_MD))
