# 遊戲的畫面文字表。掃資料段裡所有以 \0 分隔、可解成 Big5 且含中文的段落；
# 多數前面帶一個 `\x01` ＋ 排版格式字串（`[%d,%d,` 之類）。
import json
d = open('/work/workplace/strtab.bin', 'rb').read()

def big5(b):
    try:
        s = b.decode('big5hkscs')
    except Exception:
        return None
    if not any('　' <= c <= '鿿' for c in s):
        return None
    # 私用區與罕用區多半是把二進位資料誤讀成 Big5。
    if any('\ue000' <= c <= '\uf8ff' for c in s):
        return None
    if any(ord(c) < 9 or 13 < ord(c) < 32 for c in s):
        return None
    return s

out, i = [], 0
while i < len(d):
    j = d.find(b'\x00', i)
    if j < 0:
        break
    seg = d[i:j]
    if 3 < len(seg) < 200:
        if (t := big5(seg)) is not None:
            fmt = None
            # 前一段若是排版格式字串就一起記下來
            k = d.rfind(b'\x00', 0, max(i-1, 0))
            if k >= 0 and i - k < 40:
                prev = d[k+1:i-1]
                if prev[:1] == b'\x01':
                    fmt = prev[1:].decode('ascii', 'replace')
            out.append({'offset': i, 'format': fmt, 'text': t})
    i = j + 1

print('字串數', len(out))
json.dump({'source': 'GIN3PS.EXE 的資料段（執行期 DS，64 KB）',
           'note': '\\0 分隔；多數前面帶 \\x01 ＋ 排版格式字串',
           'count': len(out), 'strings': out},
          open('/work/docs/internals/data/screen-strings.json', 'w'),
          ensure_ascii=False, indent=1)
for r in out[:6] + out[-6:]:
    print(f"  +{r['offset']:05X} {r['text']!r}")
print('→ docs/internals/data/screen-strings.json')
