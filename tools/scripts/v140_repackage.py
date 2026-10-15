"""141：使用者明確授權v.1.4.0同版覆寫；舊身份備份、驗收及來源收據。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

from PIL import Image

ROOT=Path('/src')
VERSION='v.1.4.0-20261015'
OUT=ROOT/'dist-all'/VERSION
BATCH=ROOT/'workplace/settings141'
BACKUP=BATCH/'published-before-overwrite'
INDEX=ROOT/'docs/internals/data/v140-repackage.json'


def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as f:
        for block in iter(lambda:f.read(1024*1024),b''):h.update(block)
    return h.hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,data):p.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')


def archive():
    assert not BACKUP.exists() and BATCH.stat().st_uid==os.getuid()
    old=read(OUT/'SHA256SUMS-platforms.json')
    for row in old['files']:
        p=OUT/row['path']
        assert p.stat().st_uid==os.getuid() and sha(p)==row['sha256'],p
    remote=read(BATCH/'RELEASE-BEFORE.json')
    assert remote['tag_name']==VERSION and not remote['draft'] and not remote.get('immutable',False)
    tag=subprocess.check_output(['git','rev-parse',VERSION],text=True).strip()
    commit=subprocess.check_output(['git','rev-parse',VERSION+'^{commit}'],text=True).strip()
    assert (BATCH/'TAG-REMOTE-BEFORE.txt').read_text().split()[0]==tag
    downloads=ROOT/'workplace/promo139/release-download'
    expected={Path(r['path']).name:r['sha256'] for r in old['files'] if r['path'].startswith('release/')}
    expected['SHA256SUMS.txt']=sha(OUT/'release/SHA256SUMS.txt')
    assert set(expected)=={r['name'] for r in remote['assets']}
    for name,digest in expected.items():assert sha(downloads/name)==digest
    assert shutil.disk_usage(ROOT).free>5*1024**3,'空間不足以安全重建'
    BACKUP.mkdir()
    mapping=[]
    for source in [OUT,ROOT/'workplace'/('windows-package-'+VERSION)]:
        assert source.is_dir() and source.stat().st_uid==os.getuid()
        target=BACKUP/source.relative_to(ROOT)
        target.parent.mkdir(parents=True,exist_ok=True)
        shutil.move(str(source),target)
        mapping.append(dict(original=str(source.relative_to(ROOT)),archive=str(target.relative_to(ROOT))))
    shutil.copytree(downloads,BACKUP/'release-download')
    shutil.copyfile(BATCH/'RELEASE-BEFORE.json',BACKUP/'RELEASE-BEFORE.json')
    shutil.copyfile(ROOT/'docs/internals/data/v140-release-promo.json',BACKUP/'INDEX-BEFORE.json')
    (BACKUP/'TAG-OBJECT.txt').write_bytes(subprocess.check_output(['git','cat-file','tag',tag]))
    write(BACKUP/'MAPPING.json',dict(version=VERSION,authorization='使用者明確回答「複寫已發布的」',
        old_tag=tag,old_commit=commit,old_release_id=remote['id'],previous_packages=old,
        remote_assets=expected,mapping=mapping))
    print('PASS 已備份舊七包、完整目錄、來源、五下載附件、Release與tag身份',flush=True)


def promo():
    source=BACKUP/'dist-all'/VERSION/'promo'
    target=OUT/'promo'
    assert target.is_dir() and not list(target.iterdir())
    shutil.copytree(source,target,dirs_exist_ok=True)
    sums=read(target/'SHA256SUMS.json')
    for row in sums['files']:
        name=Path(row['path']).name
        assert sha(target/name)==sha(source/name)==row['sha256'],name
    print('PASS 影片與16證據檔逐位元組保留，不重拍或換音軌',flush=True)


def wine():
    target=ROOT/'workplace/music69-windows/windows141-delivery'
    assert not target.exists() and target.parent.stat().st_uid==os.getuid()
    target.mkdir()
    for kind,filename in [('full','full.zip'),('release','engine.zip')]:
        source=OUT/kind/('LOGH3-'+VERSION+'-windows-amd64.zip')
        shutil.copyfile(source,target/filename)
        assert sha(source)==sha(target/filename)
    print('READY Windows正式ZIP獨立複本',flush=True)


def receipt():
    assert not INDEX.exists()
    packages=read(OUT/'SHA256SUMS-platforms.json')
    for row in packages['files']:
        p=OUT/row['path']
        assert sha(p)==row['sha256'] and p.stat().st_size==row['bytes']
    build=read(OUT/'BUILD.json')
    for name,row in build['source_files'].items():
        if not name.startswith('vendor/'):assert sha(ROOT/name)==row['sha256'],name
    files=[OUT/'smoke/full-linux-r2/SMOKE.json',OUT/'smoke/release-linux/SMOKE.json',OUT/'smoke/macos/MACOS.json',
           ROOT/'workplace/music69-windows/windows141-delivery/RESULT.json',
           BATCH/'package-gui-f0-r2/RESULT.json',ROOT/'workplace/music69-windows/windows141-gui-f1-r2/RESULT.json']
    for p in files:assert p.is_file() and p.stat().st_uid==os.getuid()
    frames={}
    for p in files[-2:]:
        result=read(p)
        assert result['passed'] and result['show_enemies_default'] is False and result['show_enemies_functional']
        assert result['units_unchanged'] and result['turn_unchanged']
        image_path=p.parent/'settings-on.png'
        with Image.open(image_path) as raw:
            assert raw.size==(960,720)
            image=raw.convert('RGB')
        metal=(130,164,203) if result['faction']==0 else (185,144,80)
        assert all(image.getpixel((240,y))==metal and image.getpixel((242,y))==(0,0,0)
                   for y in range(66,596)),image_path
        frames[str(image_path.relative_to(ROOT))]=dict(sha256=sha(image_path),rows=530,
            metal_edge_continuous=True,inner_shadow_preserved=True)
    write(INDEX,dict(version=VERSION,authorization='使用者明確授權同版覆寫Release、附件與tag',
        source_commit=build['head'],source_files=build['source_files'],packages=packages,
        backup=read(BACKUP/'MAPPING.json'),backup_mapping_sha256=sha(BACKUP/'MAPPING.json'),
        platform_receipts={str(p.relative_to(ROOT)):sha(p) for p in files},
        generator_sha256={n:sha(ROOT/'tools/scripts'/n) for n in ['v140_repackage.py','v140_platform_delivery.py',
            'windows_hotfix_package.py','windows_delivery_smoke.py','v140_release_verify.py','compact_settings_gui.py']},
        font=read(ROOT/'internal/font/settings_font.SOURCE.json'),
        border_receipt_sha256=sha(ROOT/'docs/images/readme-settings-frame.json'),
        package_frame_checks=frames,
        promo=dict(preserved=True,manifest=read(OUT/'promo/SHA256SUMS.json')),
        tests=dict(command='tools/gui.sh test ./internal/... ./cmd/...',passed=15,no_test_files=4),
        limits=['Windows為Wine，macOS僅雙架構／依賴／ARM簽章與ZIP，未真機驗收',
                '#48正式戰術讀回重建與DOS整段同狀態保持OPEN']))
    print('PASS 七封包與來源／平台／影片／授權收據，已建立本次同版重包索引',flush=True)


if __name__=='__main__':
    assert os.getuid()==os.getgid()==1000 and Path.cwd()==ROOT
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=['archive','promo','wine','receipt'])
    args=p.parse_args()
    globals()[args.mode]()
