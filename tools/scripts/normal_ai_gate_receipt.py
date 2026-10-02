"""第45輪#89完整核對；入口docs/re/383、046。只寫已索引的JSON及本機圖片。"""
from collections import Counter
import hashlib
import json
from pathlib import Path
import struct

from pngout import game_palette, write_png

def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()

oracle = Path('workplace/oracle')
assert sha('org_game/extracted/GIN3.COM') == '770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918'
names = ['loaded-mem.bin', 'ai-before.bin', 'ai-after.bin', 'entry-mem.bin',
         'screen.bin', 'screen.png', 'pal.bin']
runs = []
for mode in ('natural', 'zero'):
    a, b = [oracle / ('normal89-r45-' + mode + '-' + n) for n in ('a', 'b')]
    for n in names:
        assert (a / n).read_bytes() == (b / n).read_bytes(), (mode, n)
    for p in (a, b):
        assert 'dosgolem HEAD: f7523f9（未提交檔案 0 個）' in (p / 'meta.txt').read_text()
        flags = (p / 'flags.txt').read_text().splitlines()
        poke = flags[flags.index('-poke') + 1]
        assert poke == ('lin:5E00C@300000000=7E 16 DC D3' +
                        (';lin:5DA76@300000000=00' if mode == 'zero' else ''))
        runs.append({'path': str(p), 'classification': 'snapshot continuation; natural flag' if mode == 'natural' else 'controlled counterfactual; flag zero',
                     'command_argv': flags, 'hashes': {n: sha(p / n) for n in names + ['meta.txt', 'flags.txt', 'report.txt']}})

natural, zero = [oracle / ('normal89-r45-' + n + '-a') for n in ('natural', 'zero')]
a, b = [(p / 'loaded-mem.bin').read_bytes() for p in (natural, zero)]
assert [i for i in range(len(a)) if a[i] != b[i]] == [0x5DA76]
before, after = [(natural / n).read_bytes() for n in ('ai-before.bin', 'ai-after.bin')]
assert before[0x5DA76] == 1 and after[0x5DA76] == 0
assert before[0x5E00C:0x5E010] == after[0x5E00C:0x5E010] == bytes.fromhex('7e16dcd3')
tables = [('FLEET', 0x60BA4, 122, 32), ('SHIP', 0x668B2, 78, 352),
          ('STELLAR', 0x5F2CC, 126, 32), ('WORK', 0x5DA44, 230, 1)]
for _, base, stride, count in tables[:3]:
    assert before[base:base+stride*count] == after[base:base+stride*count]
for p, want in ((natural, [12, 16, 17, 28, 29, 30, 31]), (zero, [12, 28, 31])):
    m = (p / 'entry-mem.bin').read_bytes()
    count = struct.unpack_from('<H', m, 0x5E3C4)[0]
    assert list(m[0x5E3C8:0x5E3C8+count]) == want

remakes = [Path('workplace/round45-normal-fixed-' + n) for n in ('a', 'b')]
rnames = ['probe.json'] + [s + '/' + n for s in ('loaded', 'entry') for n in
    ('STELLAR.DT0', 'PLANET.DT0', 'ADMIRAL.DT0', 'FLEET.DT0', 'SHIP.DT0', 'WORK.DT0', 'SAVEDATA.IDX', 'screen.bin')]
for n in rnames:
    assert (remakes[0] / n).read_bytes() == (remakes[1] / n).read_bytes(), n
for n in ('STELLAR', 'PLANET', 'ADMIRAL', 'FLEET', 'SHIP', 'WORK'):
    assert (remakes[0] / 'entry' / (n+'.DT0')).read_bytes() == (Path('workplace/round45-normal-nosave/entry') / (n+'.DT0')).read_bytes()
