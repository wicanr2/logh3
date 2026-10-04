"""凍結#79子步8-4雙跑；用原版受控亂數輸出獨立驗算，拒絕覆寫。"""
from pathlib import Path
import base64
import collections
import hashlib
import json

def sha(b):
    return hashlib.sha256(b).hexdigest()

paths = [Path(f'workplace/promotion79-r84-native-v2-{s}.json') for s in 'ab']
raw = paths[0].read_bytes()
assert raw == paths[1].read_bytes(), '原版雙跑不一致'
oracle = json.loads(raw)
assert oracle['schema'] == 'logh3.battle-promotion-native.v1'
assert oracle['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
assert not oracle['dosgolem_dirty'] and oracle['uid'] == 1000
memory_path = Path('workplace/oracle/ownership79-r82-entry-a/mem-in.bin')
memory = memory_path.read_bytes()
assert sha(memory) == 'b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900'
need = [int.from_bytes(memory[0x54B30+0x4DBE+i*2:0x54B30+0x4DC0+i*2], 'little') for i in range(40)]

def rank(v):
    return sum(v >= n for n in [2, 4, 6, 8, 12, 16, 20, 24, 29, 34, 39])

def promote(record, randoms):
    a = bytearray(record)
    level = a[6]
    cost = need[min(level, 39)]
    experience = int.from_bytes(a[2:6], 'little')
    if level == 99 or experience < cost:
        assert not randoms
        return bytes(a), 0
    a[2:6] = (experience-cost).to_bytes(4, 'little')
    flags = 1 | (2 if rank((level+1)&255) != rank(level) else 0)
    a[6] = (level+1)&255
    cursor = 0
    count = 0
    def rnd():
        nonlocal cursor
        value = randoms[cursor]
        assert 0 <= value <= 32767
        cursor += 1
        return value
    def grow(off, lo, hi, bit):
        nonlocal flags, count
        value = a[off]
        if value >= 200:
            return
        span = lo if value < 100 else hi
        delta = (rnd() % (span+1) + rnd() % (span+1)) // 100
        if delta:
            a[off] = (value+delta)&255
            flags |= bit
            count += 1
    grow(0x0B, 300, 50, 4)
    if rnd() % 2:
        grow(0x13, 200, 100, 8)
    else:
        grow(0x12, 200, 100, 16)
    choice = rnd() % 3
    grow([0x0F, 0x10, 0x11][choice], 100, 50, [32, 64, 128][choice])
    if count < 3:
        if rnd() % 2:
            grow(0x0C, 200, 50, 256)
        else:
            grow(0x0D, 200, 50, 512)
    if count < 3:
        if rnd() % 2:
            grow(0x0E, 100, 50, 1024)
        else:
            grow(0x17, 200, 50, 2048)
    assert cursor == len(randoms), (cursor, len(randoms))
    return bytes(a), flags

sources = []
def source(path):
    p = Path(path)
    assert p.is_file() and p.stat().st_uid == 1000
    b = p.read_bytes()
    sources.append({'path': str(p), 'bytes': len(b), 'sha256': sha(b)})

for p in paths + [memory_path, Path('workplace/oracle/ownership79-r82-entry-a/entry.state'),
                   Path('workplace/promotion79-r84-native-v2.go')]:
    source(p)
for p in ['tools/scripts/scout_fuel_native_prepare.py', 'tools/scripts/scout_fuel_ida.py',
          'tools/scripts/battle_promotion_prepare.py', 'tools/scripts/battle_promotion_ida.py',
          'tools/scripts/battle_promotion_native_prepare.py', 'tools/scripts/battle_promotion_receipt.py',
          'tools/battle_promotion_native_run.sh', 'tools/battle_promotion_oracle.sh']:
    source(p)
for name in ['185C', '28E0', '297B']:
    source(f'workplace/ida/promotion79-r84-{name}.com')
for name in ['185C-v2', '28E0-v2', '28E0-tail', '297B']:
    source(f'workplace/ida/promotion79-r84-{name}.json')

returns = {'28E0:098D', '28E0:09A0', '28E0:070D', '28E0:078A', '28E0:0857', '28E0:08E2'}
cases = []
summary = collections.Counter()
for row in oracle['rows']:
    full = {name: bytearray.fromhex(value) for name, value in row['before'].items()}
    model = bytearray(full['admiral'])
    active = None
    promotions = []
    for call in row['calls']:
        where = call['location']
        if where == '28E0:05F9':
            assert active is None
            active = {'index': call['args'][0]&255, 'randoms': [], 'entry': call['step']}
        elif where in returns:
            assert active is not None
            active['randoms'].append(call['ax'])
        elif where in {'185C:065B', '185C:06A8'}:
            assert active is not None
            i = active['index']
            after, flags = promote(model[i*28:(i+1)*28], active['randoms'])
            assert flags == call['ax'], (row['name'], i, flags, call['ax'])
            model[i*28:(i+1)*28] = after
            active.update({'return': call['step'], 'flags': flags, 'admiral_after_hex': after.hex()})
            promotions.append(active)
            summary['attempts'] += 1
            summary['upgrades'] += flags != 0
            summary['random_outputs'] += len(active['randoms'])
            active = None
    assert active is None
    assert model == bytes.fromhex(row['after']['admiral']), row['name']
    tables = {}
    for name, before in full.items():
        after = bytes.fromhex(row['after'][name])
        assert len(before) == len(after)
        differences = [{'offset_hex': hex(i), 'before': v, 'after': after[i]} for i, v in enumerate(before) if v != after[i]]
        tables[name] = {'bytes': len(before), 'before_sha256': sha(before), 'after_sha256': sha(after), 'differences': differences}
        summary[name+'_changed_bytes'] += len(differences)
    frames = []
    for i, frame in enumerate(row['frames'] or []):
        prefix = f'workplace/promotion79-r84-{row["name"]}-frame{i}'
        hashes = {}
        for key, suffix in [('screen', 'screen.bin'), ('palette_rgb', 'palette.bin'), ('png', 'png')]:
            b = base64.b64decode(frame[key])
            p = Path(prefix+'.'+suffix)
            with p.open('xb') as f:
                f.write(b)
            source(p)
            hashes[key] = {'path': str(p), 'bytes': len(b), 'sha256': sha(b)}
        frames.append({'step': frame['step'], **hashes})
    summary['frames'] += len(frames)
    reports = [c for c in row['calls'] if c['location'] == '3826:0405']
    summary['reports'] += len(reports)
    cases.append({k: row[k] for k in ['name', 'index', 'case', 'need', 'injections', 'calls', 'inputs', 'seed_after_hex', 'steps']} |
                 {'tables': tables, 'promotions': promotions, 'frames': frames, 'grade': 'proven',
                  'comparison': '完整ADMIRAL模型符合；完整FLEET／SHIP保留原版差分，尚無獨立重算模型'})

instructions = []
for name in ['185C-v2', '28E0-v2', '28E0-tail', '297B']:
    data = json.loads(Path(f'workplace/ida/promotion79-r84-{name}.json').read_text())
    assert data['kernel_version'] == '9.4' and data['uid'] == 1000
    instructions.extend(data['instructions'])
formats = {}
for off in [0x3DE7, 0x3DEF, 0x3DFE, 0x3E0F, 0x3E1E, 0x3E2D, 0x3E3C, 0x3E4B, 0x3E5A, 0x3E69, 0x3E78, 0x3E87, 0x3E96]:
    start = 0x54B30+off
    end = memory.index(0, start)
    b = memory[start:end]
    formats[f'DS:{off:04X}'] = {'hex': b.hex(), 'big5': b.decode('big5'), 'grade': 'proven', 'source': '固定父RAM；185C:06E8的直接sprintf消費者'}
result = {'schema': 'logh3.battle-promotion.v1', 'grade': 'proven',
          'scope': '60個8-4入口診斷；原版雙跑相同。成長模型使用原版實際亂數輸出，不宣稱remake RNG／正常玩家／完整回合同態。',
          'oracle_identity': {k: v for k, v in oracle.items() if k != 'rows'},
          'original_json_sha256': sha(raw), 'need': need, 'rank_thresholds': [0, 2, 4, 6, 8, 12, 16, 20, 24, 29, 34, 39],
          'summary': dict(summary), 'formats': formats, 'cases': cases,
          'instructions': instructions, 'sources': sources}
dest = Path('docs/internals/data/battle-promotion.json')
with dest.open('x', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write('\n')
print('原版雙跑、獨立模型、完整表驗證通過', dict(summary))
print(dest, len(dest.read_bytes()), sha(dest.read_bytes()), 'sources', len(sources))
