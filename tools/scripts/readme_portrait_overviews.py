"""以既有將領頭像製作README四張總覽，不修改來源圖片。"""
import argparse
import hashlib
import io
import json
import os
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont, ImageOps, __version__ as pillow_version

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--check', action='store_true', help='重生至記憶體並核對已保存圖片與索引')
a = p.parse_args()
root, output = Path('/ref'), Path('docs/images')
assert output.is_dir() and output.stat().st_uid == os.getuid()
catalog_path = root/'docs/internals/data/keireki.json'
manifest_path = root/'assets/hd/MANIFEST.json'
catalog = json.loads(catalog_path.read_text())['records']
manifest = json.loads(manifest_path.read_text())
hd_entries = {r['index']: r for r in manifest['files'] if 'index' in r}
assert len(catalog) == 82 and sorted(r['index'] for r in catalog) == list(range(82))
assert {r['side'] for r in catalog} == {0, 1}
font_dir = Path('/usr/share/fonts/opentype/noto')
regular = font_dir/'NotoSansCJK-Regular.ttc'
bold = font_dir/'NotoSansCJK-Bold.ttc'
title_font = ImageFont.truetype(str(bold), 44)
subtitle_font = ImageFont.truetype(str(regular), 23)
name_font = ImageFont.truetype(str(bold), 26)
index_font = ImageFont.truetype(str(regular), 18)

def sha(data):
    return hashlib.sha256(data).hexdigest()

def centered(draw, text, font, box, color):
    left, top, right, bottom = draw.textbbox((0, 0), text, font=font)
    x, y, w, h = box
    assert right-left <= w and bottom-top <= h, (text, box)
    draw.text((x+(w-(right-left))//2-left, y+(h-(bottom-top))//2-top), text, font=font, fill=color)

sources, products = [], []
source_records = {}
columns, portrait_w, portrait_h = 7, 240, 300
pitch_x, pitch_y, margin, header, footer = 264, 364, 28, 126, 32
for faction, slug, title, accent in [(0, 'empire', '銀河帝國', '#d7b96b'),
                                    (1, 'alliance', '自由行星同盟', '#ed9b8f')]:
    people = sorted((r for r in catalog if r['side'] == faction), key=lambda r: r['index'])
    assert len(people) == (40 if faction == 0 else 42)
    rows = (len(people)+columns-1)//columns
    for mode, label in [('original', '原版'), ('hd', 'HD')]:
        canvas = Image.new('RGB', (margin*2+columns*pitch_x, header+rows*pitch_y+footer), '#0d1422')
        draw = ImageDraw.Draw(canvas)
        draw.text((margin, 20), title+'｜'+label, font=title_font, fill=accent)
        draw.text((margin, 78), f'全 {len(people)} 位將領 · 依遊戲人物索引排列 · 姓名沿用原版', font=subtitle_font, fill='#c2cbd8')
        cells = []
        for n, person in enumerate(people):
            idx = person['index']
            entry = hd_entries[idx]
            assert entry['name'] == person['name']
            original = root/f'assets/original/face/face-{idx:02d}.png'
            high = root/'assets/hd'/entry['file']
            original_bytes, high_bytes = original.read_bytes(), high.read_bytes()
            assert sha(original_bytes) == entry['original_sha256']
            assert sha(high_bytes) == entry['sha256']
            source_records[idx] = dict(index=idx, name=person['name'], faction=faction,
                                      original=dict(file=str(original.relative_to(root)), sha256=sha(original_bytes)),
                                      hd=dict(file=str(high.relative_to(root)), sha256=sha(high_bytes)))
            path = original if mode == 'original' else high
            with Image.open(path) as source:
                if mode == 'original':
                    assert source.size == (80, 100)
                    picture = source.convert('RGB').resize((portrait_w, portrait_h), Image.Resampling.NEAREST)
                else:
                    assert source.size == (entry['width'], entry['height'])
                    picture = ImageOps.contain(source.convert('RGB'), (portrait_w, portrait_h), Image.Resampling.LANCZOS)
            x = margin+(n % columns)*pitch_x+12
            y = header+(n // columns)*pitch_y
            px, py = x+(portrait_w-picture.width)//2, y+(portrait_h-picture.height)//2
            canvas.paste(picture, (px, py))
            draw.rectangle((x-1, y-1, x+portrait_w, y+portrait_h), outline='#3c485b', width=1)
            centered(draw, person['name'], name_font, (x, y+portrait_h+8, portrait_w, 31), '#eff2f7')
            centered(draw, f'{idx:02d}', index_font, (x, y+portrait_h+43, portrait_w, 22), '#95a4b9')
            assert canvas.crop((px, py, px+picture.width, py+picture.height)).tobytes() == picture.tobytes()
            cells.append(dict(index=idx, name=person['name'], portrait_box=[px, py, picture.width, picture.height]))
        buffer = io.BytesIO()
        canvas.save(buffer, format='PNG', optimize=True)
        data = buffer.getvalue()
        name = f'readme-admirals-{slug}-{mode}.png'
        dest = output/name
        if a.check:
            assert dest.read_bytes() == data, name
        else:
            assert not dest.exists(), 'refuse overwrite: '+name
            dest.write_bytes(data)
        assert dest.stat().st_uid == os.getuid()
        products.append(dict(file='docs/images/'+name, faction=faction, mode=mode, count=len(people),
                             width=canvas.width, height=canvas.height, bytes=len(data), sha256=sha(data), cells=cells))
        print(name, len(people), canvas.size, len(data), flush=True)

document = dict(schema='logh3.readme-admiral-overviews.v1',
                generator='tools/scripts/readme_portrait_overviews.py',
                generator_sha256=sha((root/'tools/scripts/readme_portrait_overviews.py').read_bytes()),
                docker_image='logh3-hd-preview:20261011-r1', pillow_version=pillow_version,
                fonts=[dict(file=str(f), sha256=sha(f.read_bytes())) for f in [regular, bold]],
                catalog_sha256=sha(catalog_path.read_bytes()), hd_manifest_sha256=sha(manifest_path.read_bytes()),
                method='Existing portraits only; original nearest-neighbour 3x; HD aspect-preserving contain; no crop; identical index order.',
                classification=manifest['classification'], excluded_blank_index=82,
                people=[source_records[i] for i in sorted(source_records)], products=products)
index = output/'readme-admirals.json'
data = (json.dumps(document, ensure_ascii=False, indent=2)+'\n').encode()
if a.check:
    assert index.read_bytes() == data, 'index changed'
else:
    assert not index.exists(), 'refuse overwrite index'
    index.write_bytes(data)
assert index.stat().st_uid == os.getuid()
print('PASS: 82 people, four complete sheets; source hashes, names, cells and ownership checked')
