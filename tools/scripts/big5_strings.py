# 從一支未打包的執行檔或資料檔裡撈出 Big5（倚天）字串。
#   tools/py.sh tools/scripts/big5_strings.py <檔名> [最少幾個中文字]
#
# 判準：連續的 Big5 雙位元組（0xA1–0xF9 開頭），中間允許可列印 ASCII，
# 以 \0 或不可解的位元組收尾。
import json, os, sys

name = sys.argv[1]
MIN = int(sys.argv[2]) if len(sys.argv) > 2 else 2
d = open('/orig/' + name, 'rb').read()

def lead(b):
    return 0xA1 <= b <= 0xF9

out, i = [], 0
while i < len(d):
    if lead(d[i]) and i + 1 < len(d) and (0x40 <= d[i+1] <= 0x7E or 0xA1 <= d[i+1] <= 0xFE):
        j, han = i, 0
        while j < len(d):
            if lead(d[j]) and j + 1 < len(d) and (0x40 <= d[j+1] <= 0x7E or 0xA1 <= d[j+1] <= 0xFE):
                j += 2; han += 1
            elif 0x20 <= d[j] < 0x7F:
                j += 1
            else:
                break
        if han >= MIN:
            try:
                t = d[i:j].decode('big5hkscs')
            except UnicodeDecodeError:
                t = d[i:j].decode('big5hkscs', errors='replace')
            out.append({'offset': i, 'text': t})
        i = j if j > i else i + 1
    else:
        i += 1

base = os.path.splitext(name)[0].lower()
p = f'/work/workplace/strings-{base}.json'
json.dump({'file': name, 'count': len(out), 'strings': out},
          open(p, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
print(name, len(d), 'bytes →', len(out), '筆 →', p)
for s in out[:8]:
    print(' ', hex(s['offset']), s['text'][:50])
