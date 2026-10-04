"""第81輪：凍結正式音效截短、修正後完成閘門及正常GUI收據。"""
from pathlib import Path
import hashlib
import json
import os
import wave

os.chdir('/src')
target=Path('docs/internals/data/sound-tail.json')
assert not target.exists() and target.parent.stat().st_uid==os.getuid()
files={}
def read(path,cache=False):
    p=Path(path)
    assert p.is_file() and (cache or p.stat().st_uid==os.getuid()),str(p)
    b=p.read_bytes()
    files[str(p)]={'sha256':hashlib.sha256(b).hexdigest(),'bytes':len(b)}
    if cache:files[str(p)]['ownership']='唯讀鎖版模組'
    return b
def data(path):return json.loads(read(path))
def tree(path):
    for p in sorted(Path(path).rglob('*')):
        if p.is_file():read(p)

parent='1993f1a04b43dab7e9b5372dc6175d88d62dbab6'
assert hashlib.sha256(read('docs/internals/data/fleet-playback.json')).hexdigest()=='f3a37d50fc86d2485113a2e20f84e6c2c2c07e6b12e6697d30d5a2c3341fb209'
full=data('workplace/round81-pcm-complete-v1/pcm-gap-classification.json')
assert full['calls']==28 and full['complete_exact']==26 and full['complete_source_retained']==28
replace=data('workplace/round81-pcm-replace-v1/pcm-comparison.json')
assert replace['calls']==2 and replace['complete_exact']==2
cases=[]
for phase in ['before','after']:
    for side in range(2):
        batch=f'round81-{phase}-gui-greedy-{side}'
        root=Path('workplace')/batch
        gui=root/'gui'
        planroot=Path(f'workplace/round81-before-plan-{side}')
        plan=data(planroot/'plan.json')
        receipt=data(gui/'receipt.json')
        summary=data(gui/'playback-summary.json')
        comparison=data(root/'pcm-comparison.json')
        gaps=data(root/'pcm-gap-classification.json')
        expected=[e['id'] for e in data(f'workplace/round78-chain-a/normal-{side}/sounds.json')]
        assert summary['effect_ids']==plan['sound_ids']==expected and len(expected)==[14,10][side]
        assert receipt['six_diagnostic_saves_equal'] and receipt['intro_default'] and not receipt['script_flag_used']
        assert plan['music_off_by_public_settings']
        assert summary['music_starts']==['OPEN_M','BGM_0F','BGM_3F']
        assert comparison['calls']==len(expected)
        if phase=='after':assert comparison['complete_exact']==len(expected) and gaps['complete_source_retained']==len(expected)
        stats=data(root/'monitor-stats.json')
        assert stats['nonzero_samples']>0 and stats['samples_at_or_above_one']==0
        events=[json.loads(line) for line in read(gui/'playback-events.jsonl').decode().splitlines()]
        stops=[e for e in events if e['kind']=='effect_stop']
        assert len(stops)==len(expected)
        if phase=='after':assert all(not e['playing'] and e['buffer']['buffer_bytes']==0 and not e['buffer']['playing'] for e in stops)
        if phase=='before' and side==0:
            for index,frames in [(4,81),(12,974)]:
                c=comparison['cases'][index];stop=stops[index]
                assert not c['complete_exact'] and c['remaining_source_frames']==frames
                assert stop['playing'] and stop['buffer']['playing']
                assert stop['buffer']['buffer_bytes']==frames*8
                assert stop['buffer']['stream_bytes']==c['expected_frames']*8
        for c in receipt['comparisons']:
            name=c['name']
            assert c['x11_different']==0 and c['game_index_equal']
            assert read(gui/(name+'.bin'))==read(planroot/(name+'.bin'))
            native=data(gui/(name+'-native-comparison.json'))
            assert native['native_pixels']==307200 and native['native_different']==0 and native['max_channel_difference']==0 and native['native_bbox'] is None
        assert len(receipt['comparisons'])==3
        saves=sorted(gui.glob('*.DT0'));assert len(saves)==6
        for p in saves:
            assert read(p)==read(planroot/p.name)
            old=bytearray(read(Path(f'workplace/round77-chain-c/normal-{side}')/p.name))
            if p.name=='WORK.DT0':old[0x1F]=2
            assert p.read_bytes()==old,p.name
        tree(root);tree(planroot)
        cases.append({'phase':phase,'faction':side,'batch':str(root),
                      'sink':data(root/'sink-receipt.json'),'actions':len(receipt['actions']),
                      'additional_camera_gestures':len(receipt['camera_gestures']),
                      'sound_ids':expected,'pcm':comparison,'gaps':gaps,
                      'stops':stops,'monitor':stats,'native_zero_difference_images':3,'six_saves_equal':True})

