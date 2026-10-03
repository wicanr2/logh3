"""核對完整交戰原版雙跑、固定RAM的IDA bytes及祖先，拒絕改寫凍結收據。"""
import base64
import hashlib
import json
from pathlib import Path
import struct

ROOT = Path('/work')
files = {}


def read(path):
    p = ROOT / path
    b = p.read_bytes()
    assert p.stat().st_uid == 1000 and p.stat().st_gid == 1000, path
    files[path] = {'sha256': hashlib.sha256(b).hexdigest(), 'bytes': len(b)}
    return b


def word(b, p):
    return struct.unpack_from('<H', b, p)[0]


def far(b, p):
    return word(b, p) + 16 * word(b, p+2)


ds = 0x54b30
pairs = []
for case, expected in [('laser', (75, 12, 38746670)),
                       ('missile', (75, 12, 40346452)),
                       ('same-target', (124, 22, 82146748)),
                       ('death', (76, 13, 42946712))]:
    prefix = 'workplace/oracle/scene48-r69-v2-'
    a, b = prefix+case+'-a', prefix+case+'-b'
    ra, rb = [json.loads(read(d+'/report.json')) for d in (a, b)]
    assert ra['command'] == 'bash tools/fleet_scene_oracle.sh '+case+' a'
    assert rb['command'] == 'bash tools/fleet_scene_oracle.sh '+case+' b'
    normalized = [{k: v for k, v in r.items() if k != 'command'} for r in (ra, rb)]
    assert normalized[0] == normalized[1], case
    assert (len(ra['events']), len(ra['frames']), ra['end_step']-ra['start_step']) == expected
    assert ra['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
    assert not ra['dosgolem_dirty'] and ra['canvas'] == [640, 480]
    assert ra['irq0_every'] == 20000 and ra['irq0_pinned'] and not ra['cycle_clock']
    assert ra['new_inputs'] == [] and ra['uid'] == 1000
    assert hashlib.sha256(read('workplace/oracle/'+ra['parent'])).hexdigest() == ra['parent_sha256']
    names = ['start', 'end'] + [f['name'] for f in ra['frames']]
    for name in names:
        for suffix in ('-mem.bin', '-screen.bin', '-palette.json', '-screen.png'):
            ba, bb = [read(d+'/'+name+suffix) for d in (a, b)]
            assert ba == bb, (case, name, suffix)
            if suffix == '-mem.bin':
                assert len(ba) == 1048576
            if suffix == '-screen.bin':
                assert len(ba) == 307200
    by_loc = {}
    for e in ra['events']:
        by_loc.setdefault(e['original_location'], []).append(e)
    records = []
    for e in by_loc['3958:0041']:
        raw = list(base64.b64decode(e['comani_record']))
        assert len(raw) == 7
        records.append({'step': e['step'], 'record': raw, 'static': e['stack_words'][4],
                        'animate': e['stack_words'][5]})
    end = read(a+'/end-mem.bin')
    start = read(a+'/start-mem.bin')
    pool = far(start, ds+0xcfb4)
    target = pool + 102*78
    if case == 'death':
        assert ra['new_injections'] == [{'address': 'lin:687ce', 'step': 507999999,
            'before_hex': 'b80b', 'bytes_hex': '0100',
            'reason': '固定目標102為1艘，核對首擊死亡後第二命令跳過；不重擲seed'}]
        assert word(start, target+8) == 1 and word(end, target+8) == 0
        assert [e['stack_words'][2] for e in by_loc['1DD6:05D7']] == [82, 80]
        assert len(by_loc['1DD6:0C54']) == 1
        assert len(records) == 4 and by_loc['1DD6:0F6D'][0]['stack_words'][2] == 1
    else:
        assert ra['new_injections'] == []
    samples = []
    for f in ra['frames']:
        if f['original_location'] != '1DD6:0786' and f['original_location'] != '1DD6:0810':
            continue
        mem = read(a+'/'+f['name']+'-mem.bin')
        count = word(mem, ds+0xd7ec)
        assert count == 14
        rows = []
        for i in range(count):
            ptr = far(mem, ds+0xd78e+4*i)
            rows.append({'pointer_hex': hex(ptr), 'ship_hex': mem[ptr:ptr+78].hex(),
                         'projected_x': word(mem, ptr+0x3e), 'projected_y': word(mem, ptr+0x40),
                         'visible': word(mem, ptr+0x42), 'depth_word': word(mem, ptr+0x44)})
        for off, asset, size in [(0xf130, '3DBLOCK.SPR', 1536), (0xf134, 'COM_FLAG.SPR', 1440)]:
            raw = Path('/orig', asset).read_bytes()
            assert len(raw) == size and mem[far(mem, ds+off):far(mem, ds+off)+size] == raw
            files['org_game/extracted/'+asset] = {'sha256': hashlib.sha256(raw).hexdigest(), 'bytes': size}
        samples.append({'frame': f, 'cache': list(mem[ds+0xa456:ds+0xa458]),
                        'rotation': word(mem, ds+0xd7ea), 'pool': rows})
    pairs.append({'case': case, 'equal_files': 4*len(names), 'report': ra,
                  'comani_records': records, 'projected_samples': samples})

normal_path = 'workplace/oracle/normal79-r61-a/end-mem.bin'
normal = read(normal_path)
native_path = 'workplace/oracle/scene48-r69-v2-laser-a/frame-0786-01-mem.bin'
native = read(native_path)
ida = []
for segment in ('1DD6', '321F', '2376', '3958'):
    path = 'workplace/ida/scene48-r69-'+segment
    raw = read(path+'.com')
    export = json.loads(read(path+'.json'))
    assert export['input_sha256'] == hashlib.sha256(raw).hexdigest()
    assert export['kernel_version'] == '9.4' and export['uid'] == 1000
    source = native if segment in ('2376', '3958') else normal
    base = int(segment, 16)*16
    assert raw[256:] == source[base:base+65536]
    for row in export['instructions']:
        off = int(row['original_location'].split(':')[1], 16)
        data = bytes.fromhex(row['original_bytes_hex'])
        assert source[base+off:base+off+len(data)] == data
    ida.append({'export_path': path+'.json', 'ram_source': native_path if source is native else normal_path,
                'data': export})

ancestors = []
for name, sha in [('fleet-combat-command.json', 'e3b39582bfda07105d27837747d46bca19eaf2f1fb5d6ff383500870e4a6870d'),
                  ('combat-sequence.json', '199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8')]:
    path = 'docs/internals/data/'+name
    data = read(path)
    assert hashlib.sha256(data).hexdigest() == sha
    ancestors.append({'path': path, 'sha256': sha,
                      'injection_and_input_index': 'ancestry' if name.startswith('fleet') else 'pairs[].plans / ancestry'})

for path in ('tools/fleet_scene_oracle.sh', 'tools/scripts/fleet_scene_native.go',
             'tools/scripts/fleet_scene_ida.py', 'tools/scripts/fleet_scene_receipt.py'):
    read(path)

result = {'schema': 'logh3.fleet-scene.v1', 'date': '2026-10-04', 'status': '092 DRAFT',
          'remake_parent': '7e62f3d4cb39c14adc0b1ac699360147517112f4', 'remake_dirty': True,
          'dosgolem_head': pairs[0]['report']['dosgolem_head'], 'production_changes': [],
          'bundle_sha256': 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
          'images': read('workplace/round69-images.txt').decode(), 'ancestors': ancestors,
          'pairs': pairs, 'ida': ida, 'files': dict(sorted(files.items())),
          'limits': ['原版提交狀態續跑；祖先已受控，不能稱正常remake玩家路徑',
                     '未驗remake畫面；没有差分或原版parity聲明',
                     '正式命令、完整場景renderer、色盤、Game與保存待接',
                     '321F的IDA整區匯出與2376的partial helper匯出，不冒稱全部函式邊界已證實']}
out = ROOT/'docs/internals/data/fleet-scene.json'
data = (json.dumps(result, ensure_ascii=False, indent=2)+'\n').encode()
if out.exists():
    assert out.read_bytes() == data, '凍結收據改變；建立新證據版本，不覆寫'
else:
    assert out.parent.stat().st_uid == 1000 and out.parent.stat().st_gid == 1000
    out.write_bytes(data)
print('pairs', len(pairs), 'equal_files', sum(p['equal_files'] for p in pairs),
      'sources', len(files), 'instructions', sum(len(x['data']['instructions']) for x in ida),
      'sha256', hashlib.sha256(data).hexdigest())
