"""第80輪：凍結實際ALSA輸出、正常X11及保留的音效收尾限制。"""
from pathlib import Path
import hashlib
import json
import os
import wave

os.chdir('/src')
target=Path('docs/internals/data/fleet-playback.json')
assert not target.exists() and target.parent.stat().st_uid==os.getuid()
files={}
def read(path):
    p=Path(path)
    assert p.is_file() and p.stat().st_uid==os.getuid(),str(p)
    b=p.read_bytes()
    files[str(p)]={'sha256':hashlib.sha256(b).hexdigest(),'bytes':len(b)}
    return b
def data(path):return json.loads(read(path))
def tree(root):
    for p in sorted(Path(root).rglob('*')):
        if p.is_file():read(p)

parent='134dcae343fd08b6931b78c63957f34433673a37'
for path,want in {
    'docs/internals/data/fleet-player-gui.json':'0b1c1b79d13442420f9e9bec0e7f20b215ea72dfa262decc779f93b44ce7b6f0',
    'docs/internals/data/fleet-sound-game.json':'4388203c2d6d9c69f69c38f1107b871b92c1397923f38a9eb2b7fe72ed39b9e3',
}.items():assert hashlib.sha256(read(path)).hexdigest()==want,path
probe=data('workplace/round80-device-probe/alsa-probe.json')
assert probe['default_pcm_available'] and not probe['dev_snd_present']
before=data('workplace/round80-before/monitor-stats.json')
after=data('workplace/round80-after/monitor-stats.json')
intro=data('workplace/round80-intro-long/monitor-stats.json')
assert before['nonzero_samples']==0 and before['peak']==0
assert after['nonzero_samples']>0 and intro['nonzero_samples']>0
assert read('workplace/round80-before.png')==read('workplace/round80-after.png')
for name in ['positive','absent','broken-v2']:
    log=read(f'workplace/round80-device-{name}.log').decode()
    assert '--- PASS: TestAudioDeviceActualDefault' in log and 'FAIL' not in log and 'SKIP' not in log

cases=[]
tail_observations=[]
for side,batch,revision in [(0,'round80-gui-0','v80'),(1,'round80-gui-1-final','v81')]:
    root=Path('workplace')/batch
    gui=root/'gui'
    planroot=Path(f'workplace/round80-plan-{side}')
    plan=data(planroot/'plan.json')
    expected=[e['id'] for e in data(f'workplace/round78-chain-a/normal-{side}/sounds.json')]
    assert plan['sound_ids']==expected and len(expected)==[14,10][side]
    receipt=data(gui/'receipt.json')
    summary=data(gui/'playback-summary.json')
    assert receipt['six_diagnostic_saves_equal'] and receipt['intro_default'] and not receipt['script_flag_used']
    assert receipt['sound_device_tested']=='ALSA default / PulseAudio clocked virtual sink'
    assert summary['effect_ids']==expected and summary['effect_stops']==len(expected)
    assert summary['music_starts']==['OPEN_M','BGM_0F','BGM_3F','BGM_6F','BGM_3F']
    assert not summary['human_listening'] and not summary['physical_device_tested']
    stream_inputs=data(gui/'actual-sink-inputs.json')
    assert len(stream_inputs)==1
    si=stream_inputs[0]
    assert not si['mute'] and not si['corked']
    assert si['sample_specification']=='float32le 2ch 44100Hz'
    assert si['properties']['application.process.binary']==f'round80-gin3-{revision}'
    stats=data(root/'monitor-stats.json')
    assert stats['nonzero_samples']>0 and stats['samples_at_or_above_one']==0
    assert len(receipt['comparisons'])==3
    for c in receipt['comparisons']:
        name=c['name']
        assert c['x11_different']==0 and c['game_index_equal']
        native=data(gui/(name+'-native-comparison.json'))
        assert native['native_pixels']==307200 and native['native_different']==0
        assert native['native_bbox'] is None and native['max_channel_difference']==0
        assert read(gui/(name+'.bin'))==read(planroot/(name+'.bin'))
    saves=sorted(gui.glob('*.DT0'))
    assert len(saves)==6
    for p in saves:
        assert read(p)==read(planroot/p.name)==read(Path(f'workplace/round77-chain-c/normal-{side}')/p.name)
    events=[json.loads(line) for line in read(gui/'playback-events.jsonl').decode().splitlines()]
    previous=None
    for event in events:
        if event['kind']=='effect_start':previous=event
        if event['kind']=='effect_stop' and event['playing']:
            assert previous is not None
            # DecodeWithoutResampling為signed16、stereo，四bytes一格。
            nominal=previous['stream_bytes']*1_000_000_000//(4*44100)
            tail_observations.append({'faction':side,'id':previous['id'],'nominal_duration_ns':nominal,
                'stop_position_ns':event['position_ns'],'position_shortfall_ns':nominal-event['position_ns'],
                'start_unix_ns':previous['unix_ns'],'stop_unix_ns':event['unix_ns'],
                'classification':'播放器仍playing的已證實觀測；實際monitor尾端是否被截短unknown'})
    tree(root);tree(planroot)
    read('workplace/'+batch+'.log');read('workplace/'+batch+'-native.log')
    for p in ['workplace/round79-fleet-gui-game-'+revision+'.go',
              'workplace/round79-fleet-gui-main-'+revision+'.go',
              'workplace/round79-fleet-gui-overlay-'+revision+'.json',
              'workplace/round80-music-observe-'+revision+'.go',
              'workplace/round80-effects-observe-'+revision+'.go',
              'workplace/round80-driver-'+revision+'.py','workplace/round80-gin3-'+revision]:read(p)
    cases.append({'faction':side,'batch':str(root),'revision':revision,'normal_gui_actions':len(receipt['actions']),
        'additional_camera_gestures':len(receipt['camera_gestures']),
        'command':f'bash tools/playback.sh {batch} python3 workplace/round80-driver-{revision}.py --plan workplace/round80-plan-{side}/plan.json --out /src/workplace/{batch}/gui --scale 1',
        'music_starts':summary['music_starts'],'effect_ids':expected,'effect_stops_while_playing':summary['effect_stops_while_playing'],
        'native_zero_difference_images':3,'six_saves_equal_plan_and_round77':True,'actual_sink_input':si,'monitor':stats})

