"""第78輪音效重播overlay及完整驗證索引；歷史輸出拒絕覆寫。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

os.chdir('/src')
def digest(b): return hashlib.sha256(b).hexdigest()
def create(path, text):
    p = Path(path)
    assert not p.exists() and p.parent.stat().st_uid == os.getuid(), path
    p.write_text(text)

if sys.argv[1] == '--overlay':
    source = 'internal/game/fleetcombat.go'
    target = 'workplace/round78-sound-replay.go'
    create(target, Path(source).read_text()+'''\n// 診斷重播只固定RND，沒有局面、鏡頭或命令注入。
func (g *Game) FleetChainProbeSeed(seed uint32) { g.battleRand = rules.NewRand(seed) }
''')
    subprocess.run(['gofmt', '-w', target], check=True)
    create('workplace/round78-sound-replay.json', json.dumps({'Replace': {str(Path(source).resolve()): str(Path(target).resolve())}}, indent=2)+'\n')
    sys.exit(0)

assert sys.argv[1] == '--freeze'
files = {}
def read(path):
    p = Path(path)
    assert p.stat().st_uid == os.getuid(), str(p)
    b = p.read_bytes()
    files[str(p)] = {'sha256': digest(b), 'bytes': len(b)}
    return b

assert digest(read('docs/internals/data/fleet-sound-source.json')) == '5d3b31b8d16c4fc34b655b52a9caadfa40cdbecbf2d87f21bf1942ec75db2d66'
assert digest(read('docs/internals/data/fleet-sound-spec.json')) == 'e4c40427f3edbb6b245b1cf4150fe5faf66476feb981eaa806b7150319fb5a50'
tracks = []
receipts = [json.loads(read(f'workplace/round78-sfx-final-{x}/SFX-RECEIPT.json')) for x in ['a', 'b']]
for r in receipts: assert len(r['tracks']) == 28
for a, b in zip(receipts[0]['tracks'], receipts[1]['tracks']):
    assert a['id'] == b['id']
    n = f"SFX_{a['id']:02d}.ogg"
    assert read('workplace/round78-sfx-final-a/'+n) == read('workplace/round78-sfx-final-b/'+n)
    assert a['ogg_sha256'] == b['ogg_sha256'] and a['decoded_pcm_sha256'] == b['decoded_pcm_sha256']
    tracks.append(a)
decoded = json.loads(read('workplace/round78-sfx-decoded.json'))
assert len(decoded) == 28
normal = []
for side in [0, 1]:
    dirs = [Path(f'workplace/round78-chain-{x}/normal-{side}') for x in ['a', 'b']]
    names = sorted(p.name for p in dirs[0].iterdir() if p.is_file())
    assert names == sorted(p.name for p in dirs[1].iterdir() if p.is_file())
    for n in names: assert read(dirs[0]/n) == read(dirs[1]/n), n
    saves = [n for n in names if n.endswith('.DT0')]
    assert len(saves) == 6, saves
    for n in saves: assert read(dirs[0]/n) == read(Path(f'workplace/round77-chain-c/normal-{side}')/n)
    events = json.loads(read(dirs[0]/'sounds.json'))
    app_events = json.loads(read(f'workplace/round78-app-{side}.json'))
    assert events == app_events and len(events) > 0
    plan = json.loads(read(dirs[0]/'inputs.json'))
    assert plan['cpu_position_unchanged']
    normal.append({'faction': side, 'events': events, 'orders': plan['orders'], 'hits': plan['scenes'], 'six_saves_unchanged': True,
                   'double_run_equal': True, 'formal_bootstrap_app_events_equal': True, 'cpu_position_unchanged': True})
for name in ['round78-sound-native-fixed.log', 'round78-sound-gates.log', 'round78-standard.log', 'round78-sfx-decode.log',
             'round78-sfx-final-a.log', 'round78-sfx-final-b.log', 'round78-chain-a.log', 'round78-chain-b.log', 'round78-app.log']:
    b = read('workplace/'+name)
    assert b'--- FAIL' not in b and b'\nFAIL' not in b
standard = Path('workplace/round78-standard.log').read_text()
assert sum(l.startswith('ok ') for l in standard.splitlines()) == 13
sources = ['cmd/gin3/main.go', 'cmd/gin3/effects.go', 'cmd/gin3/effects_test.go', 'internal/asset/svf.go', 'internal/asset/svf_test.go',
           'internal/audio/library.go', 'internal/audio/effects.go', 'internal/bootstrap/load.go', 'internal/game/game.go', 'internal/game/intro.go',
           'internal/game/fleetcombat.go', 'internal/game/fleetcombatscene.go', 'internal/game/soundeffects.go', 'internal/game/soundeffects_test.go',
           'internal/game/fleetplayerchain_test.go', 'tools/scripts/fleet_sound_convert.go', 'tools/scripts/fleet_sound_game_receipt.py',
           'tools/sound_to_ogg.sh', 'workplace/round78-sound-replay.go', 'workplace/round78-sound-replay.json']
for p in sources: read(p)
create('docs/internals/data/fleet-sound-game.json', json.dumps({'schema': 'logh3.fleet-sound-game.v1', 'date': '2026-10-04',
    'parent_commit': 'e927750a782af344d77e6a3dd866e18af753ad28', 'image': 'logh3-ebiten:2; logh3-audio:2',
    'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c', 'dosgolem_unchanged': True, 'seed': 0x12345678,
    'tracks': tracks, 'formal_decoder': decoded, 'normal': normal, 'files': files,
    'classification': '原版受控COMANI序列／原池／seed；正常remake公開輸入與正式app事件；硬體時序近似',
    'limits': ['無真實音效裝置或人耳驗收', '完整GUI與正常原版整段same-state仍缺', '104保持READY，#48／#79保持OPEN',
               '#89未核准像素保持，歷史收據未覆寫', '原版音效與衍生OGG不可放公開發行包']}, ensure_ascii=False, indent=2)+'\n')
p = Path('docs/internals/data/fleet-sound-game.json')
print(p, digest(p.read_bytes()), len(files), 'sources')
