"""第105輪：凍結Windows正常播放觀測、完整失敗及工具鏈身份。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/src') and os.getuid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('--freeze', action='store_true')
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
assert args.freeze != args.check
target = Path('docs/internals/data/windows-music-playback.json')


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


if args.freeze:
    assert not target.exists()
    root = Path('workplace/music69-windows')
    expected_binary = '6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9'
    assert sha(root / 'source103-player/gin3.exe') == expected_binary
    assert sha('docs/internals/data/current-source.json') == '033ec921cd4a67b135cb34ad157fc7271c355ecd8bb7245aaa64cd346e3f71f1'
    assert sha('docs/internals/data/windows-player.json') == '5e4af9398f5eb1775e262dfaec971c9f50f4767bfd9c2484993a6b839f02c5ef'
    runtime = {}
    for batch in ['e','f']:
        folder = root / ('playeraudio69-r105-'+batch)
        check = json.loads((folder / 'client/CHECK.json').read_text())
        start = json.loads((folder / 'client/START.json').read_text())
        driver = json.loads((folder / 'DRIVER.json').read_text())
        analysis = json.loads((folder / 'ANALYSIS.json').read_text())
        fixed = json.loads((folder / 'FIXED-GATE.json').read_text())
        assert check['exit'] == 0 and check['actions'] == 33 and len(check['frames']) == 6
        assert all(r['different'] == '0' and r['exit'] == 0 for r in check['frames'])
        assert 'state=StateSettings\n' in check['state'] and 'updates=3600\n' in check['state']
        assert start['binary_sha256'] == expected_binary and len(start['music_sha256']) == 22
        assert start['default_intro'] and not start['script_used'] and not start['state_injection']
        assert start['mesa_threads'] == '1' and not start['port_mapping_changed']
        assert driver['driver'] == 'alsa' and not driver['port_mapping_changed']
        assert analysis['normal_gui_passed'] and analysis['additional_phase_samples_verified']
        assert analysis['first_feature_gate_passed'] == (batch == 'f')
        assert fixed['exit'] == (0 if batch == 'f' else 1)
        assert all(r['nonzero'] == 0 for r in analysis['silence'])
        assert not any(analysis[k] for k in ['full_playback_accepted','physical_sound_card','human_listening','loop_verified'])
        assert sha(folder / 'monitor.f32le') == analysis['monitor']['sha256']
        for track in analysis['tracks']:
            assert sha(track['clip']['path']) == track['clip']['sha256']
            assert track['clip']['contains_observed_output_without_gap_removal']
        runtime[batch] = {'ui':check,'start':start,'driver':driver,'analysis':analysis,'first_gate':fixed}
    assert not (root / 'playeraudio69-r105-e/PCM-CHECK.json').exists()
    pcm = json.loads((root / 'playeraudio69-r105-f/PCM-CHECK.json').read_text())
    assert [r['track'] for r in pcm['matching_sources']] == ['OPEN_M','BGM_0F','BGM_1F']
    assert all(r['continuous_exact_seconds'] >= 3 and r['overlap_with_requested_state_seconds'] >= 3
               and r['different'] == 0 for r in pcm['matching_sources'])
    controls = {}
    for name, expected in [('pulse-control', True),('oto-control',False),('alsa-control',True)]:
        folder = root / ('playeraudio69-r105-'+name)
        control = json.loads((folder / 'CONTROL-CHECK.json').read_text())
        assert control['full_source_found'] == expected and not control['production_player_verified']
        assert not control['gui_used']
        if expected: assert control['continuous_exact_seconds'] == 8
        controls[name] = control
    decoded = []
    for name in ['OPEN_M','BGM_0F','BGM_0M','BGM_1F']:
        linux = root / 'playeraudio69-r105-decode' / (name+'.f32le')
        windows = root / 'playeraudio69-r105-decode-windows' / (name+'.f32le')
        assert linux.read_bytes() == windows.read_bytes()
        decoded.append({'track':name,'pcm_sha256':sha(linux),'bytes':linux.stat().st_size,
                        'linux_equals_windows':True})
    tests = Path('workplace/round105-standard-gui.log').read_text().splitlines()
    assert sum(t.split()[:1] == ['ok'] for t in tests) == 13
    assert sum('[no test files]' in t for t in tests) == 5 and not any(t.startswith('FAIL') for t in tests)
    lock = Path('tools/windows-audio-packages.lock').read_text().splitlines()
    assert len(lock) == len(set(lock)) == 50
    assert 'libasound2-plugins:amd64=1.2.7.1-1ubuntu5' in lock
    before = dict(l.split('\t') for l in Path('workplace/round105-wine-parent-packages.txt').read_text().splitlines())
    after = dict(l.split('\t') for l in Path('workplace/round105-wine-audio-packages.txt').read_text().splitlines())
    assert all(after[name] == version for name,version in before.items())
    assert sorted(lock) == sorted(name+'='+version for name,version in after.items() if name not in before)
    images = json.loads(Path('workplace/round105-windows-audio-images.json').read_text())
    assert images[0]['Id'] == 'sha256:28adb713b85470848a83d297279520f1e02bb2ea7f74c411008cb07c253633e0'
    names = ['windows_music_player','windows_music_compare','windows_music_analysis',
             'windows_music_control_compare','windows_music_oto_control','windows_music_pulse_control',
             'windows_music_playback_receipt','wine_audio_driver','wine_playback_sink']
    paths = {'tools/scripts/'+name+'.py' for name in names}
    paths.update({'tools/scripts/windows_music_decode.go','tools/scripts/windows_music_oto_probe.go',
                  'tools/player_wine_audio.sh','tools/build_windows_audio.sh','tools/Dockerfile.windows-audio',
                  'tools/windows-audio-packages.lock','tools/music_wine_entry.sh','tools/gui.sh',
                  'docs/goal/2026-10-05-windows-playback-round-105.md',
                  'docs/playtest/093-windows-music-playback.md',
                  'docs/internals/data/current-source.json','docs/internals/data/windows-player.json',
                  'docs/internals/data/windows-combined-music-tools.json',
                  'workplace/round105-standard-gui.log','workplace/round105-windows-audio-images.json',
                  'workplace/round105-wine-parent-packages.txt','workplace/round105-wine-audio-packages.txt',
                  'workplace/round105-windows-audio-build.log','workplace/round105-windows-audio-locked-build.log',
                  'workplace/round105-windows-audio-prototype.Dockerfile',
                  'workplace/music69-windows/source103-player/gin3.exe',
                  'workplace/music69-windows/windows-music-decode-r105.exe',
                  'workplace/music69-windows/windows-music-oto-probe-r105.exe','go.mod','go.sum'})
    paths.update(str(p) for base in ['cmd','internal'] for p in Path(base).rglob('*.go'))
    for folder in sorted(root.glob('playeraudio69-r105-*')):
        if folder.is_file():
            paths.add(str(folder))
        elif folder.is_dir():
            for base, dirs, files in os.walk(folder):
                dirs[:] = [d for d in dirs if d != 'wine-prefix']
                for name in files:
                    p = Path(base)/name
                    if p.is_file() and not p.is_symlink(): paths.add(str(p))
    catalog = json.loads(Path('docs/internals/data/windows-combined-music-tools.json').read_text())
    for group in ['GM','FM']:
        for row in catalog['full'][group]['tracks']:
            path = root / 'combined-utf8-full/生成 曲庫' / (row['track']+'.ogg')
            assert sha(path) == row['ogg_sha256']
            paths.add(str(path))
    sources = {}
    for name in sorted(paths):
        p = Path(name)
        assert p.stat().st_uid == p.stat().st_gid == 1000 and not p.is_symlink(),name
        sources[name] = {'bytes':p.stat().st_size,'sha256':sha(p)}
    for name in names: ast.parse(Path('tools/scripts/'+name+'.py').read_text())
    for path in ['tools/player_wine_audio.sh','tools/build_windows_audio.sh']:
        subprocess.run(['bash','-n',path],check=True)
    parent = json.loads(Path('docs/internals/data/windows-player.json').read_text())
    record = {'schema':'logh3.windows-music-playback-receipts.v1','date':'2026-10-05','issue':69,
              'parent_commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
              'binary_sha256':expected_binary,'runtime':runtime,'controls':controls,
              'decoded':decoded,'images':images,'source_count':len(sources),'sources':sources,
              'original_bundle':parent['original_bundle'],'original_files':parent['original_files'],
              'standard':{'packages':18,'pass':13,'no_tests':5},
              'limitations':['e首段FAIL與缺口保持；f只通過所列連續內段及停止',
                             '虛擬Wine／ALSA音訊，精確缺口成因unknown，非實體卡或人耳',
                             '整首／loop、Windows公開存讀、macOS及正式dist-all仍缺，#69 OPEN',
                             '沒有新增原版same-state或修改產品規則']}
    with target.open('x') as stream:
        json.dump(record,stream,ensure_ascii=False,indent=2);stream.write('\n')
record = json.loads(target.read_text())
for name,row in record['sources'].items():
    p = Path(name)
    assert p.stat().st_size == row['bytes'] and sha(p) == row['sha256'],name
for name,row in record['original_files'].items():
    p = Path('/orig') / name
    assert p.stat().st_size == row['bytes'] and sha(p) == row['sha256'],name
print(len(record['sources']),'來源與',len(record['original_files']),
      '原版輸入身份通過；e失敗保留／f限定內段，索引SHA-256',sha(target))
