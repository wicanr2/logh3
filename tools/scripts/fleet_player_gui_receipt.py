"""第79輪：凍結真實X11、公開Game計畫及診斷保存的來源索引。"""
from pathlib import Path
import hashlib
import json
import os

os.chdir('/src')
target=Path('docs/internals/data/fleet-player-gui.json')
assert not target.exists() and target.parent.stat().st_uid==os.getuid()
files={}
def read(path):
    p=Path(path)
    assert p.is_file() and p.stat().st_uid==os.getuid(),str(p)
    b=p.read_bytes()
    files[str(p)]={'sha256':hashlib.sha256(b).hexdigest(),'bytes':len(b)}
    return b
def data(path):return json.loads(read(path))

parents={
    'docs/internals/data/fleet-command-display-game.json':'830e5c5615be977c5e18cff23fe990c18cc6ee0ad01802b17562a8338ee33405',
    'docs/internals/data/fleet-sound-game.json':'4388203c2d6d9c69f69c38f1107b871b92c1397923f38a9eb2b7fe72ed39b9e3',
}
for path,want in parents.items():
    assert hashlib.sha256(read(path)).hexdigest()==want,path
image=data('workplace/round79-gui-image.json')
cases=[]
for side in [0,1]:
    planroot=Path(f'workplace/round79-plan-final-{side}')
    plan=data(planroot/'plan.json')
    assert plan['six_saves_equal_parent'] and plan['faction']==side
    seeds=[a['seed'] for a in plan['actions'] if a['kind']=='seed']
    assert seeds and set(seeds)=={0x12345678}
    for p in sorted(planroot.iterdir()):
        if p.is_file():read(p)
    for p in sorted(Path(f'workplace/round77-chain-c/normal-{side}').iterdir()):
        if p.suffix=='.DT0' or p.name=='inputs.json':read(p)
    for scale in [1,2]:
        batch=f'round79-gui-complete-{side}-{scale}'
        root=Path('workplace')/batch
        receipt=data(root/'receipt.json')
        assert receipt['faction']==side and receipt['scale']==scale
        assert receipt['six_diagnostic_saves_equal'] and receipt['intro_default']
        assert not receipt['script_flag_used'] and not receipt['sound_device_tested']
        assert len(receipt['actions'])==len(plan['actions'])
        assert len(receipt['comparisons'])==3
        comparisons=[]
        for c in receipt['comparisons']:
            name=c['name']
            assert c['pixels']==307200 and c['x11_different']==0 and c['x11_bbox'] is None and c['game_index_equal']
            raw=data(root/(name+'-x11-comparison.json'))
            native=data(root/(name+'-native-comparison.json'))
            assert raw['width']==640*scale and raw['height']==480*scale
            assert raw['depth']==24 and raw['bits_per_pixel']==32 and raw['max_channel_difference']==0
            assert native['native_pixels']==307200*scale*scale
            assert native['native_different']==0 and native['native_bbox'] is None and native['max_channel_difference']==0
            assert read(root/(name+'.bin'))==read(planroot/(name+'.bin'))
            comparisons.append({'name':name,'logical':raw,'native':native,'game_index_equal':True})
        saves=sorted(root.glob('*.DT0'))
        assert len(saves)==6
        for p in saves:
            assert read(p)==read(planroot/p.name)==read(Path(f'workplace/round77-chain-c/normal-{side}')/p.name),p
        for p in sorted(root.iterdir()):
            if p.is_file():read(p)
        log=read('workplace/'+batch+'.log').decode()
        assert 'Traceback' not in log and '六份診斷保存相同' in log
        assert log.count('原生全圖 0')==3
        cases.append({'faction':side,'scale':scale,'batch':str(root),'command':f'bash tools/fleet_player_gui.sh {batch} {side} {scale}',
                      'canonical_actions':len(receipt['actions']),'additional_camera_gestures':len(receipt['camera_gestures']),
                      'input_mapping':receipt['input_mapping'],'comparisons':comparisons,'six_diagnostic_saves_equal_plan_and_round77':True,
                      'final':data(root/'status.json')})

standard=read('workplace/round79-standard.log').decode()
assert '\nFAIL' not in standard
assert sum(x.startswith('ok ') for x in standard.splitlines())==13
assert sum('[no test files]' in x for x in standard.splitlines())==5
for path in ['go.mod','go.sum','cmd/gin3/main.go','cmd/gin3/effects.go','internal/game/fleetcombat.go',
             'internal/game/battle.go','internal/game/fleetcombatcommand.go',
             'internal/game/fleetcombatscene.go','internal/game/soundeffects.go',
             'tools/gui.sh','tools/fleet_player_gui.sh','tools/fleet_player_gui_xrun.sh',
             'tools/scripts/fleet_player_gui_prepare.py','tools/scripts/fleet_player_gui_plan.go',
             'tools/scripts/fleet_player_gui_driver.py','tools/scripts/fleet_player_gui_capture.go',
             'tools/scripts/fleet_player_gui_native_check.go','tools/scripts/fleet_player_gui_input.go',
             'tools/scripts/fleet_player_gui_receipt.py','workplace/round79-fleet-gui-game-v3.go',
             'workplace/round79-fleet-gui-main-v3.go','workplace/round79-fleet-gui-overlay-v3.json',
             'workplace/round79-gin3-v3','workplace/round79-input-v2','workplace/round79-capture','workplace/round79-native-check']:
    read(path)
result={'schema':'logh3-fleet-player-gui-v1','date':'2026-10-04',
        'production_parent_commit':'838e9a125be40c411eca78d1635e52324fd1a5e4','production_sources_modified':False,
        'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','dosgolem_unchanged':True,
        'image':{'tag':'logh3-ebiten:2','id':image['Id']},
        'classification':'正常remake主選單／138／S3／實行／戰術命令／交戰返回的真實X11；比較公開Game計畫，沒有新增原版same-state',
        'seed':{'value_hex':'12345678','when':'各戰場執行前','method':'test-only RND overlay，沒有位置、命令或RAM注入'},
        'cases':cases,'standard_tests':{'pass':13,'no_test_files':5},
        'limits':['原版七張命令完整圖各一像素未核准，沿066與#89，沒有遮罩或升格',
                  '右鍵以實際鏡頭回授使用額外X11手勢，完整記錄每個Update與相對時間，非同一實體時序',
                  '短暫自動階段以逐Update原始紀錄核曾實際抵達，不要求低頻主機輪詢看見',
                  '六份保存使用診斷SaveSlot API，不是玩家存檔選單的驗收',
                  '本輪未指定音樂目錄，沒有聲音裝置或人耳驗收',
                  'Linux Xvfb建置，沒有macOS真機或正式發行包驗收'],
        'files':files}
target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('GUI收據',len(cases),'案／12完整圖，來源',len(files),'SHA-256',hashlib.sha256(target.read_bytes()).hexdigest())
