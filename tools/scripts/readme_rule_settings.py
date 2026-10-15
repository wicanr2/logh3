"""從已驗 Windows 正常操作收據匯出 README 的開／關原尺寸截圖。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil

ROOT=Path('/ref')
OUTPUT=Path('/work/docs/images')
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check',action='store_true')
args=parser.parse_args()
parent=ROOT/'docs/internals/data/windows-rule-settings.json'
data=json.loads(parent.read_text())
gui=next(row for row in data['gui'] if row['platform']=='Windows Wine')
rows=[]
for source_name,target_name,enabled in [
    ('01-default-on.png','readme-v140-windows-settings.png',True),
    ('02-off-hd.png','readme-v140-windows-settings-off.png',False),
]:
    row=next(item for item in gui['screenshots'] if Path(item['file']['path']).name==source_name)
    source=ROOT/row['file']['path']
    raw=source.read_bytes()
    digest=hashlib.sha256(raw).hexdigest()
    assert len(raw)==row['file']['bytes'] and digest==row['file']['sha256']
    assert row['observed']['state']=='StateSettings' and row['observed']['graphics']=='hd'
    assert row['observed']['rules']==dict(flagship_transfer=enabled,hero_movement=enabled)
    target=OUTPUT/target_name
    if target.exists():
        assert target.stat().st_uid==target.stat().st_gid==os.getuid()==1000
        assert target.read_bytes()==raw,'拒絕覆寫不同內容的既有截圖'
    else:
        assert not args.check and OUTPUT.stat().st_uid==os.getuid()==1000
        shutil.copyfile(source,target)
    rows.append(dict(path='docs/images/'+target_name,bytes=len(raw),sha256=digest,
                     source=row['file'],observed=row['observed']))
manifest=dict(schema='logh3.readme-rule-settings.v1',version=data['version'],
    parent_receipt=dict(path='docs/internals/data/windows-rule-settings.json',
                        sha256=hashlib.sha256(parent.read_bytes()).hexdigest()),
    product_source_commit=data['verified_product_source_commit'],
    exporter='tools/scripts/readme_rule_settings.py',
    exporter_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
    method='複製已驗實際Windows ZIP的Wine正常GUI完整視窗，不裁切、不縮放、不重畫',
    scope='系統設定開／關與正常GUI證據；不是自然遊玩轉移旗艦事件或Windows真機截圖',
    screenshots=rows)
serialized=json.dumps(manifest,ensure_ascii=False,indent=2)+'\n'
path=OUTPUT/'readme-v140-rules.json'
if args.check:
    assert path.read_text()==serialized
else:
    assert not path.exists(),'拒絕覆寫既有來源索引'
    path.write_text(serialized)
print('PASS README兩張開／關原尺寸截圖、狀態與來源雜湊')
