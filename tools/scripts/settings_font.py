"""148：固定Docker字庫產生設定專用原生12／高解析36字形；--check獨立重生核對。"""
import argparse
import hashlib
import io
import json
from pathlib import Path
import re

from PIL import Image, ImageDraw, ImageFont, __version__ as pillow_version
from PIL import features

ROOT = Path('/src')
DEST = ROOT / 'internal/font'
SOURCE = Path('/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc')


def sha(data):
    return hashlib.sha256(data).hexdigest()


def generate():
    text = (ROOT / 'cmd/gin3/compact_settings.go').read_text()
    chars = ''.join(sorted(set(re.findall(r"'([^']+)': 0x", text) + [chr(i) for i in range(32, 127)])))
    face_index = 8
    images = {}
    for size in (12, 36):
        # Noto的實際墨跡會超出名義em，留字格內邊距並逐字檢查。
        face = ImageFont.truetype(str(SOURCE), 10 if size == 12 else 32, index=face_index)
        assert face.getname() == ('Noto Sans Mono CJK TC', 'Regular'), face.getname()
        atlas = Image.new('L', (16 * size, ((len(chars) + 15) // 16) * size))
        for i, char in enumerate(chars):
            width = size // 2 if ord(char) < 128 else size
            glyph = Image.new('L', (width, size))
            box = face.getbbox(char, anchor='ls')
            if char != ' ':
                ink = Image.new('L', (box[2] - box[0], box[3] - box[1]))
                ImageDraw.Draw(ink).text((-box[0], -box[1]), char, font=face, anchor='ls', fill=255)
                if ink.width > width or ink.height > size:
                    # ASCII豎線等少數超出em的字形完整縮入，絕不靠裁切。
                    ink = ink.resize((min(width, ink.width), min(size, ink.height)), Image.Resampling.LANCZOS)
                baseline = round(size * (.75 if ord(char) < 128 else .86))
                y = max(0, min(size - ink.height, baseline + box[1]))
                x = (width - ink.width) // 2
                assert x >= 0 and y >= 0 and x + ink.width <= width and y + ink.height <= size
                glyph.paste(ink, (x, y))
            assert glyph.getbbox() is not None or char == ' ', char
            atlas.paste(glyph, ((i % 16) * size, (i // 16) * size))
        buf = io.BytesIO()
        atlas.save(buf, format='PNG')
        images['settings_font_' + str(size) + '.png'] = buf.getvalue()
    images['settings_font.json'] = (json.dumps(dict(chars=chars, columns=16, sizes=[12, 36]), ensure_ascii=False) + '\n').encode()
    images['settings_font.LICENSE.txt'] = Path('/usr/share/doc/fonts-noto-cjk/copyright').read_bytes()
    manifest = dict(source=str(SOURCE), source_sha256=sha(SOURCE.read_bytes()), face_index=face_index,
                    family='Noto Sans Mono CJK TC Regular', license='SIL Open Font License 1.1',
                    glyph_count=len(chars), native_size=[12, 12, 6, 12], hd_size=[36, 36, 18, 36],
                    font_pixels=[10, 32], ink_bounds_checked=True,
                    pillow=pillow_version, freetype=features.version('freetype2'),
                    image='logh3-hd-preview:20261011-r1', generator='tools/scripts/settings_font.py',
                    generator_sha256=sha(Path(__file__).read_bytes()),
                    files={n: sha(data) for n, data in images.items()})
    images['settings_font.SOURCE.json'] = (json.dumps(manifest, ensure_ascii=False, indent=2) + '\n').encode()
    return images


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--check', action='store_true')
    p.add_argument('--refresh', action='store_true', help='更新本輪尚未凍結的產生物，檢查個別檔案擁有權')
    args = p.parse_args()
    assert DEST.is_dir() and DEST.stat().st_uid == 1000
    for name, data in generate().items():
        target = DEST / name
        if args.check:
            assert target.read_bytes() == data, name
        else:
            assert not target.exists() or args.refresh and target.stat().st_uid == 1000, '拒絕覆寫既有字型：' + name
            target.write_bytes(data)
    print('PASS 設定專用12／36字形、完整授權與來源雜湊' + ('獨立重生一致' if args.check else '已產生'))
