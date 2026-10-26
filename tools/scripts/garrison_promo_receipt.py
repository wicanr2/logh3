"""146獨立交付核對：凍結來源、Windows包、正常GUI、六表、影片與連續音訊。Docker內執行。"""
import hashlib
import json
import os
from pathlib import Path
import struct
import zipfile

ROOT = Path('/src')
VERSION = 'v.1.5.2-20261024'
OUT = ROOT/'dist-all'/VERSION


def read(p): return json.loads(Path(p).read_text())
def sha(p):
    h = hashlib.sha256()
    with Path(p).open('rb') as f:
        for chunk in iter(lambda: f.read(1024*1024), b''): h.update(chunk)
    return h.hexdigest()


build = read(OUT/'BUILD.json')
assert build['version'] == VERSION
archive = OUT/build['archive']['path']
assert sha(archive) == build['archive']['sha256'] and archive.stat().st_size == build['archive']['bytes']
frozen = ROOT/'workplace'/('personnel-patch-'+VERSION)/'source'
product_count = 0
for name, digest in build['source_files'].items():
    assert sha(frozen/name) == digest, name
    if not name.startswith('vendor/'):
        assert sha(ROOT/name) == digest, ('封包與現行產品不同', name)
        product_count += 1
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    prefix = 'LOGH3-'+VERSION+'-windows-amd64-patch/'
    manifest = json.loads(z.read(prefix+'MANIFEST.json'))
    assert manifest['version'] == VERSION
    for name, row in manifest['files'].items():
        b = z.read(prefix+name)
        assert hashlib.sha256(b).hexdigest() == row['sha256'] and len(b) == row['bytes']
        assert name in ['gin3.exe', 'LICENSE', 'README.txt'] or name.startswith('licenses/')
    assert {i.filename for i in z.infolist()} == {prefix+n for n in manifest['files']} | {prefix+'MANIFEST.json'}
    text = z.read(prefix+'README.txt'); assert text.startswith(b'\xef\xbb\xbf')
    assert b'\n' not in text.replace(b'\r\n', b''); text.decode('utf-8-sig')
    binary = z.read(prefix+'gin3.exe'); assert binary[:2] == b'MZ'
    pe = struct.unpack_from('<I', binary, 0x3c)[0]
    assert binary[pe:pe+4] == b'PE\0\0' and struct.unpack_from('<H', binary, pe+4)[0] == 0x8664
win_hash = hashlib.sha256(binary).hexdigest()
assert sha(ROOT/'workplace/music69-windows/garrison146/gin3.exe') == win_hash
version_path = ROOT/'workplace/music69-windows/garrison146/version.txt'
assert version_path.read_text().strip() == VERSION
plan_path = ROOT/'workplace/garrison146-plan-final/PLAN.json'
model = read(plan_path)
gui = []
for folder in [ROOT/'workplace/garrison146-recording-r2', ROOT/'workplace/music69-windows/garrison146/verified-r2']:
    value = read(folder/'RESULT.json')
    assert value['passed'] and value['normal_entry']
    assert not any(value[k] for k in ['state_injection', 'seed_injection', 'script_flag'])
    assert value['plan_sha256'] == sha(plan_path)
    if 'wine' in value['command']: assert value['binary_sha256'] == win_hash
    else: assert value['binary_sha256'] == sha(ROOT/'workplace/garrison146-gin3')
    snapshots = {e['name']: e for e in value['events'] if e['kind'] == 'capture'}
    assert len(snapshots) == 19
    for e in snapshots.values(): assert sha(folder/e['file']) == e['sha256']
    steps = [e for e in value['events'] if e['kind'] == 'model-step']
    assert len(steps) == len(model['events'])
    for step in steps:
        expected = model['events'][step['index']]; actual = step['observed']
        assert step['source'] == expected
        if expected['state'] in ['StateTurnThinking', 'StateTurnWindow', 'StateTurnReport', 'StateEngagement']: continue
        assert actual['state'] == expected['state']
        want = expected.get('strategic') or {}; got = actual.get('strategic') or {}
        for key in ['turn', 'faction', 'region', 'detail', 'message', 'stock', 'capital_stock', 'idle']:
            if key in want: assert want[key] == got.get(key), (step['index'], key)
        for key in ['fleets', 'editor']:
            left, right = want.get(key, []), got.get(key, [])
            assert len(left) == len(right)
            for l, r in zip(left, right):
                if key == 'fleets':
                    assert all(l[k] == r[k] for k in ['Index', 'Number', 'Admiral', 'Adjutant', 'Region', 'Target', 'Move'])
                    assert len(l['Groups']) == len(r['Groups']); pairs = zip(l['Groups'], r['Groups'])
                else: pairs = [(l, r)]
                for x, y in pairs: assert all(x[k] == y[k] for k in ['kind', 'ships', 'slot2', 'slot3'])
    get = lambda n: snapshots[n]['observed']['strategic']
    arrived = get('garrison-arrived'); formed = get('formed-at-garrison'); departed = get('departed')
    assert not arrived.get('fleets') and arrived['region'] == model['garrison']
    assert arrived['stock'] == [9000, 3000, 16000, 2000, 12000, 3000, 1000]
    assert get('first-commander')['detail'] == model['admirals'][0]
    assert get('changed-commander')['detail'] == model['admirals'][1]
    local = get('local-ship-stock')
    for name in ['support-rejected-5', 'support-rejected-6']:
        rejected = get(name)
        assert rejected['message'] and rejected['stock'] == local['stock'] and rejected['editor'] == local['editor']
    # formed檢查點仍在編輯器；按「決定」後的fleet-ready才有正式艦隊。
    ready = get('fleet-ready')
    assert ready['stock'] == formed['stock'] and ready['capital_stock'] == formed['capital_stock']
    f = ready['fleets'][0]
    assert len(ready['fleets']) == 1 and f['Number'] == 3 and f['Admiral'] == model['admirals'][1]
    assert f['Groups'] == formed['editor']
    assert f['Region'] == model['garrison'] and f['Groups'][0]['kind'] == 2
    assert formed['stock'] == [6000, 3000, 12000, 2000, 9000, 3000, 1000]
    assert arrived['capital_stock'] == formed['capital_stock'] == [102000, 30000, 104000, 36000, 9000, 6000, 5000]
    assert departed['turn'] == 2 and departed['fleets'][0]['Region'] == model['destination']
    loaded = value['loaded']['strategic']['fleets'][0]
    assert loaded['Region'] == model['destination'] and loaded['Admiral'] == f['Admiral'] and loaded['Groups'][0]['kind'] == 2
    assert not snapshots['option-default-off']['observed']['rules']['non_battleship_flagship']
    assert snapshots['option-enabled']['observed']['rules']['non_battleship_flagship']
    assert value['restarted']['rules']['non_battleship_flagship']
    pref = folder/'save/preferences.json'; assert read(pref)['non_battleship_flagship']
    assert len(value['saved']) == 6
    for name, digest in value['saved'].items(): assert sha(folder/'save'/name) == digest
    gui.append(dict(path=str(folder.relative_to(ROOT)), result_sha256=sha(folder/'RESULT.json'),
                    trace_sha256=sha(folder/'FRAMES.jsonl'), binary_sha256=value['binary_sha256'], captures=len(snapshots),
                    model_steps=len(steps), stock_before=arrived['stock'], stock_after=formed['stock'], capital_stock=formed['capital_stock'],
                    fleet=f, destination=model['destination'], saved=value['saved'], preference_sha256=sha(pref)))
