"""153：獨立核對三平台完整包、實際AppImage及正常玩家收據。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import struct
import zipfile
from PIL import Image,ImageChops

import full_appimage_release as build
from local_test_macos_package import universal

p=argparse.ArgumentParser(description=__doc__);p.add_argument('--check',action='store_true');a=p.parse_args()
ROOT,OUT,STAGE,BATCH,VERSION=build.ROOT,build.OUT,build.STAGE,build.BATCH,build.VERSION
read,write,sha=build.read,build.write,build.sha
record=lambda f:dict(path=str(f.relative_to(ROOT)),bytes=f.stat().st_size,sha256=sha(f))
build.configure();inputs=build.previous.source_check()
old=read(OUT/'BUILD.json')
assert set(old['source_files'])==set(inputs['source_files'])
assert all(old['source_files'][n]==v['sha256'] for n,v in inputs['source_files'].items())
packages=[];architectures={};assets_by_platform={}
for name,platform in [('WINDOWS','windows-amd64'),('MACOS','macos-universal')]:
    result=read(STAGE/(name+'.json'));row=result['package'];file=OUT/row['path'];folder=file.with_suffix('')
    assert sha(file)==row['sha256'] and file.stat().st_size==row['bytes']
    manifest=read(folder/'MANIFEST.json');assert manifest['version']==VERSION
    with zipfile.ZipFile(file) as z:
        assert z.testzip() is None
        assert set(z.namelist())=={folder.name+'/'+n for n in [*manifest['files'],'MANIFEST.json']}
        for n,info in manifest['files'].items():
            entry=z.getinfo(folder.name+'/'+n);value=z.read(entry)
            assert not Path(entry.filename).is_absolute() and '..' not in Path(entry.filename).parts
            assert entry.filename.isascii() or entry.flag_bits&0x800
            assert len(value)==info['bytes'] and hashlib.sha256(value).hexdigest()==info['sha256']
            assert sha(folder/n)==info['sha256']
            assert (entry.external_attr>>16)&0o777==int(info['mode'],8)
    assert read(folder/'VERSION.json')['version']==VERSION
    notes=(folder/'README.txt').read_bytes();assert notes.startswith(b'\xef\xbb\xbf') and '思源黑體' in notes.decode('utf-8-sig')
    assets=folder if name=='WINDOWS' else folder/'銀河英雄傳說III.app/Contents/Resources'
    assets_by_platform[platform]=assets
    for n,v in inputs['common_files'].items():assert sha(assets/n)==v['sha256'],(platform,n)
    assert not any(f.name=='preferences.json' for f in folder.rglob('*'))
    if name=='WINDOWS':
        assert b'\n' not in notes.replace(b'\r\n',b'')
        binary=folder/'gin3.exe';raw=binary.read_bytes();offset=struct.unpack_from('<I',raw,0x3c)[0]
        assert raw[:2]==b'MZ' and raw[offset:offset+4]==b'PE\0\0' and struct.unpack_from('<H',raw,offset+4)[0]==0x8664
        assert sha(binary)==old['commands'][0]['sha256']
        installed=ROOT/'workplace/music69-windows/full153'/folder.name
        for n,v in manifest['files'].items():assert sha(installed/n)==v['sha256'],n
        for mode in ['original','enhanced']:
            content=(folder/('start-'+mode+'.cmd')).read_bytes();assert content.isascii() and b'\n' not in content.replace(b'\r\n',b'')
        architectures[platform]='PE amd64，與152輪已驗程式逐位元組相同'
    else:
        contents=folder/'銀河英雄傳說III.app/Contents'
        info=plistlib.loads((contents/'Info.plist').read_bytes())
        assert info['LOGH3FullVersion']==VERSION and info['CFBundleVersion']==info['CFBundleShortVersionString']=='1.6.1'
        binary=contents/'MacOS/gin3';slices=universal(binary);assert slices==result['slices']
        assert {v['architecture'] for v in slices}=={'amd64','arm64'}
        assert all(v['deployment']==[{'macos':'11.0.0','sdk':'15.5.0'}] for v in slices)
        assert next(v for v in slices if v['architecture']=='arm64')['code_directory'][0]['all_page_hashes_match']
        for f in [contents/'MacOS/launch',binary,folder/'start-original.command',folder/'start-enhanced.command']: assert os.access(f,os.X_OK)
        architectures[platform]=slices
    assert sha(binary)==result['binary_sha256']
    packages.append(dict(**row,binary_sha256=sha(binary),manifest_sha256=sha(folder/'MANIFEST.json')))

app=read(OUT/'full/APPIMAGE.json');appfile=OUT/app['package']['path']
assert sha(appfile)==app['package']['sha256'] and appfile.stat().st_size==app['package']['bytes'] and os.access(appfile,os.X_OK)
with appfile.open('rb') as f: prefix=f.read(app['offset'])
assert prefix[:4]==b'\x7fELF' and prefix[8:11]==b'AI\x02' and hashlib.sha256(prefix).hexdigest()==app['runtime_sha256']
appmanifest=read(BATCH/'APPDIR-MANIFEST.json')
assert sha(BATCH/'APPDIR-MANIFEST.json')==app['appdir_manifest_sha256']
extracted=BATCH/'extracted/squashfs-root';assert build.package.inventory(extracted)==appmanifest['files']
assert build.package.inventory(BATCH/'AppDir')==appmanifest['files']
assets=extracted/'usr/bin';assets_by_platform['linux-x86_64']=assets
assert sha(assets/'gin3')==read(STAGE/'LINUX.json')['binary_sha256']==old['commands'][1]['sha256']
raw=(assets/'gin3').read_bytes();assert raw[:6]==b'\x7fELF\x02\x01' and struct.unpack_from('<H',raw,18)[0]==62
for n,v in inputs['common_files'].items(): assert sha(assets/n)==v['sha256'],n
assert not any(f.name=='preferences.json' for f in extracted.rglob('*'))
assert extracted.joinpath('AppRun').stat().st_mode&0o111
assert extracted.joinpath('.DirIcon').is_symlink() and os.readlink(extracted/'.DirIcon')=='logh3.svg'
assert len(app['dependencies']['libraries'])==15
assert max(tuple(map(int,v.split('.'))) for v in app['dependencies']['glibc_requirements'].values() if v)==(2,34)
architectures['linux-x86_64']='Type-2 AppImage，ELF64 x86_64，實際ELF最大GLIBC_2.34需求，15附帶函式庫'
packages.append(dict(**app['package'],binary_sha256=sha(assets/'gin3'),appdir_files=len(appmanifest['files']),runtime_sha256=app['runtime_sha256']))

normal=[];diagnostics=[];saves=[]
for platform,directory in [('linux-x86_64',BATCH/'player-linux'),('windows-amd64',ROOT/'workplace/music69-windows/full153/player-windows')]:
    result=read(directory/'RESULT.json');assert result['passed'] and result['normal_entry']
    assert not any(result[k] for k in ['state_injection','seed_injection','script_flag'])
    assert result['driver_sha256']==sha(BATCH/'pointer_toolbar_gui.py')
    assert len(result['events'])==31 and len(result['saved'])==6
    expected=sha(appfile) if platform=='linux-x86_64' else old['commands'][0]['sha256']
    assert result['binary_sha256']==expected
    if platform=='linux-x86_64':
        assert result['appimage'] and result['automatic_resources']
        assert not set(['-save','-orig','-data','-art']).intersection(result['command'])
        save_dir=directory/'config/logh3/save';assert result['save_dir']==str(save_dir)
    else:save_dir=directory/'save'
    for n,h in result['saved'].items():assert sha(save_dir/n)==h
    saves.append(result['saved'])
    for event in result['events']:
        f=directory/event['screenshot'];assert sha(f)==event['sha256']
        with Image.open(f) as im: assert im.size==tuple(event['window']);im.verify()
    events={v['name']:v for v in result['events']}
    assert events['loaded']['observed']['font_style']=='sans' and events['loaded']['observed']['ai_level']==3
    assert result['performance']['uploads_before']==result['performance']['uploads_after']
    for code in ['zh-TW','zh-CN','ja','en','ko']:
        assert events['font-'+code]['observed']['locale']==events['ai-'+code]['observed']['locale']==code
    center=Image.open(directory/'poster-map.png').convert('RGB').crop((160,0,1120,720))
    plain=Image.open(directory/'map-sans.png').convert('RGB');assert ImageChops.difference(center,plain).getbbox() is None
    for off,on,box in [('map-off','button-hover',(565,122,607,138)),('menu-off','menu-hover',(565,162,590,178))]:
        roi=tuple(int(v*1.5) for v in box)
        delta=ImageChops.difference(Image.open(directory/(off+'.png')).convert('RGB').crop(roi),Image.open(directory/(on+'.png')).convert('RGB').crop(roi))
        assert delta.getbbox() is not None
    normal.append(dict(platform=platform,**record(directory/'RESULT.json'),screenshots=31,saved=result['saved'],
        appimage_automatic_resources=platform=='linux-x86_64',poster_center_pixels=691200,poster_center_different=0))
    diagdir=BATCH/'diagnostic-linux' if platform=='linux-x86_64' else ROOT/'workplace/music69-windows/full153/diagnostic'
    d=read(diagdir/'RESULT.json');assert d['passed'] and d['automatic_resources'] and len(d['runs'])==3
    assert d.get('appimage_sha256',d.get('binary_sha256'))==expected
    diagnostics.append(dict(platform=platform,**record(diagdir/'RESULT.json'),runs=3))
assert saves[0]==saves[1]==read(ROOT/'workplace/pointer152/baseline-linux-r2/RESULT.json')['saved']
for platform,assets in assets_by_platform.items():
    for n,v in inputs['common_files'].items():assert sha(assets/n)==v['sha256'],(platform,n)
standard=ROOT/'workplace/pointer152/standard-final.log';assert sha(standard)==read(ROOT/'docs/internals/data/pointer-toolbar152.json')['standards']['sha256']
generators=[ROOT/'tools/scripts'/n for n in ['full_appimage_release.py','full_appimage_receipt.py','appimage_build.sh','appimage_licenses.py','appimage_smoke.py','pointer_toolbar_gui.py']]+[ROOT/'tools/appimage.sh']
data=dict(version=VERSION,classification='含原版及衍生素材，本機完整版，不公開',source_files=len(inputs['source_files']),
    source_nonvendor=sum(not n.startswith('vendor/') for n in inputs['source_files']),common_files=len(inputs['common_files']),
    frozen_inputs=record(STAGE/'INPUTS.json'),pointer_build=record(OUT/'BUILD.json'),standards=dict(**record(standard),passed=16,no_tests=4),
    packages=packages,architectures=architectures,appimage_metadata=record(OUT/'full/APPIMAGE.json'),
    normal_player=normal,diagnostics=diagnostics,generators=[record(f) for f in generators],
    platform_correction='使用者明確表示152輪試玩為Linux；先前Windows推定不作測試平台事實。',
    limits=['AppImage採extract-and-run及Xvfb驗證，未驗桌面FUSE掛載。','Windows採Wine，macOS僅架構／依賴／ARM程式頁簽章，未作兩平台真機驗收。','不新增DOS parity、固定硬體延遲或全發行版相容聲明。','既有更新包、v1.6.0三平台與影片保留；未發布GitHub Release。'])
target=ROOT/'docs/internals/data/v161-full-delivery.json';buildfile=OUT/'full/BUILD.json'
fullsums=''.join(r['sha256']+'  '+Path(r['path']).name+'\n' for r in packages)
rootsums=old['archive']['sha256']+'  '+old['archive']['path']+'\n'+''.join(r['sha256']+'  '+r['path']+'\n' for r in packages)
if a.check:
    assert read(target)==data and read(buildfile)==data
    assert (OUT/'full/SHA256SUMS.txt').read_text()==fullsums and (OUT/'SHA256SUMS.txt').read_text()==rootsums
else:
    assert not target.exists() and not buildfile.exists()
    assert (OUT/'SHA256SUMS.txt').read_text()==old['archive']['sha256']+'  '+old['archive']['path']+'\n'
    write(target,data);write(buildfile,data)
    (OUT/'full/SHA256SUMS.txt').write_text(fullsums);(OUT/'SHA256SUMS.txt').write_text(rootsums)
print(json.dumps(dict(passed=True,version=VERSION,packages=packages,normal_player=normal),ensure_ascii=False,indent=2))