assert len(tail_observations)==4
with wave.open('workplace/round80-gui-0/listen.wav','rb') as w:
    assert w.getframerate()==44100 and w.getnchannels()==2 and w.getsampwidth()==2
    assert w.getnframes()*8==files['workplace/round80-gui-0/monitor.f32le']['bytes']

standard=read('workplace/round80-standard.log').decode()
assert sum(line.startswith('ok ') for line in standard.splitlines())==13
assert sum('[no test files]' in line for line in standard.splitlines())==5
assert '\nFAIL' not in standard
images=data('workplace/round80-gui-images.json')
assert images[0]['Id']=='sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8'
assert images[1]['Id']=='sha256:ecb3612ac7c18d0cd225fd55b7b186052a8bf0029e82aa786a23f4b88f196fa8'
tree('workplace/round80-library')
for name in ['round80-device-probe','round80-before','round80-after','round80-intro-long','round80-device-positive']:
    tree('workplace/'+name);read('workplace/'+name+'.log')
for path in ['go.mod','go.sum','cmd/gin3/main.go','cmd/gin3/music.go','cmd/gin3/effects.go',
    'cmd/gin3/audio_device_linux.go','cmd/gin3/audio_device_other.go','cmd/gin3/audio_device_linux_test.go',
    'internal/game/soundeffects.go','docs/spec/104-fleet-sound.md',
    'docs/goal/2026-10-04-fleet-playback-round-80.md',
    'tools/Dockerfile.gui-playback','tools/playback.sh','tools/gui.sh','tools/audio.sh',
    'tools/scripts/playback_sink.py','tools/scripts/playback_stats.py','tools/scripts/playback_library.py',
    'tools/scripts/alsa_device_probe.go','tools/scripts/fleet_playback_prepare.py','tools/scripts/fleet_playback_receipt.py',
    'tools/scripts/fleet_player_gui_prepare.py','tools/scripts/fleet_player_gui_plan.go','tools/scripts/fleet_player_gui_driver.py',
    'tools/scripts/fleet_player_gui_capture.go','tools/scripts/fleet_player_gui_input.go','tools/scripts/fleet_player_gui_native_check.go',
    'workplace/round80-gui-packages.txt','workplace/round80-alsa-probe','workplace/round80-gin3-before','workplace/round80-gin3-after',
    'workplace/round80-music-before.go','workplace/round80-main-before.go',
    'workplace/round80-plan-source-v80.go','workplace/round80-plan-source-v81.go',
    'workplace/round80-prepare.log','workplace/round80-prepare-v81.log','workplace/round80-model.log',
    'workplace/round80-gui-1.log','workplace/round80-gui-1/gui/failed-trace.json',
    'workplace/round80-gui-1/gui/updates.jsonl','workplace/round80-device-broken.log',
    'workplace/round79-input-v2','workplace/round79-capture','workplace/round79-native-check',
    'workplace/gomodcache/github.com/hajimehoshi/ebiten/v2@v2.8.8/audio/vorbis/vorbis.go',
    'workplace/gomodcache/github.com/ebitengine/oto/v3@v3.3.3/driver_unix.go']:
    if 'gomodcache/' in path:
        p=Path(path);b=p.read_bytes();files[path]={'sha256':hashlib.sha256(b).hexdigest(),'bytes':len(b),'ownership':'唯讀既有模組快取'}
    else:read(path)

result={'date':'2026-10-04','parent_git_head':parent,'worktree_modified':True,
    'original_executor_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','original_executor_modified':False,
    'original_bundle_sha256':'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673',
    'classification':'正式ALSA/default經有時鐘PulseAudio虛擬輸出；正常remake GUI，非新的原版same-state',
    'images':{'parent':images[0]['Id'],'playback':images[1]['Id']},
    'toolchain':{'pulseaudio':'16.1+dfsg1-2+b1','libasound2-plugins':'1.2.7.1-1','complete_packages':'workplace/round80-gui-packages.txt',
        'oto':'3.3.3','ebitengine':'2.8.8','sample_rate':44100,'channels':2},
    'device_probe':probe,'before_monitor':before,'after_monitor':after,'same_frame_png_equal':True,
    'intro_monitor':intro,'cases':cases,'tail_observations':tail_observations,
    'physical_sound_card':False,'human_listening':False,'loop_seam_acceptance':False,
    'native_one_pixel_decision':'#89 pending，未修改順位或豁免像素',
    'spec104':'READY','full092':'DRAFT','standard_tests':{'pass':13,'no_test_files':5},
    'files':files}
target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('凍結來源',len(files),'SHA-256',hashlib.sha256(target.read_bytes()).hexdigest())
