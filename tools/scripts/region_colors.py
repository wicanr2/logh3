"""32 個星域各塗什麼顏色，以及那份顏色從哪裡讀。

`2376:0BAF` 是「塗一格」的常式，顏色在 `[bp+12]`；呼叫端
（`375B0` 前一行）取的是 **`ds:[F162 + 星域編號]`** 的 word，
只用低位元組——所以 `ds:F162` 起的 32 個 bytes 就是每個星域的色號。

實測的呼叫順序：第一趟 626 次顏色全 0（626 ＝ 全部星域的格數，等於
先清一遍），第二趟才塗，而且**星域 0 排到最後**（順序是 1…31、0）。

用法：tools/py.sh tools/scripts/region_colors.py <記憶體dump> [顏色序列檔] [輸出json]

⚠ 輸出檔要指定。預設會寫 `star-region-colors.json`，那是**主畫面的勢力色**，
remake 的對拍在用；拿別的畫面的 dump 跑一次就會把它蓋掉。
"""
import json, pathlib, struct, sys

DS, TABLE = 0x54B3, 0xF162
NAMES = {0x06: '帝國首都', 0x01: '帝國', 0x03: '無主（底色）', 0x0C: '同盟'}

mem = pathlib.Path('/work/' + sys.argv[1]).read_bytes()
rgn = pathlib.Path('/orig/STRMAP.RGN').read_bytes()

rows, off = [], 0
while True:
    x, y, w, h, o = struct.unpack_from('<BBBBH', rgn, off)
    if w == 0 or h == 0 or o != sum(r[2] * r[3] * 40 for r in rows):
        break
    rows.append((x, y, w, h, o)); off += 6
cells = [r[2] * r[3] for r in rows]

base = DS * 16 + TABLE
colors = list(mem[base:base + len(rows)])
print('ds:%04X（線性 %05X）起的 %d 個 bytes：' % (TABLE, base, len(rows)))
for i, c in enumerate(colors):
    print('  %2d. 色 %2d  %s' % (i, c, NAMES.get(c, '？')))

if len(sys.argv) > 2:
    vals = [int(l.split(':')[1], 16) & 0xFF for l in
            pathlib.Path('/work/' + sys.argv[2]).read_text().split()]
    total = sum(cells)
    assert set(vals[:total]) == {0}, '第一趟不是全 0'
    # 順序不比對（星域 0 排在後面），比對**加權後的多重集合**：
    # 某個顏色塗掉幾格，要等於表裡用那個顏色的星域格數和。
    #
    # 首都（表裡是色 6 的那個星域）**會被塗兩次**：第二趟跟其他帝國星域
    # 一樣塗色 1，之後再單獨用色 6 重塗一次。所以帝國色會多出首都的格數。
    import collections
    got = collections.Counter(vals[total:])
    want = collections.Counter()
    for idx, c in enumerate(colors):
        want[c] += cells[idx]
    capital = [i for i, c in enumerate(colors) if c == 0x06]
    for i in capital:
        want[0x01] += cells[i]
    print('\n第二趟之後每色的格數：', dict(sorted(got.items())))
    print('表算出來（首都算兩次）：', dict(sorted(want.items())))
    print('一致' if got == want else '**不一致**')

out_path = sys.argv[3] if len(sys.argv) > 3 else \
    '/work/docs/internals/data/star-region-colors.json'
json.dump({
    'note': '32 個星域在 796 年 1 月帝國開局的著色色號。表在 ds:F162，'
            '每個星域一個 byte；著色常式 2376:0BAF 的 [bp+12] 就是從這裡取的。'
            '06 ＝ 帝國首都、01 ＝ 帝國、03 ＝ 無主（底色黑）、0C ＝ 同盟。'
            '這份表是執行期算出來的，來源未解（docs/re/17）。'
            '首都會被塗兩次：先跟其他帝國星域一樣塗 01，再用 06 重塗。',
    'table': 'ds:F162',
    'colors': colors,
}, open(out_path, 'w'), ensure_ascii=False, indent=1)
print('寫出', out_path)
