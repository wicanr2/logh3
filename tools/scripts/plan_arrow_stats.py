"""#80 對拍統計與差分圖；入口及契約：docs/playtest/025-plan-arrows.md。

tools/py.sh tools/scripts/plan_arrow_stats.py
只讀收據及正式 GUI 的未縮放輸出；不使用遮罩。輸出各批次 stats.json、diff.png。
"""
import hashlib
import json
import pathlib
import struct
import zlib

from png import write_rgb


def png_rgb(path):
    data = path.read_bytes()
    assert data[:8] == b'\x89PNG\r\n\x1a\n'
    chunks, pos = {}, 8
    while pos < len(data):
        size = struct.unpack_from('>I', data, pos)[0]
        key = data[pos + 4:pos + 8]
        chunks[key] = chunks.get(key, b'') + data[pos + 8:pos + 8 + size]
        pos += size + 12
    w, h, depth, kind, compression, filtering, interlace = struct.unpack('>IIBBBBB', chunks[b'IHDR'])
    assert (w, h, compression, filtering, interlace) == (640, 480, 0, 0, 0)
    assert depth == 8 or (kind == 3 and depth == 4)
    channels = {2: 3, 3: 1, 6: 4}[kind]
    stride = (w * channels * depth + 7) // 8
    raw, previous, pixels, pos = zlib.decompress(chunks[b'IDAT']), bytearray(stride), [], 0
    for _ in range(h):
        method = raw[pos]
        row = bytearray(raw[pos + 1:pos + 1 + stride])
        pos += stride + 1
        assert method in range(5)
        for x in range(stride):
            a = row[x - channels] if x >= channels else 0
            b = previous[x]
            c = previous[x - channels] if x >= channels else 0
            p = a + b - c
            pa, pb, pc = abs(p - a), abs(p - b), abs(p - c)
            predictor = a if pa <= pb and pa <= pc else b if pb <= pc else c
            row[x] = (row[x] + (0, a, b, (a + b) // 2, predictor)[method]) & 255
        for x in range(w):
            if kind == 3:
                value = row[x] if depth == 8 else (row[x // 2] >> (4 if x % 2 == 0 else 0)) & 15
                index = value * 3
                pixels.append(tuple(chunks[b'PLTE'][index:index + 3]))
            else:
                pixels.append(tuple(row[x * channels:x * channels + 3]))
                if kind == 6:
                    assert row[x * channels + 3] == 255
        previous = row
    return pixels


def stats(a, b, indices):
    bad = [i for i in indices if a[i] != b[i]]
    return {'pixels': len(indices), 'equal': len(indices) - len(bad), 'different': len(bad),
            'max_channel_difference': max((max(abs(c - d) for c, d in zip(a[i], b[i])) for i in indices), default=0),
            'bounding_box': [min(i % 640 for i in bad), min(i // 640 for i in bad),
                             max(i % 640 for i in bad), max(i // 640 for i in bad)] if bad else None}


regions = {'full': (0, 0, 640, 480), 'panel': (128, 60, 384, 340),
           'tax_up': (464, 280, 16, 20), 'tax_down': (480, 280, 16, 20),
           'pct_up': (464, 320, 16, 20), 'pct_down': (480, 320, 16, 20)}
for batch in ('tax-up', 'tax-down', 'pct-up', 'pct-down', 'hover', 'moved', 'updown', 'reopen'):
    directory = pathlib.Path('workplace/oracle/planarrow-' + batch)
    original, remake = png_rgb(directory / 'screen.png'), png_rgb(directory / 'remake.png')
    a, b = (directory / 'screen.bin').read_bytes(), (directory / 'remake.bin').read_bytes()
    assert len(a) == len(b) == 640 * 480
    report = {'comparison': 'same-state', 'canvas': [640, 480], 'crop': None, 'scale': 1,
              'hashes': {name: hashlib.sha256((directory / name).read_bytes()).hexdigest()
                         for name in ('screen.bin', 'screen.png', 'pal.bin', 'remake.bin', 'remake.png',
                                      'flags.txt', 'inputs.json', 'remake-script.txt')},
              'index_different': sum(x != y for x, y in zip(a, b)), 'regions': {}}
    for key, (x, y, w, h) in regions.items():
        indices = [j * 640 + i for j in range(y, y + h) for i in range(x, x + w)]
        report['regions'][key] = {'rectangle': [x, y, w, h], **stats(original, remake, indices)}
    rows = [bytes(c for i in range(y * 640, (y + 1) * 640)
                  for c in ((255, 0, 0) if original[i] != remake[i] else (0, 0, 0))) for y in range(480)]
    write_rgb(directory / 'diff.png', 640, 480, rows)
    report['hashes']['diff.png'] = hashlib.sha256((directory / 'diff.png').read_bytes()).hexdigest()
    (directory / 'stats.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(directory.name, report['regions']['full'])
    assert report['index_different'] == report['regions']['full']['different'] == 0
