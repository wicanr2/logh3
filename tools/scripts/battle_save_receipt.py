"""第136輪：原版公開存讀與145保存返回的可回查索引。"""
import argparse
import hashlib
import json
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--freeze', action='store_true')
a = p.parse_args()
root, work = Path('/ref'), Path('workplace/round136')
index = Path('docs/internals/data/battle-save-return.json')
def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()
def read(path):
    return json.loads(path.read_text())
files = {}
def include(path):
    files[str(path)] = dict(size=path.stat().st_size, sha256=digest(path))

native = []
for first, second, count in [('c', 'd', 37), ('e', 'f', 45)]:
    left = Path('workplace/oracle/battle-load-r136-'+first)
    right = Path('workplace/oracle/battle-load-r136-'+second)
    paths = sorted(p for p in left.rglob('*') if p.is_file())
    assert len(paths) == count
    for src in paths:
        dst = right/src.relative_to(left)
        assert src.read_bytes() == dst.read_bytes(), (src, dst)
        include(src)
        include(dst)
    r = read(left/'report.json')
    assert r['complete'] and r['dosgolem_head'] == 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'
    assert not r['memory_injection'] and not r['seed_injection']
    assert digest(left/'observer-source.go') == r['observer_source_sha256']
    cold = next(x for x in r['rows'] if x['name'] == '06-cold-load-direct-battle')
    assert cold['step'] == 214146187 and cold['location'] == '12AA:02D4'
    assert cold['control']['active'] == 1 and cold['control']['region'] == 26 and cold['control']['battle_turn_zero'] == 1
    # Entry precedes the scheduled faction click: no faction reselection.
    if first == 'c':
        assert not any(x['step'] == 240000000 for x in r['inputs'])
    if first == 'e':
        assert r['rows'][-1]['name'] == '08-cold-load-next-turn'
        assert r['rows'][-1]['control']['battle_turn_zero'] == 2
    native.append(dict(pair=[first, second], identical_files=count, rows=r['rows'], inputs=r['inputs']))

source = Path('workplace/oracle/battle-load-r136-e')
memory = (source/'06-cold-load-direct-battle-mem.bin').read_bytes()
tables = {}
for name, base, size in [('stellar', 0x5f2cc, 4032), ('planet', 0x53132, 6642),
                         ('admiral', 0x602ac, 2296), ('fleet', 0x60ba4, 3904),
                         ('ship', 0x668b2, 27456), ('work', 0x5da44, 230)]:
    b = (source/'scratch'/(name+'.dt0')).read_bytes()
    assert len(b) == size
    dif = [dict(offset=i, before=v, after=memory[base+i]) for i, v in enumerate(b) if v != memory[base+i]]
    tables[name] = dict(linear_base=hex(base), size=size, changes=dif)
assert [x['offset'] for x in tables['fleet']['changes']] == [1570]
assert [x['offset'] for x in tables['ship']['changes']] == [6287, 6443, 6521, 6599, 6677, 6755, 6833, 6911]
assert [x['offset'] for x in tables['work']['changes']] == [42, 43, 44, 45]
assert all(not tables[n]['changes'] for n in ['stellar', 'planet', 'admiral'])

gui = work/'gui-save-c'
result = read(gui/'RESULT.json')
assert result['passed'] and result['normal_entry']
assert not any(result[k] for k in ['state_injection', 'seed_injection', 'passive_override', 'script_flag'])
assert digest(work/'gin3') == result['binary_sha256']
images = sorted(gui.glob('*.png'))
assert len(images) == 8
for path in [*images, gui/'RESULT.json', gui/'FRAMES.jsonl', gui/'game.log', work/'gin3', *sorted((gui/'save').glob('*'))]:
    if path.is_file():
        include(path)
positive = (work/'save-test-final.log').read_text()
negative = (work/'save-negative.log').read_text()
assert '--- PASS: TestBattleSaveSuccessReturnsWithoutAdvancing' in positive
assert '--- FAIL: TestBattleSaveSuccessReturnsWithoutAdvancing' in negative and 'StateSaveGame <nil>' in negative
standard = (work/'standard.log').read_text().splitlines()
passed, empty = sum(x.startswith('ok ') for x in standard), sum('[no test files]' in x for x in standard)
assert (passed, empty) == (15, 4) and not any(x.startswith('FAIL') for x in standard)
for name in ['save-test-final.log', 'save-negative.log', 'standard.log', 'build.log', 'gui-save-c.log']:
    include(work/name)
sources = {}
for name in ['internal/game/game.go', 'internal/game/battle.go', 'internal/game/battlesave_test.go',
             'tools/scripts/battle_load_native.go', 'tools/battle_load_oracle.sh', 'tools/battle_load_native_run.sh',
             'tools/scripts/battle_save_gui.py', 'tools/scripts/battle_save_receipt.py',
             'tools/scripts/landing_player_plan.go', 'tools/scripts/landing_player_gui.py',
             'tools/scripts/fortress_presentation_gui.py']:
    sources[name] = digest(root/name)
document = dict(schema='logh3.battle-save-return.v1', parent_commit='e402794d',
                dosgolem_head='f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
                native=native, cold_load_tables=tables, gui=dict(path=str(gui), pictures=len(images)),
                standard=dict(passed=passed, no_tests=empty), sources_sha256=sources, files=files,
                limits=['save success return only implemented', 'cold battle reconstruction pending',
                        'no complete DOS/remake engagement same-state claim', 'no new audio parity claim'])
if a.freeze:
    assert not index.exists(), 'refuse overwrite'
    index.write_text(json.dumps(document, ensure_ascii=False, indent=2)+'\n')
else:
    assert read(index) == document, 'receipt or source changed'
print('PASS', len(files), 'sources', len(sources), 'native pairs 82; GUI 8; standard 15/4')
print('index SHA-256', digest(index), 'bytes', index.stat().st_size)
