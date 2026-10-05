"""核對第96輪標題原版收據、素材catalog及驗證來源；在tools/gui.sh容器執行。"""
from pathlib import Path
import ast,hashlib,json,sys

target=Path('docs/internals/data/title-asset-receipts.json')
catalog=Path('docs/internals/data/remaining-assets.json')
v=json.loads(catalog.read_text())
for r in v['images']:
    p=Path('assets/original')/r['path'];b=p.read_bytes()
    assert len(b)==r['bytes'] and hashlib.sha256(b).hexdigest()==r['sha256'],p
assert len(v['images'])==794
native=Path('workplace/oracle')
for a,b in [('title68-r96-a','title68-r96-b'),('title68-r96-left-a','title68-r96-left-b')]:
    for name in ['screen.bin','pal.bin']:
        assert (native/a/name).read_bytes()==(native/b/name).read_bytes(),(a,b,name)
title=(native/'title68-r96-a/screen.bin').read_bytes()
assert hashlib.sha256(title).hexdigest()=='addde5e0929eb5f99a5430a714b19041a2cbe5530a4720a3641abebbfa3bf8ae'
rgb=Path('/orig/SPTIT.RGB').read_bytes()
assert bytes(((x*4)<<2)|((x*4)>>4) for x in rgb)==(native/'title68-r96-a/pal.bin').read_bytes()[:48]
for run in ['title68-r96-right-a','title68-r96-key-a','title68-r96-bios-a']:
    assert (native/run/'screen.bin').read_bytes()==title,run
ast.parse(Path('tools/scripts/assets_manifest.py').read_text())
assert '還沒匯出成圖的' not in Path('tools/scripts/assets_manifest.py').read_text()

if sys.argv[1]=='--freeze':
    assert not target.exists()
    standard=Path('workplace/round96-standard-gui.log').read_text()
    assert standard.count('\nok  ')+int(standard.startswith('ok  '))==13
    assert standard.count('[no test files]')==5 and 'FAIL' not in standard
    paths={str(catalog),'docs/spec/111-title-image.md','docs/re/430-title-image.md',
        'docs/playtest/085-original-asset-exports.md','docs/goal/2026-10-05-title-player-round-96.md',
        'tools/scripts/title_asset_receipt.py','tools/scripts/remaining_assets_export.go',
        'tools/scripts/assets_manifest.py','internal/asset/title.go','internal/asset/title_test.go',
        'internal/asset/export_catalog_test.go','internal/asset/intro.go','internal/asset/spr.go',
        'internal/asset/amg.go','internal/asset/cmp.go','internal/screen/intro.go',
        'docs/internals/data/screen-main-menu.json','assets/original/README.md','assets/original/MANIFEST.tsv',
        'workplace/round96-standard.log','workplace/round96-standard-gui.log'}
    for p in native.glob('title68-r96-*'):
        paths.update(str(x) for x in p.rglob('*') if x.is_file())
    for pattern in ['title68-r96*','assets-remaining-r96*']:
        for p in Path('workplace').glob(pattern):
            if p.is_file():paths.add(str(p))
            elif p.is_dir():paths.update(str(x) for x in p.rglob('*') if x.is_file())
    sources={}
    for name in sorted(paths):
        b=Path(name).read_bytes();sources[name]={'bytes':len(b),'sha256':hashlib.sha256(b).hexdigest()}
    receipt={'schema':'logh3.title-asset-receipts.v1','date':'2026-10-05',
        'parent_commit':'51d3d7ea7b985f4dc15b1b5cd317daca7bb703f0',
        'dosgolem_head':'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c','dosgolem_dirty':False,
        'go_image':'golang:1.24-bookworm','gui_image':'logh3-ebiten:2',
        'images':794,'png_bytes':5853987,'original_files':44,'palette_json_files':1,
        'title_pixels':224000,'title_pixel_differences':0,'title_palette_differences':0,
        'title_repeats':2,'left_to_main_menu_repeats':2,
        'player_title_implemented':False,'spec_111_scope':'source image only',
        'standard_tests':{'packages':18,'pass':13,'no_tests':5},'sources':sources}
    with target.open('x') as f:f.write(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
elif sys.argv[1]!='--check':raise SystemExit('用法：--freeze／--check')
for name,expected in json.loads(target.read_text())['sources'].items():
    b=Path(name).read_bytes()
    assert len(b)==expected['bytes'] and hashlib.sha256(b).hexdigest()==expected['sha256'],name
print(len(v['images']),'張PNG及',len(json.loads(target.read_text())['sources']),'項來源符合；索引SHA',hashlib.sha256(target.read_bytes()).hexdigest())
