"""#53：正式 GUI 正常戰場入口、實行一次及獨立存讀抽樣；不宣稱原版八步 parity。"""
import argparse
import hashlib
import json
import os
import struct
import subprocess
from pathlib import Path

p=argparse.ArgumentParser()
p.add_argument('--head',required=True)
p.add_argument('--dirty',required=True)
p.add_argument('--image-id',required=True)
p.add_argument('--output',default='/src/workplace/oracle/battle53-routines')
args=p.parse_args()
out=Path(args.output)
assert out.is_dir() and out.stat().st_uid==os.getuid()
prefix='move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 544 125; click 560 125; tick 400; click 300 240; tick 2'
save_path=out/'save'
save_path.mkdir(exist_ok=False)
save_script='move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 544 125; click 560 125; tick 18; click 544 125; click 560 165; click 300 142; rclick 300 142; click 544 125; click 560 145; click 300 122; click 544 125; click 560 165; click 300 162'
cases=[('normal-entry',prefix,'StateBattle',[]),
       ('normal-turn',prefix+'; click 240 240; click 240 240','StateBattle',[]),
       ('normal-save',save_script,'StateSaveGame',['-save',str(save_path)])]
subprocess.run(['go','build','-o','/tmp/logh3-battle53','./cmd/gin3'],check=True)
def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()
result={'scope':'remake 正常 GUI 入口、戰術實行、正式存讀抽樣；不宣稱同狀態原版畫面或八步 parity',
        'head':args.head,'dirty':args.dirty,'image':'logh3-ebiten:2','image_id':args.image_id,
        'cases':[],'source_sha256':{n:sha(Path('/src')/n) for n in
        ['internal/rules/battle.go','internal/rules/planetassault.go','internal/rules/fortressgun.go',
         'internal/rules/strategicassault.go','tools/scripts/battle_rules_gui_check.py']}}
for name,script,state,extra in cases:
    paths={ext:out/(name+'.'+ext) for ext in ('png','bin','txt','log','script')}
    assert not any(path.exists() for path in paths.values()),'不可覆寫既有收據'
    paths['script'].write_text(script+'\n')
    cmd=['/tmp/logh3-battle53','-orig','/orig','-data','/src/docs/internals/data','-intro=false',
         '-script',script,'-shot',str(paths['png']),'-shot-bin',str(paths['bin']),
         '-state-out',str(paths['txt']),'-shot-after','2']+extra
    with paths['log'].open('w') as log:
        subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=120)
    text=paths['txt'].read_text()
    assert 'state='+state+'\n' in text,text
    assert 'window=640x480\n' in text,text
    assert paths['bin'].stat().st_size==640*480
    image=paths['png'].read_bytes()
    assert image[:8]==b'\x89PNG\r\n\x1a\n' and struct.unpack('>II',image[16:24])==(640,480)
    result['cases'].append({'name':name,'command':cmd,'state':text,
                            'sha256':{ext:sha(path) for ext,path in paths.items()}})
    print(name,state,'通過',flush=True)
assert sha(out/'normal-entry.bin')!=sha(out/'normal-turn.bin'),'實行後完整畫面應改變'
files={s+'.DT'+ext:(save_path/(s+'.DT'+ext)).read_bytes() for s in ['STELLAR','PLANET','ADMIRAL','FLEET','SHIP','WORK'] for ext in '01'}
for stem,size in [('FLEET',3904),('SHIP',27456)]:
    assert len(files[stem+'.DT0'])==size and files[stem+'.DT0']==files[stem+'.DT1'],stem
assert files['PLANET.DT0']==files['PLANET.DT1']
fleet=files['FLEET.DT0']; counts=[0,0]
for i in range(32):
    rec=fleet[i*122:(i+1)*122]
    if rec[4]!=255: counts[rec[1]]+=1
assert counts==[1,3],counts
result['save']={'scope':'正常新遊戲編隊、回合、存、讀、再存；沒有保存中途戰場的聲明',
                'fleet_counts':counts,'fleet_bytes':3904,'ship_bytes':27456,'different':0,
                'sha256':{n:hashlib.sha256(b).hexdigest() for n,b in files.items()}}
for path in out.rglob('*'):
    assert path.stat().st_uid==os.getuid() and path.stat().st_gid==os.getgid(),path
with (out/'gui-check.json').open('x') as f: json.dump(result,f,ensure_ascii=False,indent=2)
print('三份正式 GUI、完整 FLEET／SHIP 存讀通過；輸出擁有權正確。',flush=True)
