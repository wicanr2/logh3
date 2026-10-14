"""第135輪：正常GUI、公開六表及原版要塞登陸完整幀核對。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--freeze',action='store_true')
p.add_argument('--validate',action='store_true')
a=p.parse_args()
root=Path('/ref')
assert root.is_dir() and os.getuid()==1000
w=root/'workplace/round135'
target=Path('/work/docs/internals/data/engagement-integration.json')

def digest(path):
    h=hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda:f.read(1024*1024),b''):h.update(block)
    return {'bytes':path.stat().st_size,'sha256':h.hexdigest()}

def read(path):return json.loads(path.read_text())

paths=set()
pair_count=0
for variant in ['fortress','iserlohn']:
    for kind in ['failure','capture']:
        left=root/('workplace/oracle/landing48-r135-'+variant+'-'+kind+'-a')
        right=left.with_name(left.name[:-1]+'b')
        names=sorted(p.name for p in left.iterdir() if p.is_file() and p.name!='report.json')
        assert names==sorted(p.name for p in right.iterdir() if p.is_file() and p.name!='report.json')
        for name in names:assert digest(left/name)==digest(right/name),(variant,kind,name)
        for folder in [left,right]:
            report=read(folder/'report.json')
            assert report['dosgolem_head']=='f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c' and not report['dosgolem_dirty']
            assert report['seed_once'] and report['injections'] and report['canvas']==[640,480]
            assert report['parent_sha256']=='c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1'
            assert 'landing_fortress_oracle.sh' in report['command']
            paths.update(p for p in folder.iterdir() if p.is_file())
        pair_count+=len(names)
assert pair_count==792
typed=read(w/'fortress-typed.json')
assert len(typed['Cases'])==162
for name,want in typed['Sources'].items():assert digest(root/name)==want,name
comparison=read(w/'fortress-comparison-final/comparison.json')
assert len(comparison)==162 and sum(row['pixels'] for row in comparison)==49766400
ordinary=read(w/'landing-comparison-final/comparison.json')
assert len(ordinary)==297 and sum(row['pixels'] for row in ordinary)==91238400
for row in comparison+ordinary:
    assert row['different']==row['max_channel_difference']==0 and row['difference_bbox'] is None
    assert len(row['regions'])==8 and all(r['different']==0 for r in row['regions'])

plan=read(w/'plan-g/plan.json')
assert plan['passed'] and len(plan['stages'])==24
assert not any(plan[k] for k in ['state_injection','seed_injection','passive_override'])
for name in ['plan.json','s3-alliance.png','s3-alliance.bin','after-capture.png','after-capture.bin','public-saved.png','public-saved.bin','public-loaded.png','public-loaded.bin']:
    assert digest(w/'plan-g'/name)==digest(w/'plan-h'/name),name
expected={row['id']:row['stage'] for row in plan['stages']}
gui_results={}
for mode in ['hd','original']:
    folder=w/('gui-'+mode+'-k')
    result=read(folder/'RESULT.json')
    assert result['passed'] and result['normal_entry'] and result['capture_turn']==7 and result['scene_count']==24
    assert not any(result[k] for k in ['state_injection','seed_injection','script_flag','passive_override'])
    assert result['six_tables_equal_public_plan']
    assert result['binary_sha256']==digest(w/'gin3-final-bases')['sha256']
    assert result['plan_sha256']==digest(w/'plan-g/plan.json')['sha256']
    assert result['audio_format']==[2,2,44100] and result['audio_peak']>32
    assert result['audio_sha256']==digest(folder/'audio.wav')['sha256']
    assert list(result['load_changes'])==['STELLAR']
    assert [d['offset'] for d in result['load_changes']['STELLAR']]==[3318,3319,3401,3527,3653]
    for slot in ['DT0','DT1']:
        for name in ['ADMIRAL','FLEET','SHIP','PLANET','STELLAR','WORK']:
            file=name+'.'+slot
            assert digest(folder/'save'/file)==digest(w/'plan-g/save'/file)==digest(w/'plan-h/save'/file),file
    scene_order=[]; frames={}; pictures={}; results=set(); sounds={}; playing=False; modes=set()
    with (folder/'FRAMES.jsonl').open() as f:
        for line in f:
            row=json.loads(line)
            if not row.get('planet'):continue
            stage,pb=row['planet'],row['playback']
            identity=pb['id']+'/'+str(stage['Kind'])
            assert stage==expected[identity],identity
            if identity not in frames:scene_order.append(identity);frames[identity]=[]
            n=pb['applied']
            if not frames[identity] or frames[identity][-1]!=n:frames[identity].append(n)
            pictures[identity]=len(pb.get('pictures') or [])
            sounds[identity]=pb.get('sounds',[])
            if pb['result']:results.add(identity)
            playing|=pb['sound_playing'];modes.add(row['graphics'])
    assert scene_order==list(expected),(mode,scene_order)
    assert results==set(expected) and playing
    for identity,sequence in frames.items():
        assert sequence==list(range(pictures[identity]+1)),(mode,identity,sequence,pictures[identity])
        assert all(s['played'] for s in sounds[identity]),(mode,identity,sounds[identity])
    assert result['f7_during_landing']==(mode=='hd')
    assert modes==({'hd','original'} if mode=='hd' else {'original'})
    for event in result['events']:
        if 'screenshot' in event:assert digest(folder/event['screenshot'])['sha256']==event['sha256']
    gui_results[mode]={'stages':24,'pictures':sum(pictures.values()),'frame_order_complete':True,'audio_peak':result['audio_peak'],'elapsed_seconds':result['elapsed_seconds'],'load_changes':result['load_changes']}
    paths.update(p for p in folder.rglob('*') if p.is_file() and not p.is_symlink())
standard=(w/'standard-final.log').read_text()
assert standard.count('ok  \t')==15 and standard.count('[no test files]')==4 and '\nFAIL' not in standard
negative=(w/'base-negative-test-final-2.log').read_text()
assert 'missing D0 borrowed an opposite-side base' in negative and '\nFAIL' in negative
assert 'ok  \t' in (w/'base-side-test-final-2.log').read_text()
for name in ['plan-g','plan-h','fortress-comparison-final','landing-comparison-final']:
    paths.update(p for p in (w/name).rglob('*') if p.is_file())
paths.update(w.glob('*.log'))
paths.update([w/'fortress-typed.json',root/'workplace/landing48-r110-typed-r3.json',w/'gin3-final-bases',w/'parent-head.txt',w/'images.txt',w/'planet_frame-before.go',w/'base-negative-overlay.json'])
for folder in ['cmd','internal']:
    paths.update((root/folder).rglob('*.go'))
paths.update(p for p in (root/'assets/hd').rglob('*') if p.is_file())
for name in ['tools/scripts/landing_player_plan.go','tools/scripts/landing_player_gui.py','tools/scripts/landing_fortress_native.go','tools/scripts/landing_fortress_prepare.py','tools/scripts/engagement_integration_receipt.py','tools/landing_fortress_oracle.sh','tools/landing_fortress_native_run.sh','tools/gui.sh','tools/go.sh','tools/py.sh','tools/with-xvfb.sh','docs/spec/118-landing-scene.md','docs/spec/133-hd-animation-surfaces.md','docs/internals/data/engagement-panel.json','docs/internals/data/landing-soundfont.json','go.mod','go.sum']:
    paths.add(root/name)
orig=Path('/orig')
originals={str(p.relative_to(orig)):digest(p) for p in sorted(orig.rglob('*')) if p.is_file()}
bundle=root/'org_game/银河英雄传说3SP.zip'
assert digest(bundle)['sha256']=='de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
if a.validate:
    print('PASS: 792原版雙跑檔；459完整幀；兩種顯示各24段／完整格序／音訊／公開六表；標準15 PASS／4無測試')
    raise SystemExit(0)
if a.freeze:
    assert not target.exists()
    index={'schema':'logh3.engagement-integration.v1','date':'2026-10-14','base_head':(w/'parent-head.txt').read_text().strip(),'source_dirty':True,'original_pairs':pair_count,'frames':459,'pixels':141004800,'gui':gui_results,'classification':'typed完整renderer同態與獨立正常remake GUI；不宣稱DOS整場同態或戰場續讀','original_bundle':{'path':str(bundle.relative_to(root)),**digest(bundle)},'originals':originals,'sources':{str(p.relative_to(root)):digest(p) for p in sorted(paths)}}
    with target.open('x') as f:json.dump(index,f,ensure_ascii=False,indent=2);f.write('\n')
index=read(target)
assert index['originals']==originals
for name,want in index['sources'].items():assert digest(root/name)==want,name
print('PASS:',pair_count,'原版雙跑檔；459完整幀；兩種顯示各24段／完整格序／音訊／公開六表；標準15 PASS／4無測試')
print('index',digest(target))
