"""規格139：正常戰鬥播放、設定選單及新版封包驗收完成後產生來源索引。"""
import argparse
import hashlib
from pathlib import Path
import shutil
import tarfile
import zipfile

from toolbar_delivery_finalize import ROOT, read, record, sha, write

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--version', required=True)
a = p.parse_args()
version = a.version
out = ROOT/'dist-all'/version
batch = ROOT/'workplace/battle129'
build = read(out/'BUILD.json')
assert build['version'] == version and len(build['packages']) == 8
frozen = ROOT/'workplace'/('package92-'+version+'-final')/'source'
for name, expected in build['source_files'].items():
    assert sha(frozen/name) == expected['sha256']
    if not name.startswith('vendor/'):
        assert sha(ROOT/name) == expected['sha256'], ('建置後來源改變', name)
manifests = []
for item in build['packages']:
    path = out/item['path']
    assert sha(path) == item['sha256'] and path.stat().st_size == item['bytes']
    if path.suffix == '.zip':
        with zipfile.ZipFile(path) as archive:
            assert archive.testzip() is None
            manifest = read(path.parent/path.stem/'MANIFEST.json')
            assert manifest['version'] == version
            assert set(archive.namelist()) == {path.stem+'/'+n for n in manifest['files']} | {path.stem+'/MANIFEST.json'}
            for n, expected in manifest['files'].items():
                assert not Path(n).is_absolute() and '..' not in Path(n).parts
                data = archive.read(path.stem+'/'+n)
                assert len(data) == expected['bytes'] and hashlib.sha256(data).hexdigest() == expected['sha256']
                assert sha(path.parent/path.stem/n) == expected['sha256']
                if path.parent.name == 'release':
                    assert not {'original','art','music','soundfont'}.intersection(Path(n).parts)
                if path.parent.name == 'full' and n == 'README.txt':
                    text = data.decode('utf-8-sig')
                    assert 'F1' in text and 'F7' in text and '設定選單' in text and 'F8' not in text
                if 'windows' not in path.name and (n.endswith(('.sh','.command','/MacOS/gin3','/MacOS/launch')) or n == 'gin3'):
                    assert archive.getinfo(path.stem+'/'+n).external_attr >> 16 & 0o111
            manifests.append({'path':item['path'],'files':len(manifest['files']),'passed':True})
    else:
        with tarfile.open(path, 'r:gz') as archive:
            files = [m for m in archive.getmembers() if m.isfile()]
            assert len(files) == len(build['source_files'])
            for member in files:
                n = str(Path(member.name).relative_to('logh3-source'))
                assert hashlib.sha256(archive.extractfile(member).read()).hexdigest() == build['source_files'][n]['sha256']
linux = out/'full'/('LOGH3-'+version+'-linux-amd64')
normal = []
for mode, faction in [('original',0),('enhanced',1)]:
    folder = batch/('v116-full-'+mode)
    r = read(folder/'RESULT.json')
    assert r['passed'] and r['normal_entry'] and not r['state_injection'] and not r['script_flag']
    assert r['faction'] == faction and r['command'][0] == str(linux/('start-'+mode+'.sh'))
    assert sum('check' in e for e in r['events']) >= 32
    shutil.copytree(folder, out/'smoke'/('normal-'+mode), ignore=shutil.ignore_patterns('pulse','save'), dirs_exist_ok=True)
    normal.append(record(folder/'RESULT.json'))
playback = []
for name in ['normal-hd-final','normal-original','normal-fleet-final']:
    folder = batch/name
    r = read(folder/'PLAYBACK-AUDIT.json')
    t = read(folder/'TRACE.json')
    assert r['passed'] and r['scenes'] and t['ai'] == -1 and not t['state_injection'] and not t['script_flag']
    assert sha(folder/'FRAMES.jsonl') == r['frame_trace_sha256']
    playback.append(record(folder/'PLAYBACK-AUDIT.json'))
    target = out/'smoke'/('playback-'+name)
    shutil.copytree(folder, target, ignore=shutil.ignore_patterns('pulse','save','player-save','RAW.mp4'), dirs_exist_ok=True)
    shutil.copyfile(folder/'RAW.mp4', out/'promo'/(name+'.mp4'))
