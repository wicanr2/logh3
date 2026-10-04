"""#79 子步8-3：核完整雙跑、原池增量、等待時點與IDA原始指令。"""
import base64
import hashlib
import json
import os
from pathlib import Path

sources = {}


def read(path):
    b = Path(path).read_bytes()
    sources[path] = {'sha256': hashlib.sha256(b).hexdigest(), 'size': len(b)}
    return b


def write(path, b):
    p = Path(path)
    assert p.parent.stat().st_uid == os.getuid()
    with p.open('xb') as f:
        f.write(b)
    read(path)


a = read('workplace/scout79-r83-native-v7-a.json')
assert a == read('workplace/scout79-r83-native-v7-b.json')
d = json.loads(a)
assert len(d['rows']) == 281 and d['uid'] == os.getuid()
assert d['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert d['dosgolem_dirty'] is False and d['tick'] == 20000
parent = read('workplace/oracle/ownership79-r82-entry-a/entry.state')
assert hashlib.sha256(parent).hexdigest() == d['parent_state_sha256']
ram = read('workplace/oracle/ownership79-r82-entry-a/mem-in.bin')
assert len(ram) == 1048576 and hashlib.sha256(ram).hexdigest() == d['memory_sha256']
for name in ('flags.txt', 'meta.txt', 'report.txt'):
    read('workplace/oracle/ownership79-r82-entry-a/' + name)
cases = []
reports = recovered = 0
frames = []
for r in d['rows']:
    f = bytearray.fromhex(r['before_fleet_hex'])
    s = bytearray.fromhex(r['before_ship_hex'])
    assert len(f) == 122 and len(s) == 78
    expiry = False
    if r['fleet_count']:
        if f[0x3A] == 255:
            if f[0x3B]:
                f[0x3B] -= 1
        else:
            f[0x3B] = (f[0x3B] + 1) & 255
            expiry = f[0x3B] >= 4
            if expiry:
                f[0x3A:0x3C] = bytes([255, 3])
                s[0x19] &= 0x3F  # 本組明示C5；1727擦bit6後185C清bit7。
                s[0x1E] = 255
                s[0x28:0x2C] = s[0x34:0x38] = f[0x5C:0x60]
    assert f.hex() == r['after_fleet_hex'], r['name']
    assert s.hex() == r['after_ship_hex'], r['name']
    assert r['seed_after_hex'] == '78563412'
    calls = r['calls'] or []
    shown = [c for c in calls if c['entry'] == '3826:0405']
    human = bool(r['human_mask'] & (1 << f[1]))
    assert len(shown) == int(expiry and human)
    expected = ['185C:05A9']
    if human:
        expected += ['3826:0405', '3765:006C']
    expected += ['25DB:0535', '25DB:0759', '1727:0298', '25DB:0714', '25DB:0759']
    assert [c['entry'] for c in calls] == (expected if expiry else []), r['name']
    for c in calls:
        cf = bytes.fromhex(c['fleet_hex'])
        assert cf[0x3A] == r['mode'] and cf[0x3B] == ((r['fuel'] + 1) & 255)
    for c in shown:
        assert c['args'][2] == 65535 and c['args'][3] & 255 == f[4]
        assert c['ship_hex'] == r['before_ship_hex']
    assert len(r['inputs'] or []) == 2 * len(shown)
    assert len(r['frames'] or []) == len(shown)
    for im in r['frames'] or []:
        pix = base64.b64decode(im['screen'])
        rgb = base64.b64decode(im['palette_rgb'])
        assert len(pix) == 640*480 and max(pix) < 16 and len(rgb) == 48
        frames.append({'case': r['name'], 'step': im['step'],
                       'screen_sha256': hashlib.sha256(pix).hexdigest(),
                       'palette_rgb_sha256': hashlib.sha256(rgb).hexdigest(),
                       'png_sha256': hashlib.sha256(base64.b64decode(im['png'])).hexdigest()})
        if r['mode'] == 0 and r['fuel'] == 3 and r['human_mask'] == 3:
            prefix = f"workplace/scout79-r83-fleet{r['fleet']}-original"
            for suffix, key in (('screen.bin', 'screen'), ('palette.bin', 'palette_rgb'), ('png', 'png')):
                write(prefix + '.' + suffix, base64.b64decode(im[key]))
    reports += len(shown)
    recovered += int(expiry)
    cases.append({k: v for k, v in r.items() if k != 'frames'})
assert reports == 48 and recovered == 96 and len(frames) == 48

ida = []
for segment, name in ((0x185C, '185C'), (0x3982, '3982-v2'), (0x3765, '3765-v2')):
    p = f'workplace/ida/scout79-r83-{name}-ida.json' if name == '185C' else f'workplace/ida/scout79-r83-{name.split("-")[0]}-ida-v2.json'
    data = json.loads(read(p))
    binary = read(f'workplace/ida/scout79-r83-{segment:04X}.com')
    assert data['kernel_version'] == '9.4' and data['uid'] == os.getuid()
    assert data['input_sha256'] == hashlib.sha256(binary).hexdigest()
    assert binary[256:] == ram[segment*16:segment*16+4096]
    for insn in data['instructions']:
        off = int(insn['original_location'].split(':')[1], 16)
        b = bytes.fromhex(insn['original_bytes_hex'])
        assert binary[256+off:256+off+len(b)] == b
    ida.append(data)
semantics = [
    ('185C:04A8', '3a069498', '依DS9894長度掃DS9898戰場艦隊清單'),
    ('185C:04D2', '263804', '索敵模式byte等FF走冷卻分支'),
    ('185C:04E3', '26fe07', '派出燃料先byte加一，255回繞0'),
    ('185C:04F2', '7303', '加一後unsigned燃料至少4才回收'),
    ('185C:04FB', '9aa9055c18', '回收前先等待人類報告'),
    ('185C:0506', '268b4738', '索敵原池讀FLEET+38'),
    ('185C:050F', '8b16b6cf', 'SHIP遠指標segment讀DSCFB6'),
    ('185C:0513', '8b3eb4cf', 'SHIP遠指標offset讀DSCFB4'),
    ('185C:0521', '9a3505db25', '先擦戰場索敵圖示'),
    ('185C:052C', '9a98022717', '再擦原小地圖標記'),
    ('185C:0536', '268067197f', '回收清SHIP+19 bit7，其他bit依擦點端保留'),
    ('185C:0545', '26894728', 'TargetX寫FLEET+5C'),
    ('185C:0549', '26894734', 'CurrentX寫FLEET+5C'),
    ('185C:0557', '2689472a', 'TargetY寫FLEET+5E'),
    ('185C:055B', '26894736', 'CurrentY寫FLEET+5E'),
    ('185C:0572', '26c6471eff', '重畫後Action整byte寫FF'),
    ('185C:057A', '26c604ff', '回收後Mode寫FF'),
    ('185C:0584', '26c60703', '回收後冷卻值寫3'),
    ('185C:0596', '26fe0f', '未派出且非零冷卻byte減一'),
    ('185C:05C3', '7540', '電腦座位不顯示報告'),
    ('185C:05DB', 'bbc83d', '艦隊標頭追加DS3DC8失去連絡原文'),
    ('185C:05F2', '26ff7704', '報告第三參數為FLEET+4原word'),
    ('185C:05FE', '9a05042638', 'FFFF分支使用提督低byte畫頭像，沿3826既有證據'),
    ('3765:0091', '89ec', '等待事件讀取返回點；本探針於此送左键，並保存實際步數'),
]
lookup = {x['original_location']: x for data in ida for x in data['instructions']}
for loc, sig, meaning in semantics:
    insn = lookup[loc]
    assert insn['original_bytes_hex'] == sig
    insn.update({'semantic': meaning, 'grade': 'proven', 'warning': None,
                 'source': '原始bytes／caller／281案完整保存雙跑；範圍見研究417'})
formats = {key: ram[0x54B30+off:0x54B30+off+100].split(b'\0')[0].hex()
           for key, off in [('battle_scout_prefix', 0x8EA6), ('battle_scout_admiral', 0x8EAB), ('battle_scout_tail', 0x3DC8)]}
for p in ('tools/scripts/scout_fuel_probe.go', 'tools/scripts/scout_fuel_native_prepare.py',
          'tools/scripts/scout_fuel_ida.py', 'tools/scout_fuel_native_run.sh', 'tools/scout_fuel_oracle.sh',
          'tools/scripts/scout_fuel_receipt.py', 'workplace/scout79-r83-native-v7.go',
          'docs/internals/data/battle-ownership-v2.json'):
    read(p)
result = {'schema': 'logh3.scout-fuel.v1', 'parent_commit': 'e57f18293421d99797f0ef0f743ab7ff8dc0f4d9',
          'dosgolem_head': d['dosgolem_head'], 'program_sha256': d['program_sha256'],
          'scope': d['setup'], 'cases': cases, 'reports': reports, 'recovered': recovered,
          'frames': frames, 'report_formats': formats, 'ida': ida, 'sources': sources}
out = 'docs/internals/data/scout-fuel.json'
write(out, (json.dumps(result, ensure_ascii=False, indent=2) + '\n').encode())
print('281案／48報告／96回收／24原指令；SHA-256', hashlib.sha256(Path(out).read_bytes()).hexdigest())
