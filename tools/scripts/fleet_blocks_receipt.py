"""核雙跑、原始 bytes，保存可離線驗證的整段下方場景；拒絕不同內容覆寫。"""
import base64
import hashlib
import json
import struct
import zlib
from pathlib import Path

work = Path('/work')
oracle = work/'workplace/oracle'
files = {}


def read(p):
    b = p.read_bytes()
    files[str(p.relative_to(work))] = hashlib.sha256(b).hexdigest()
    assert p.stat().st_uid == 1000 and p.stat().st_gid == 1000
    return b


def pack(b):
    return base64.b64encode(zlib.compress(b, 9)).decode('ascii')


def pair(kind):
    a = oracle/f'blocks48-r70-{kind}a'
    b = oracle/f'blocks48-r70-{kind}b'
    ra = json.loads(read(a/'report.json'))
    rb = json.loads(read(b/'report.json'))
    ca, cb = dict(ra), dict(rb)
    ca.pop('command')
    cb.pop('command')
    assert ca == cb
    names = sorted(x.name for x in a.iterdir() if x.suffix in ('.bin', '.png', '.json') and x.name != 'report.json')
    for name in names:
        assert read(a/name) == read(b/name), name
    return a, ra, names


def png_indices(b):
    assert b[:8] == b'\x89PNG\r\n\x1a\n'
    off, compressed, pal = 8, b'', []
    while off < len(b):
        n = struct.unpack('>I', b[off:off+4])[0]
        kind, data = b[off+4:off+8], b[off+8:off+8+n]
        if kind == b'IHDR':
            w, h, depth, color, _, _, interlace = struct.unpack('>IIBBBBB', data)
            assert (w, h, depth, color, interlace) == (640, 480, 8, 3, 0)
        if kind == b'PLTE':
            pal = [list(data[i:i+3]) for i in range(0, len(data), 3)]
        if kind == b'IDAT':
            compressed += data
        off += n+12
    raw = zlib.decompress(compressed)
    prev = bytearray(w)
    pixels = bytearray()
    for y in range(h):
        method = raw[y*(w+1)]
        row = bytearray(raw[y*(w+1)+1:(y+1)*(w+1)])
        for x in range(w):
            left, up, corner = row[x-1] if x else 0, prev[x], prev[x-1] if x else 0
            if method == 1:
                row[x] = (row[x]+left) & 255
            elif method == 2:
                row[x] = (row[x]+up) & 255
            elif method == 3:
                row[x] = (row[x]+(left+up)//2) & 255
            elif method == 4:
                p = left+up-corner
                d = (abs(p-left), abs(p-up), abs(p-corner))
                value = (left, up, corner)[d.index(min(d))]
                row[x] = (row[x]+value) & 255
            else:
                assert method == 0
        pixels.extend(row)
        prev = row
    return pixels, pal


a, report, names = pair('')
edge, edge_report, edge_names = pair('edges-valid-')
regs = []
for side in ('a', 'b'):
    regs.append(json.loads(read(oracle/f'blocks48-r70-registers-{side}.json')))
ra, rb = dict(regs[0]), dict(regs[1])
ra.pop('command')
rb.pop('command')
assert ra == rb
assert len(report['frames']) == 40 and len(edge_report['cases']) == 20
assert len(names) == 164 and len(edge_names) == 20
assert hashlib.sha256(read(a/'scene-01.state')).hexdigest() == '07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef'
# state serialization itself is not claimed identical across runs.
source_ram = read(a/'scene-01-01-entry-mem.bin')
sprites = {}
frames = []
cursor = 0
logical_palette = [None]*16
for f in report['frames']:
    raw = read(a/(f['name']+'-screen.bin'))
    pixels, pal = png_indices(read(a/(f['name']+'-screen.png')))
    for logical, dac in zip(raw, pixels):
        if logical_palette[logical] is not None:
            assert logical_palette[logical] == pal[dac]
        logical_palette[logical] = pal[dac]
    ops = []
    while cursor < len(report['events']) and report['events'][cursor]['step'] < f['step']:
        ev = report['events'][cursor]
        cursor += 1
        if ev['original_location'] == '2376:0AA5':
            ops.append({'kind': 'background'})
        elif ev['original_location'] == '2376:07FA':
            x, y, wb, h, count, off, seg = ev['stack_words'][2:9]
            key = f'{seg:04X}:{off:04X}:{wb}:{h}'
            sprites[key] = {'width_bytes': wb, 'height': h, 'raw_hex': source_ram[seg*16+off:seg*16+off+wb*h*4].hex()}
            ops.append({'kind': 'sprite', 'x': x, 'y': y, 'count': count, 'sprite': key})
        elif ev['original_location'] == '3870:07C9':
            ops.append({'kind': 'line', 'args': ev['stack_words'][2:8]})
    frames.append({'metadata': f, 'operations': ops, 'native_zlib_base64': pack(raw)})
assert all(x is not None for x in logical_palette)
cases = []
for c in edge_report['cases']:
    x, y, wb, h, count, off, seg = c['stack_words'][2:9]
    key = f'{seg:04X}:{off:04X}:{wb}:{h}'
    sprites[key] = {'width_bytes': wb, 'height': h, 'raw_hex': source_ram[seg*16+off:seg*16+off+wb*h*4].hex()}
    cases.append({'metadata': c, 'operation': {'kind': 'sprite', 'x': x, 'y': y, 'count': count, 'sprite': key},
                  'native_zlib_base64': pack(read(edge/(c['name']+'-screen.bin')))})
exports = []
for name, ram_path in (
        ('2376-composition', oracle/'scene48-r69-v2-laser-a/frame-0786-01-mem.bin'),
        ('3870-linear', a/'scene-01-01-entry-mem.bin'),
        ('39d7-callee', a/'scene-01-01-entry-mem.bin'),
        ('39d7-body', a/'scene-01-01-entry-mem.bin'),
        ('2376-lineconfig', oracle/'scene48-r69-v2-laser-a/frame-0786-01-mem.bin')):
    p = work/'workplace/ida'/f'blocks48-r70-{name}.json'
    r = json.loads(read(p))
    data = read(p.with_suffix('.com'))
    assert hashlib.sha256(data).hexdigest() == r['input_sha256']
    ram = read(ram_path)
    segment = int(r['runtime_segment'], 16)
    assert data[256:] == ram[segment*16:segment*16+65536]
    r['ram_source'] = str(ram_path.relative_to(work))
    r['ram_sha256'] = hashlib.sha256(ram).hexdigest()
    for row in r['instructions']:
        cs, ip = [int(x, 16) for x in row['original_location'].split(':')]
        expect = bytes.fromhex(row['original_bytes_hex'])
        assert data[256+ip:256+ip+len(expect)] == expect
        if row['original_location'] == '2376:0B43':
            row.update(grade='proven', semantic='單一retf，沒有背景搬移', source='固定bytes及40層重播')
        elif row['original_location'] in ('2376:0A88', '39D7:0849', '2376:17CB'):
            row.update(grade='proven', semantic={'2376:0A88': '堆疊最後寫平面遮罩8', '39D7:0849': '垂直線沿用平面遮罩', '2376:17CB': '線段VGA設定不改序列器遮罩'}[row['original_location']], source='原bytes及雙跑56暫存器邊界')
        if row['grade'] == 'proven':
            row['warning'] = ''
    exports.append(r)
asset_inputs = {}
for name, offset in (('3DBLOCK.SPR', 21790), ('COM_FLAG.SPR', 23326)):
    original = (Path('/orig')/name).read_bytes()
    assert source_ram[0x3d220+offset:0x3d220+offset+len(original)] == original
    asset_inputs[name] = {'size': len(original), 'sha256': hashlib.sha256(original).hexdigest()}
failures = []
for name in ('edges-a', 'edges-reviewed-a'):
    p = oracle/f'blocks48-r70-{name}'
    failures.append(json.loads(read(p/'aborted.json')))
    assert hashlib.sha256(read(p/'aborted-source.go')).hexdigest() == failures[-1]['source_sha256']
out = {'schema': 'logh3.fleet-blocks.v1', 'date': '2026-10-04',
       'classification': 'same-state indexed compositor fixture; original cached coordinates, not remake Game/player parity',
       'remake_parent': 'ee5561916311c8445a2434fa6d831bf9d88a73ed',
       'dosgolem_head': report['dosgolem_head'], 'report': report, 'registers': regs,
       'logical_palette_rgb8': logical_palette, 'asset_inputs': asset_inputs, 'sprites': sprites, 'frames': frames,
       'edge_report': edge_report, 'edges': cases, 'ida': exports, 'aborted_diagnostics': failures,
       'limits': ['完整092仍DRAFT，Game未接', 'zero-height直接注入未返回；正式可達性未知，不默認修補',
                  '原生logical色號，PNG以AC映射DAC；不是目前Game色盤通過', '固定serialized state只使用a，未宣稱state封包雙跑相同'],
       'files': files}
blob = (json.dumps(out, ensure_ascii=False, indent=2)+'\n').encode()
target = work/'docs/internals/data/fleet-blocks.json'
assert target.parent.stat().st_uid == 1000 and target.parent.stat().st_gid == 1000
if target.exists():
    assert target.read_bytes() == blob, 'refuse differing overwrite'
else:
    target.write_bytes(blob)
print(len(blob), 'bytes', len(files), 'sources', hashlib.sha256(blob).hexdigest())
