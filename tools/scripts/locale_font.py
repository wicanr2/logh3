"""153：由唯一TSV生成Noto五語遮罩及Big5顯示解碼表。Docker內執行。"""
import csv
import hashlib
import json
import math
import os
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path('/src'); OUT = ROOT/'internal/localize'
FONT = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')
LOCALES = ['zh-TW', 'zh-CN', 'ja', 'en', 'ko']
INDEX = [3, 2, 0, 0, 1]


def main():
    assert OUT.stat().st_uid == os.getuid() == os.getgid() == 1000
    with (OUT/'ui.tsv').open() as f: rows = list(csv.reader(f, delimiter='\t'))
    assert rows[0] == ['key', *LOCALES]
    assert len({r[0] for r in rows[1:]}) == len(rows)-1
    assert all(len(r) == 6 and all(v.strip() for v in r) for r in rows[1:])
    original = set(''.join(r[1] for r in rows[1:])) | set(map(chr, range(32, 127))) | set(map(chr, range(0xFF01, 0xFF61))) | set('　…。，：！？％／「」・﹕、▲▼−＋×%')
    config = dict(locales=LOCALES, source_sha256=hashlib.sha256(FONT.read_bytes()).hexdigest(), glyphs={})
    sizes = [12, 36]
    characters = {locale: sorted((original | set(''.join(r[li+1] for r in rows[1:]))) - set('\n\r\t')) for li, locale in enumerate(LOCALES)}
    for size in sizes:
        cell = size*4//3; columns = 64; height = math.ceil(sum(map(len, characters.values()))/columns)*cell
        image = Image.new('L', (columns*cell, height)); draw = ImageDraw.Draw(image)
        total = 0
        for locale, index in zip(LOCALES, INDEX):
            font = ImageFont.truetype(str(FONT), size, index=index)
            missing = bytes(font.getmask('\uffff'))
            for ch in characters[locale]:
                assert ch.isspace() or bytes(font.getmask(ch)) != missing, ('缺字', locale, ch)
                x, y = (total%columns)*cell, (total//columns)*cell
                width = max(1, math.ceil(font.getlength(ch)))
                assert width <= cell
                draw.text((x, y+size*11//12), ch, font=font, fill=255, anchor='ls')
                g = config['glyphs'].setdefault(locale+'/'+str(ord(ch)), {})
                g[str(size)] = dict(x=x, y=y, w=width, h=cell)
                total += 1
        image.save(OUT/f'font-{size}.png')
    (OUT/'font.json').write_text(json.dumps(config, ensure_ascii=False, separators=(',', ':'))+'\n')
    mapping = {}
    for high in range(0xA1, 0xFA):
        for low in [*range(0x40, 0x7F), *range(0xA1, 0xFF)]:
            code = high*256+low
            try: value = bytes([high, low]).decode('cp950')
            except UnicodeDecodeError: continue
            if len(value) == 1: mapping[str(code)] = ord(value)
    (OUT/'big5.json').write_text(json.dumps(mapping, separators=(',', ':'))+'\n')
    source = dict(font=str(FONT), font_sha256=config['source_sha256'], catalog_sha256=hashlib.sha256((OUT/'ui.tsv').read_bytes()).hexdigest(),
                  generator_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), licenses='Noto font SIL OFL 1.1，沿settings_font.LICENSE.txt',
                  rows=len(rows)-1, glyphs=len(config['glyphs']), big5=len(mapping), sizes=sizes,
                  outputs={n: hashlib.sha256((OUT/n).read_bytes()).hexdigest() for n in ['font.json', 'font-12.png', 'font-36.png', 'big5.json']})
    (OUT/'SOURCE.json').write_text(json.dumps(source, ensure_ascii=False, indent=2)+'\n')
    (OUT/'LICENSE.font.txt').write_bytes((ROOT/'internal/font/settings_font.LICENSE.txt').read_bytes())
    print('PASS', source['rows'], '五語TSV，', source['glyphs'], '字形；無缺字', flush=True)


if __name__ == '__main__': main()
