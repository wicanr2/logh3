"""147：匯出正式Linux ZIP正常操作的README原尺寸截圖及來源。Docker內執行。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil

from PIL import Image

ROOT=Path('/src')
SOURCE=ROOT/'workplace/promo139/compact-delivery-f0'
OUTPUT=ROOT/'docs/images'
VERSION='v.1.4.0-20261015'


def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()


p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--check',action='store_true')
args=p.parse_args()
result=json.loads((SOURCE/'RESULT.json').read_text())
assert result['passed'] and result['show_enemies_default'] is False and result['show_enemies_functional']
assert result['ai_levels']==[-1,0,1,2,3] and result['units_unchanged'] and result['turn_unchanged']
captures={e['name']:e for e in result['events'] if e.get('kind')=='capture'}
files=[]
for source,name,kind in [('settings-on','readme-v140-compact-settings.png','設定：敵軍開啟、AI強化1、兩規則開啟'),
                         ('ai-menu','readme-v140-compact-ai.png','AI原版／強化0至3選單'),
                         ('battle-enemies-on-again','readme-v140-compact-enemies.png','實際戰場敵方分隊顯示')]:
    event=captures[source];state=event['observed'];path=SOURCE/(source+'.png');target=OUTPUT/name
    assert sha(path)==event['sha256'] and state['show_enemies'] is True
    if source!='battle-enemies-on-again':
        assert state['state']=='StateSettings' and state['ai_level']==1
        assert state['rules']==dict(flagship_transfer=True,hero_movement=True)
        assert state['settings']['glyph']==[12,12,6,12]
    if source=='ai-menu':assert state['ai_menu'] is True
    with Image.open(path) as image:assert image.size==(960,720)
    if not args.check:
        if target.exists():assert sha(target)==sha(path),'拒絕覆寫其他來源圖'
        else:shutil.copyfile(path,target)
    assert sha(target)==sha(path)
    files.append(dict(file=name,bytes=target.stat().st_size,sha256=sha(target),source=str(path.relative_to(ROOT)),
        role=kind,size=[960,720],show_enemies=True,ai_level=state['ai_level'],graphics=state['graphics'],
        settings=state.get('settings'),rules=state['rules']))
manifest=dict(version=VERSION,source='修正後實際Linux完整ZIP，正常片頭至S1、系統設定及戰場，無注入',
    source_commit=json.loads((ROOT/'workplace'/('package92-'+VERSION+'-release139')/'INPUTS.json').read_text())['head'],
    binary_sha256=result['binary_sha256'],source_result_sha256=sha(SOURCE/'RESULT.json'),
    full_zip_sha256=sha(ROOT/'dist-all'/VERSION/'full'/('LOGH3-'+VERSION+'-linux-amd64.zip')),
    show_enemies_default=False,enemy_visibility_test=dict(target=result['target'],box=result['target_screen_box'],
        changed_channels=result['target_changed_channels'],units_unchanged=True,turn_unchanged=True),
    files=files,generator_sha256=sha(ROOT/'tools/scripts/readme_compact_settings.py'),
    transformation='完整原尺寸複製，不裁切、不縮放、不改圖')
encoded=json.dumps(manifest,ensure_ascii=False,indent=2)+'\n'
index=OUTPUT/'readme-v140-compact.json'
if args.check:assert index.read_text()==encoded
else:
    assert not index.exists()
    index.write_text(encoded)
print('PASS README三張原尺寸正式ZIP截圖、AI／敵軍狀態及SHA-256')
