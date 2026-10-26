"""150：下載官方經典動畫人物參考，建立供image_gen辨認的聯絡表。Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import urllib.request

from PIL import Image, ImageDraw, ImageFont

ROOT = Path('/src')
OUT = ROOT/'workplace/poster150/references'
PEOPLE = {
    'reinhard': ('g01', 'Reinhard'), 'kircheis': ('g02', 'Siegfried Kircheis'),
    'hilda': ('g03', 'Hildegard'), 'annerose': ('g07', 'Annerose'),
    'mittermeyer': ('g04', 'Mittermeyer'), 'reuenthal': ('g05', 'Reuenthal'),
    'oberstein': ('g06', 'Oberstein'), 'yang': ('d01', 'Yang Wen-li'),
    'frederica': ('d02', 'Frederica'), 'julian': ('d03', 'Julian'),
    'attenborough': ('d04', 'Attenborough'), 'cazerne': ('d08', 'Cazerne'),
    'schoenkopf': ('d05', 'Schoenkopf'), 'poplan': ('d07', 'Poplan'),
}
SETS = {
    'empire-admirals': ['reinhard', 'kircheis', 'mittermeyer', 'reuenthal', 'oberstein'],
    'alliance-admirals': ['yang', 'attenborough', 'cazerne', 'schoenkopf', 'poplan'],
    'empire-close': ['reinhard', 'annerose', 'kircheis', 'hilda'],
    'alliance-close': ['yang', 'frederica', 'julian', 'cazerne', 'attenborough', 'schoenkopf'],
}


def main():
    assert os.getuid() == os.getgid() == 1000
    assert not OUT.exists() and (ROOT/'workplace').stat().st_uid == os.getuid()
    OUT.mkdir(parents=True)
    rows = []
    for name, (code, label) in PEOPLE.items():
        url = 'https://ginei.jp/img/char/'+code+'.jpg'
        request = urllib.request.Request(url, headers={'User-Agent': 'LOGH3 local artwork reference'})
        data = urllib.request.urlopen(request, timeout=30).read(2*1024*1024)
        target = OUT/(name+'.jpg'); target.write_bytes(data)
        with Image.open(target) as image: image.verify()
        rows.append(dict(name=name, label=label, file=target.name, url=url, bytes=len(data), sha256=hashlib.sha256(data).hexdigest()))
    font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf', 22)
    for name, members in SETS.items():
        sheet = Image.new('RGB', (900, 660), '#101927'); draw = ImageDraw.Draw(sheet)
        for i, member in enumerate(members):
            image = Image.open(OUT/(member+'.jpg')).convert('RGB'); image.thumbnail((280, 270))
            x, y = (i%3)*300, (i//3)*330
            sheet.paste(image, (x+(300-image.width)//2, y+8))
            draw.text((x+10, y+285), PEOPLE[member][1], font=font, fill='white')
        sheet.save(OUT/(name+'.png'))
    (OUT/'SOURCES.json').write_text(json.dumps(dict(pages=['https://ginei.jp/character_g.php', 'https://ginei.jp/character_d.php'], people=rows, sheets=SETS), ensure_ascii=False, indent=2)+'\n')
    print('PASS', len(rows), '官方人物參考與4份辨認聯絡表', flush=True)


if __name__ == '__main__': main()
