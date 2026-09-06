#!/usr/bin/env python3
"""`COMANI.BMG` — 戰鬥動畫的封存檔。

版面（實測）：

    +0      u16   第一段資料的位移（＝目錄結束的地方）
    +2 起   目錄：每筆 10 bytes
              +0  char[4]  名稱
              +4  u16      這一段的長度
              +6  u32      這一段在檔案裡的位移

驗證：把每一筆的「位移 ＋ 長度」拿去跟下一筆的位移比，整份完全接得起來；
最後一筆的結尾等於檔案長度。
"""
import json, os, struct, sys

SRC = sys.argv[1] if len(sys.argv) > 1 else '/orig/COMANI.BMG'
OUT = '/work/docs/internals/data/comani-bmg.json'

d = open(SRC, 'rb').read()
first = struct.unpack_from('<H', d, 0)[0]
entries, p = [], 2
while p + 10 <= first:
    name = d[p:p + 4]
    ln = struct.unpack_from('<H', d, p + 4)[0]
    off = struct.unpack_from('<I', d, p + 6)[0]
    if off < first or off + ln > len(d):
        break
    # 名稱的第 2 個位元組不一定是 ASCII（0xDD、0xB1…），當成代碼保留。
    entries.append({'name': name.decode('latin-1'), 'name_hex': name.hex(),
                    'offset': off, 'length': ln})
    p += 10

ok = True
for i, e in enumerate(entries):
    end = e['offset'] + e['length']
    nxt = entries[i + 1]['offset'] if i + 1 < len(entries) else len(d)
    if end != nxt:
        ok = False
        print('第 %d 筆 %s：結尾 %d，下一筆從 %d 開始' % (i, e['name'], end, nxt))

print('目錄結束於 %d（檔頭寫的第一段位移 %d），%d 筆，接得起來：%s'
      % (p, first, len(entries), ok))
print('前 8 筆：', [(e['name'], e['offset'], e['length']) for e in entries[:8]])
print('名稱第 2 個位元組：', sorted({e['name_hex'][2:4] for e in entries}))
print('名稱第 3–4 個位元組：', sorted({bytes.fromhex(e['name_hex'][4:]).decode('latin-1')
                                     for e in entries}))
json.dump({'file': os.path.basename(SRC), 'bytes': len(d),
           'directory_ends': p, 'contiguous': ok, 'count': len(entries),
           'entries': entries},
          open(OUT, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
print('→', OUT)