root=Path('workplace/round81-after-gui-mixed-0');gui=root/'gui'
mixed=data(gui/'receipt.json');summary=data(gui/'playback-summary.json')
assert len(mixed['comparisons'])==3 and mixed['six_diagnostic_saves_equal']
assert summary['music_starts']==['OPEN_M','BGM_0F','BGM_3F','BGM_6F','BGM_3F']
assert len(summary['effect_ids'])==14 and summary['effect_stops_while_playing']==0
events=[json.loads(s) for s in read(gui/'playback-events.jsonl').decode().splitlines()]
assert all(e['buffer']['buffer_bytes']==0 for e in events if e['kind']=='effect_stop')
mixedstats=data(root/'monitor-stats.json');assert mixedstats['nonzero_samples']>0 and mixedstats['samples_at_or_above_one']==0
for c in mixed['comparisons']:
    assert c['x11_different']==0 and c['game_index_equal']
    native=data(gui/(c['name']+'-native-comparison.json'));assert native['native_different']==0
for p in gui.glob('*.DT0'):assert read(p)==read(Path('workplace/round77-chain-c/normal-0')/p.name)
tree(root)
with wave.open(str(root/'listen.wav'),'rb') as w:
    assert w.getframerate()==44100 and w.getnchannels()==2 and w.getsampwidth()==2
    assert w.getnframes()*8==(root/'monitor.f32le').stat().st_size
tree('workplace/round80-plan-0')
tree('workplace/round80-library')
standard=read('workplace/round81-standard.log').decode()
assert sum(line.startswith('ok ') for line in standard.splitlines())==13
assert sum('[no test files]' in line for line in standard.splitlines())==5 and '\nFAIL' not in standard
tests=read('workplace/round81-target-final.log').decode()
assert '--- PASS: TestFleetSoundBackendCompletionWait' in tests and 'SKIP' not in tests and '\nFAIL' not in tests
for p in Path('workplace').glob('round81-*'):
    if p.is_dir():tree(p)
    elif p.is_file():read(p)
for revision in ['t81before','t81b2','t81b3','t81b4','t81after','t81mixed']:
    for prefix,suffix in [('round79-fleet-gui-game-','.go'),('round79-fleet-gui-main-','.go'),
                          ('round79-fleet-gui-overlay-','.json'),('round80-music-observe-','.go'),
                          ('round80-effects-observe-','.go'),('round80-driver-','.py'),
                          ('round80-plan-source-','.go'),('round80-gin3-','')]:
        read('workplace/'+prefix+revision+suffix)
for p in ['go.mod','go.sum','cmd/gin3/main.go','cmd/gin3/music.go','cmd/gin3/effects.go','cmd/gin3/effects_test.go',
          'internal/game/soundeffects.go','internal/game/soundeffects_test.go','docs/spec/105-sound-completion.md',
          'docs/spec/104-fleet-sound.md','docs/goal/2026-10-04-sound-tail-round-81.md',
          'tools/playback.sh','tools/gui.sh','tools/audio.sh','tools/scripts/playback_sink.py','tools/scripts/playback_stats.py',
          'tools/scripts/fleet_playback_prepare.py','tools/scripts/fleet_player_gui_prepare.py',
          'tools/scripts/fleet_player_gui_driver.py','tools/scripts/fleet_player_gui_plan.go',
          'workplace/round79-input-v2','workplace/round79-capture','workplace/round79-native-check']:
    read(p)
for p in Path('tools/scripts').glob('sound_tail*'):read(p)
for p in ['github.com/ebitengine/oto/v3@v3.3.3/internal/mux/mux.go',
          'github.com/hajimehoshi/ebiten/v2@v2.8.8/audio/player.go',
          'github.com/hajimehoshi/ebiten/v2@v2.8.8/audio/audio.go']:
    read(Path('workplace/gomodcache')/p,cache=True)
result={'date':'2026-10-04','parent_git_head':parent,'worktree_modified':True,
        'original_executor_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','original_executor_modified':False,
        'original_bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
        'classification':'正常remake真實X11／正式ALSA音效；非新增原版same-state',
        'images':data('workplace/round81-gui-images.json'),
        'full28':full,'fixed_replacement_probe':replace,'cases':cases,
        'mixed':{'summary':summary,'monitor':mixedstats,'six_saves_equal':True,'native_zero_difference_images':3},
        'spec105':'CONFORMED：限定正式完成回報','spec104':'READY','full092':'DRAFT',
        'physical_sound_card':False,'human_listening':False,'native_one_pixel_decision':'#89 pending',
        'standard_tests':{'pass':13,'no_test_files':5},'files':files}
with target.open('x') as f:f.write(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('凍結來源',len(files),'SHA-256',hashlib.sha256(target.read_bytes()).hexdigest())