assert (remakes[0] / 'entry/screen.bin').read_bytes() == Path('workplace/round45-normal-nosave/entry/screen.bin').read_bytes()
meta = json.loads((remakes[0] / 'probe.json').read_text())
assert meta['strategy_seed_observed_before'] == meta['strategy_seed_before'] == 'D3DC167E'
assert meta['strategy_seed_after'] == 'B6E4B78A'
m44 = (oracle / 'battle79-normal-r44-a/entry-mem.bin').read_bytes()
data_tables = {}
for name, base, stride, count in tables:
    length = stride * count
    original = m44[base:base+length]
    old = Path('workplace/round44-normal-fixed-a/entry', name+'.DT0').read_bytes()
    new = (remakes[0] / 'entry' / (name+'.DT0')).read_bytes()
    diffs = [i for i, (x, y) in enumerate(zip(original, new)) if x != y]
    data_tables[name] = {'bytes': length, 'old_different_bytes': sum(x!=y for x,y in zip(original,old)),
                         'different_bytes': len(diffs), 'offset_counts': dict(sorted(Counter(i%stride for i in diffs).items())),
                         'original_hex': original.hex(), 'remake_before_hex': old.hex(), 'remake_after_hex': new.hex()}
    if name != 'WORK':
        cold = (oracle / 'battle79-normal-r44-a/loaded-mem.bin').read_bytes()[base:base+length]
        assert a[base:base+length] == cold == (remakes[0] / 'loaded' / (name+'.DT0')).read_bytes()
f, s = [(remakes[0] / 'entry' / (name+'.DT0')).read_bytes() for name in ('FLEET', 'SHIP')]
assert [i for i in range(32) if f[i*122+4] != 255 and f[i*122+6] == 26] == [12,16,17,28,29,30,31]
assert sum(s[i*78+2] == 26 for i in range(352)) == 55
assert (remakes[0] / 'loaded/WORK.DT0').read_bytes()[0x32] == 1
assert (remakes[0] / 'entry/WORK.DT0').read_bytes()[0x32] == 0
prototype = Path('workplace/round45-normal-proto')
for n in ('STELLAR', 'PLANET', 'ADMIRAL', 'FLEET', 'SHIP', 'WORK'):
    assert (prototype / 'entry' / (n+'.DT0')).read_bytes() == (remakes[0] / 'entry' / (n+'.DT0')).read_bytes()

