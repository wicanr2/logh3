"""141：逐行核對設定視窗金屬外緣及內側陰影，保存正常操作原尺寸截圖。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil

from PIL import Image

ROOT=Path('/src')
BATCH=ROOT/'workplace/settings141'
DEST=ROOT/'docs/images'


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())


p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--check',action='store_true')
args=p.parse_args()
themes=read(ROOT/'assets/hd/interface-themes.json')['themes']
receipts,checks,files={},{},[]
for faction,theme in [(0,'empire'),(1,'alliance')]:
    folder=BATCH/('f'+str(faction)+'-r2')
    result=read(folder/'RESULT.json')
    assert result['passed'] and result['normal_entry'] and result['show_enemies_default'] is False
    assert result['show_enemies_functional'] and result['turn_unchanged'] and result['units_unchanged']
    assert result['binary_sha256']==sha(ROOT/'workplace/settings141-gin3-r2')
    receipts[str((folder/'RESULT.json').relative_to(ROOT))]=sha(folder/'RESULT.json')
    captures={e['name']:e for e in result['events'] if e.get('kind')=='capture'}
    for name,event in captures.items():
        state=event['observed']
        if state['state']!='StateSettings':continue
        source=folder/(name+'.png')
        assert sha(source)==event['sha256']
        with Image.open(source) as raw:
            assert raw.size==(960,720)
            image=raw.convert('RGB')
        # 640×480正常視窗的1.5倍顯示，來源矩形由正式唯讀觀測取得。
        bounds=state['settings']['bounds']
        x=round(bounds[0]*1.5)
        ys=range(round(bounds[1]*1.5)+4,round(bounds[3]*1.5)-4)
        metal=themes[theme]['metal'] if state['graphics']=='hd' else '#b2b2c3'
        want=tuple(bytes.fromhex(metal.lstrip('#')))
        assert all(image.getpixel((x,y))==want for y in ys),(theme,name,'外緣斷線')
        assert all(image.getpixel((x+2,y))==(0,0,0) for y in ys),(theme,name,'內側陰影遺失')
        checks[theme+'/'+name]=dict(graphics=state['graphics'],rows=len(ys),outer_colour=metal,
                                      metal_edge_continuous=True,inner_shadow_preserved=True)
    for capture,label in [('settings-on',theme)]+([('ai-menu','ai')] if faction==0 else []):
        source=folder/(capture+'.png')
        target=DEST/('readme-settings-frame-'+label+'.png')
        if not args.check:
            assert not target.exists() and DEST.stat().st_uid==1000
            shutil.copyfile(source,target)
        assert sha(target)==sha(source)
        files.append(dict(file=target.name,sha256=sha(target),bytes=target.stat().st_size,
                          source=str(source.relative_to(ROOT)),observed=captures[capture]['observed']))
manifest=dict(source='141：原始碼正常Linux片頭至S1設定，非舊v.1.4.0 ZIP',
              binary_sha256=sha(ROOT/'workplace/settings141-gin3-r2'),gui_receipts=receipts,
              frame_checks=checks,files=files,
              source_sha256={n:sha(ROOT/n) for n in ['cmd/gin3/compact_settings.go',
                  'internal/presentation/faction.go','assets/hd/interface-themes.json',
                  'internal/font/settings_font.SOURCE.json','tools/scripts/settings_border_receipt.py']})
target=DEST/'readme-settings-frame.json'
if args.check:assert read(target)==manifest
else:
    assert not target.exists()
    target.write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print('PASS',len(checks),'張正常設定／選單，金屬外緣逐行連續，內側陰影完整；README三張原圖'+('獨立核對' if args.check else '已匯出'))