promo = OUT/'promo'; audit = read(promo/'AUDIT.json'); edit = read(promo/'PLAN.json')
assert audit['passed'] and audit['visual_review'].startswith('已目視')
assert sha(promo/audit['file']) == audit['sha256'] and sha(promo/'PLAN.json') == audit['plan_sha256']
for name, digest in edit['input_sha256'].items(): assert sha(ROOT/name) == digest
assert not audit['black_segments'] and not audit['subtitle_clipping'] and audit['audio']['snr_db'] > 25
for line in (OUT/'SHA256SUMS.txt').read_text().splitlines():
    digest, name = line.split('  ', 1)
    assert sha(OUT/name) == digest, name
receipt = dict(version=VERSION, platform='Windows amd64，Wine正常操作驗證，未作Windows真機驗收',
               scope='使用者指定非戰艦旗艦選項；原速正常視窗駐地改選提督、領船及移動。無新增DOS parity聲明。',
               archive=build['archive'], binary_sha256=win_hash, build_sha256=sha(OUT/'BUILD.json'),
               frozen_files=len(build['source_files']), product_files=product_count,
               version_output_sha256=sha(version_path), model_sha256=sha(plan_path), gui=gui,
               tests='tools/gui.sh test ./internal/... ./cmd/...：19套件，15 PASS／4無測試；新艦種矩陣與轉移保存測試通過',
               movie=dict(file=str((promo/audit['file']).relative_to(ROOT)), sha256=audit['sha256'], seconds=audit['seconds'],
                          audit_sha256=sha(promo/'AUDIT.json'), plan_sha256=sha(promo/'PLAN.json')),
               reproduction=dict(plan='tools/scripts/garrison_promo_plan.go', gui='tools/scripts/garrison_promo_gui.py',
                                 movie='tools/scripts/garrison_promo_movie.py', verification='tools/scripts/garrison_promo_receipt.py'),
               rights='影片含原版及衍生音畫，僅本機；Windows替換包不含遊戲素材。未發布Release。')
receipt['reproduction_sha256'] = {name: sha(ROOT/name) for name in receipt['reproduction'].values()}
receipt['checksums_sha256'] = sha(OUT/'SHA256SUMS.txt')
target = ROOT/'docs/internals/data/garrison-flagship146.json'
assert target.parent.stat().st_uid == os.getuid()
target.write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print(json.dumps(dict(passed=True, archive=receipt['archive'], movie=receipt['movie'], captures=sum(g['captures'] for g in gui)), ensure_ascii=False, indent=2))