orig = (oracle / 'battle79-normal-r44-a/screen.bin').read_bytes()
new = (remakes[0] / 'entry/screen.bin').read_bytes()
pal = bytes(game_palette(str(oracle / 'battle79-normal-r44-a/screen.png')))
assert len(orig) == len(new) == 307200 and len(pal) == 48
write_png('workplace/round45-normal-remake-original-palette.png',640,480,new,pal)
write_png('workplace/round45-normal-entry-diff.png',640,480,bytes(int(x!=y) for x,y in zip(orig,new)),bytes([0,0,0,255,0,0]))
pair = b''.join(orig[y*640:(y+1)*640] + new[y*640:(y+1)*640] for y in range(480))
write_png('workplace/round45-normal-entry-pair.png',1280,480,pair,pal)
regions = {}
for name, rect in {'full': (0,0,640,480), 'field': (0,0,448,360)}.items():
    x0,y0,x1,y1 = rect
    offsets = [y*640+x for y in range(y0,y1) for x in range(x0,x1)]
    different = [i for i in offsets if orig[i] != new[i]]
    rgb = [i for i in different if pal[orig[i]*3:orig[i]*3+3] != pal[new[i]*3:new[i]*3+3]]
    regions[name] = {'rectangle':rect, 'compared_pixels':len(offsets), 'same_pixels':len(offsets)-len(different),
                     'different_index_pixels':len(different), 'different_rgb_pixels':len(rgb),
                     'maximum_channel_difference':max((abs(pal[orig[i]*3+k]-pal[new[i]*3+k]) for i in different for k in range(3)),default=0),
                     'difference_bbox_inclusive':[min(i%640 for i in different), min(i//640 for i in different), max(i%640 for i in different), max(i//640 for i in different)] if different else None}
ida = json.loads(Path('workplace/ida/normal89-r45-gate.json').read_text())
assert ida['uid'] == 1000 and ida['kernel_version'] == '9.4'
assert ida['input_sha256'] == 'e88b68361851a9fc8214b88f56ec352c44b6b83cc267c2c6195dbe2d87c8da1d'
assert hashlib.sha256(bytes(0x100) + a[0x2D100:0x2D760]).hexdigest() == ida['input_sha256']
assert [r['original_bytes_hex'] for r in ida['instructions'] if r['original_location'] in ('2D10:0005','2D10:000A','2D10:0030')] == ['803e468f00','7524','c606468f00']
sources = ['tools/normal_ai_gate_oracle.sh', 'tools/scripts/normal_ai_gate_ida.py', 'tools/scripts/normal_ai_gate_receipt.py',
           'tools/scripts/battle_normal_probe.go', 'tools/scripts/battle_normal_probe_seed.py',
           'internal/game/game.go', 'internal/game/normal_ai_gate_test.go', 'internal/game/battlesamples_test.go',
           'internal/game/rebellion.go', 'internal/bootstrap/load.go']
logs = ['workplace/round45-gate-tests.log', 'workplace/round45-shipped-tests.log', 'workplace/round45-standard-tests.log', 'workplace/round45-standard-retest.log']
data = {'schema':'logh3.normal-ai-once-gate.v1', 'grade':'proven gate and measured first-entry fleet restoration; complete entry parity unknown',
        'classification':'nearby-state for normal full-screen comparison; controlled flag-zero explicitly separate',
        'parent_head':'c05be3dd1afa8bfb745a6cda94ce02e46b6a178a', 'remake_worktree':'dirty; source hashes below',
        'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_dirty':False,
        'program_sha256':'770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918',
        'bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'toolchain': {'go_image':'golang:1.24-bookworm', 'go_image_id':'sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
                      'gui_image':'logh3-ebiten:2', 'gui_image_id':'sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8',
                      'ida_image':'ida-pro-9.4-idapython:locked-v1','ida_image_id':'sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'},
        'snapshot_parent': {'path':'/home/anr2/cht/dosgolem-logh3/workplace/fcmd/fcmd-s3-base.state',
                            'sha256':'0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7',
                            'metadata':'workplace/oracle/fcmd-s3-base/meta.txt', 'metadata_sha256':sha(oracle/'fcmd-s3-base/meta.txt'),
                            'reproducer':'tools/fleetcmd_receipts.sh; 299999999 cold S3 snapshot',
                            'loaded_table_check':'three full tables equal round44 cold-loaded RAM and current formal Game'},
        'original_runs':runs, 'original_double_equal_files_per_branch':len(names),
        'counterfactual_loaded_ram_different_offsets':['0x5DA76'],
        'natural_gate': {'flag_before':1,'flag_after':0,'enter_step':306103454,'clear_observed_step':306103457,'return_step':306103458,'rng_before_and_after':'D3DC167E', 'full_three_tables_unchanged':True},
        'zero_gate': {'flag_before':0,'branch_step':306103456,'builder_call_step':306388650,'decide_call_step':323879549,'return_step':324822153},
        'ida_export':ida, 'remake_metadata':meta,
        'remake_runs':[{'path':str(p), 'hashes':{n:sha(p/n) for n in rnames}} for p in remakes],
        'remake_double_equal_files':len(rnames), 'no_loaded_save_control':{n:sha(Path('workplace/round45-normal-nosave/entry')/n) for n in ['STELLAR.DT0','PLANET.DT0','ADMIRAL.DT0','FLEET.DT0','SHIP.DT0','WORK.DT0','screen.bin']},
        'disposable_prototype':{'entry_six_tables_equal_formal_product':True,
                                'hashes':{n:sha(n) for n in ['workplace/round45-normal-overlay.json','workplace/round45-normal-game.go','workplace/round45-normal-seed.go','workplace/round45-normal-proto/probe.json']}},
        'first_entry':{'fleets_before':[12,28,31], 'fleets_after':[12,16,17,28,29,30,31], 'original_fleets':[12,16,17,28,29,30,31], 'pool_count':55,'active_unit_count':48,'turn_original':380,'turn_remake':381,'final_seed_original_and_remake':'B6E4B78A'},
        'tables':data_tables, 'regions':regions,
        'images':{n:sha('workplace/'+n) for n in ['round45-normal-remake-original-palette.png','round45-normal-entry-diff.png','round45-normal-entry-pair.png']},
        'source_hashes':{n:sha(n) for n in sources}, 'test_logs':{n:sha(n) for n in logs},
        'limits':['global remaining bytes unclassified; do not claim complete table parity','date and seat selector still open','round45 branch end images not substituted for round44 cold normal receipt','no cross-sequence RNG guarantee','no original S1 or S2 full-entry parity claim']}
Path('docs/internals/data/normal-ai-once-gate.json').write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('JSON',sha('docs/internals/data/normal-ai-once-gate.json'))
print('images',regions)