assert read(batch/'normal-original/TRACE.json')['binary_sha256'] == sha(linux/'gin3')
fleet = read(batch/'normal-fleet-final/PLAYBACK-AUDIT.json')
assert {s['weapon'] for s in fleet['scenes'] if s['kind']=='fleet'} == {0,1,2}
assert all(s['graphics']==['hd'] for s in fleet['scenes'])
assert read(batch/'ANIMATION-EQUIVALENCE.json')['identical_pngs'] == 157
wine = ROOT/'workplace/music69-windows'/('package92-smoke-'+version)
engine_wine = ROOT/'workplace/music69-windows'/('release-smoke-'+version)
assert len(read(wine/'SMOKE.json')['runs']) == 6
assert len(read(engine_wine/'SMOKE.json')['runs']) == 4
assert len(read(out/'smoke/release-linux/SMOKE.json')['runs']) == 4
shutil.copytree(wine, out/'smoke/full-windows-wine', ignore=shutil.ignore_patterns('LOGH3-*','save-*'), dirs_exist_ok=True)
shutil.copytree(engine_wine, out/'smoke/release-windows-wine', ignore=shutil.ignore_patterns('LOGH3-*','save-*'), dirs_exist_ok=True)
mac = read(out/'smoke/macos/MACOS.json')
assert {s['architecture'] for s in mac['slices']} == {'amd64','arm64'}
assert all(c['all_page_hashes_match'] for s in mac['slices'] for c in s['code_directory'])
lines = (batch/'final-standard-tests-final.log').read_text().splitlines()
assert not any(line.startswith(('FAIL','--- FAIL')) for line in lines)
tests = {'passed':sum(line.startswith('ok ') for line in lines),'no_tests':sum(line.startswith('? ') for line in lines)}
assert tests == {'passed':15,'no_tests':4}
shutil.copyfile(batch/'final-standard-tests-final.log', out/'smoke/standard-tests.log')
previous = ROOT/'dist-all/v.1.1.3-20261011/promo/LOGH3-v.1.1.3-20261011-promo.mp4'
assert sha(previous) == '52c7353548a1175153a5d79f312f9145a386eafaa7ef5f61f0702b9127be37cb'
(out/'promo/README.md').write_text('推廣片沿用v.1.1.3的116秒影片，保留原身份。\n\n新版F7切原版／HD；敵軍顯示只在設定選單，沒有快捷鍵。舊片F7顯示敵軍已不適用。\n正常播放核對實錄及收據見docs/playtest/118-battle-playback-and-settings.md。\n')
check = {'version':version,'passed':True,'tests':tests,'source_files':len(build['source_files']),
         'normal_gui':normal,'playback':playback,'manifests':manifests,
         'windows_wine_runs':6,'linux_engine_runs':4,'windows_engine_wine_runs':4,'macos':mac}
write(out/'smoke/PACKAGE-CHECK.json', check)
sources = [record(ROOT/n) for n in build['source_files'] if not n.startswith('vendor/')]
sources += [record(q) for q in (ROOT/'assets/hd').rglob('*') if q.is_file()]
sources += [record(ROOT/n) for n in ['tools/scripts/battle_delivery_finalize.py',
    'tools/scripts/battle_playback_audit.py','tools/scripts/toolbar_delivery_finalize.py',
    'tools/scripts/battle_fleet_prepare.py','tools/scripts/battle_fleet_record.py',
    'tools/scripts/fleet_player_gui_prepare.py','tools/scripts/fleet_player_gui_driver.py',
    'tools/scripts/fleet_player_gui_input.go','tools/scripts/fleet_player_gui_plan.go',
    'tools/scripts/toolbar_gui_check.py','tools/scripts/ai_gameplay_record.py',
    'tools/scripts/local_test_package.py','tools/scripts/local_test_macos_package.py',
    'tools/scripts/local_test_macos_build.sh','tools/scripts/v1_release.py',
    'docs/spec/139-battle-playback-and-settings.md','docs/playtest/118-battle-playback-and-settings.md']]
shutil.copyfile(batch/'CLEANUP.json',out/'smoke/CLEANUP.json')
evidence = [record(q) for q in batch.rglob('*') if q.is_file() and q.name!='finalize.log' and (q.suffix in ('.json','.jsonl','.log','.png','.mp4','.go','.bin','.DT0','.pprof') or q.name.startswith(('gin3-observe','input-tap','capture'))) and not {'save','pulse','provisional-v116'}.intersection(q.relative_to(batch).parts)]
evidence += [record(q) for q in (out/'smoke').rglob('*') if q.is_file() and q.suffix in ('.json','.jsonl','.png','.log','.txt')]
write(ROOT/'docs/internals/data/battle-playback-settings.json', {
    'version':version,'status':'PASS 正常播放及敵軍設定選單', 'packages':build['packages'],
    'source_head':build['head'],'dirty_source_snapshot':True,
    'frozen_build_source':str(frozen.relative_to(ROOT)), 'checks':check,
    'sources':sources,'evidence':evidence,'previous_promo':record(previous)})
print('PASS 正常戰鬥、設定選單、八封包及來源身份')
