"""166：獨立核對三平台完整包、實際GUI及替換海報後的完整推廣片。Docker內執行。"""
import argparse
import hashlib
import json
import plistlib
from pathlib import Path
import struct
import zipfile

from PIL import Image

import poster_frame_release as build
from local_test_macos_package import universal

ROOT,OUT,STAGE,BATCH,VERSION=build.ROOT,build.OUT,build.STAGE,build.BATCH,build.VERSION
sha,read,write=build.sha,build.read,build.write
def record(p):return dict(path=str(p.relative_to(ROOT)),bytes=p.stat().st_size,sha256=sha(p))

def main(check=False):
    build.configure();inputs=build.app.previous.source_check();packages=[];architectures={};assets={}
    for name,platform in [('WINDOWS','windows-amd64'),('MACOS','macos-universal')]:
        result=read(STAGE/(name+'.json'));item=result['package'];file=OUT/item['path'];folder=file.with_suffix('')
        assert file.stat().st_size==item['bytes'] and sha(file)==item['sha256']
        manifest=read(folder/'MANIFEST.json');assert manifest['version']==VERSION
        with zipfile.ZipFile(file)as z:
            assert z.testzip()is None
            assert set(z.namelist())=={folder.name+'/'+n for n in manifest['files']}|{folder.name+'/MANIFEST.json'}
            for n,row in manifest['files'].items():
                entry=z.getinfo(folder.name+'/'+n);data=z.read(entry)
                assert '..'not in Path(entry.filename).parts and not Path(entry.filename).is_absolute()
                assert entry.filename.isascii()or entry.flag_bits&0x800
                assert len(data)==row['bytes']and hashlib.sha256(data).hexdigest()==row['sha256']==sha(folder/n)
                assert (entry.external_attr>>16)&0o777==int(row['mode'],8)
            notes=z.read(folder.name+'/README.txt');assert notes.startswith(b'\xef\xbb\xbf')and VERSION in notes.decode('utf-8-sig')
            assert '60秒'in notes.decode('utf-8-sig')
        if name=='WINDOWS':
            assert b'\n'not in notes.replace(b'\r\n',b'')
            binary=folder/'gin3.exe';raw=binary.read_bytes();offset=struct.unpack_from('<I',raw,0x3c)[0]
            assert raw[:2]==b'MZ'and raw[offset:offset+4]==b'PE\0\0'and struct.unpack_from('<H',raw,offset+4)[0]==0x8664
            assets[platform]=folder;architectures[platform]='PE64 amd64，實際Wine另驗'
            for mode in ['original','enhanced']:
                command=(folder/('start-'+mode+'.cmd')).read_bytes();assert command.isascii()and b'\n'not in command.replace(b'\r\n',b'')
        else:
            contents=folder/'銀河英雄傳說III.app/Contents';binary=contents/'MacOS/gin3'
            info=plistlib.loads((contents/'Info.plist').read_bytes())
            assert info['LOGH3FullVersion']==VERSION and info['CFBundleVersion']==info['CFBundleShortVersionString']=='1.6.3'
            slices=universal(binary);assert slices==result['slices']
            assert all(s['deployment'][0]['sdk']=='15.5.0'for s in slices)
            assets[platform]=contents/'Resources';architectures[platform]=slices
        assert sha(binary)==result['binary_sha256']
        packages.append(dict(**item,binary_sha256=sha(binary),manifest_sha256=sha(folder/'MANIFEST.json')))
        print('PASS package',platform,flush=True)
    meta=read(OUT/'full/APPIMAGE.json');appfile=OUT/meta['package']['path']
    assert sha(appfile)==meta['package']['sha256']and appfile.stat().st_size==meta['package']['bytes']
    with appfile.open('rb')as f:prefix=f.read(meta['offset'])
    assert prefix[:4]==b'\x7fELF'and prefix[8:11]==b'AI\x02'and hashlib.sha256(prefix).hexdigest()==meta['runtime_sha256']
    appmanifest=read(BATCH/'APPDIR-MANIFEST.json');extracted=BATCH/'extracted/squashfs-root'
    assert sha(BATCH/'APPDIR-MANIFEST.json')==meta['appdir_manifest_sha256']
    assert build.app.package.inventory(extracted)==build.app.package.inventory(BATCH/'AppDir')==appmanifest['files']
    assets['linux-x86_64']=extracted/'usr/bin';binary=assets['linux-x86_64']/'gin3'
    raw=binary.read_bytes();assert raw[:6]==b'\x7fELF\x02\x01'and struct.unpack_from('<H',raw,18)[0]==62
    assert sha(binary)==read(STAGE/'LINUX.json')['binary_sha256']and len(meta['dependencies']['libraries'])==15
    architectures['linux-x86_64']='Type-2 AppImage、ELF64 x86_64、15項相依、extract-and-run'
    packages.append(dict(**meta['package'],binary_sha256=sha(binary),runtime_sha256=meta['runtime_sha256']))
    for platform,folder in assets.items():
        for n,row in inputs['common_files'].items():assert sha(folder/n)==row['sha256'],(platform,n)
        assert not any(p.name=='preferences.json'for p in folder.rglob('*'))
        q=read(folder/'art/poster/approved/quotes.json');assert q['interval_seconds']==60 and len(q['sets'])==4
        assert sha(folder/'art/poster/approved/QUOTE-SOURCES.json')==q['source_sha256']
    receipts=[];saves={}
    winroot=ROOT/'workplace/music69-windows/full166'
    for name,folder,file in [('linux-alliance',BATCH/'player-linux',appfile),('linux-empire',BATCH/'player-empire',appfile),
        ('windows-alliance',winroot/'player-windows-r2',winroot/('LOGH3-'+VERSION+'-windows-amd64')/'gin3.exe')]:
        result=read(folder/'RESULT.json');assert result['passed']and result['normal_entry']and result['binary_sha256']==sha(file)
        assert not any(result[k]for k in['state_injection','seed_injection','script_flag'])
        if name.startswith('linux'):
            assert result['appimage']and result['automatic_resources']
            assert not {'-orig','-data','-art','-save'}.intersection(result['command'])
        for e in result['events']:
            assert sha(folder/e['screenshot'])==e['sha256']
            with Image.open(folder/e['screenshot'])as image:assert image.size==tuple(e['window'])
        save=folder/('config/logh3/save'if name.startswith('linux')else'save')
        assert len(result['saved'])==6
        for n,h in result['saved'].items():assert sha(save/n)==h
        saves[name]=result['saved'];receipts.append(dict(name=name,**record(folder/'RESULT.json'),screenshots=len(result['events'])))
    assert saves['linux-alliance']==saves['windows-alliance']
    gfolder=BATCH/'garrison-linux-r3';g=read(gfolder/'RESULT.json');assert g['passed']and g['normal_entry']and g['appimage']
    assert g['binary_sha256']==sha(appfile)and not any(g[k]for k in['state_injection','seed_injection','script_flag'])
    captures={e['name']:e for e in g['events']if e.get('kind')=='capture'};assert len(captures)==19
    for event in captures.values():
        assert sha(gfolder/event['file'])==event['sha256']
        with Image.open(gfolder/event['file'])as image:assert image.size==(1280,720)
    before=captures['garrison-arrived']['observed']['strategic'];after=captures['fleet-ready']['observed']['strategic']
    assert before['stock']==[9000,3000,16000,2000,12000,3000,1000]
    assert after['stock']==[6000,3000,12000,2000,9000,3000,1000]and before['capital_stock']==after['capital_stock']
    fleet=g['loaded']['strategic']['fleets'][0];assert (fleet['Region'],fleet['Number'],fleet['Admiral'],fleet['Groups'][0]['kind'])==(18,3,41,2)
    assert len(g['saved'])==6
    for n,h in g['saved'].items():assert sha(gfolder/'config/logh3/save'/n)==h
    for folder in [BATCH/'diagnostic-linux',winroot/'diagnostic-windows']:
        result=read(folder/'RESULT.json');assert result['passed']and result['version']==VERSION and len(result['runs'])==3
        receipts.append(record(folder/'RESULT.json'))
    lines=(BATCH/'standard-final.log').read_text().splitlines()
    assert sum(s.startswith('ok ')for s in lines)==16 and sum('[no test files]'in s for s in lines)==4
    assert not any(s.startswith('FAIL')for s in lines)
    promo=OUT/'promo';audit=read(promo/'AUDIT.json');plan=read(promo/'PLAN.json')
    assert audit['passed']and audit['visual_review'].startswith('已目視')and not audit['black_segments']and not audit['subtitle_clipping']
    assert audit['audio']['snr_db']>25 and audit['retained_scene_count']==18 and audit['new_poster_scene_count']==14
    assert sha(promo/audit['file'])==audit['sha256']and sha(promo/'PLAN.json')==audit['plan_sha256']
    for n,h in plan['input_sha256'].items():assert sha(ROOT/n)==h
    listing=[OUT/r['path']for r in packages]+[p for p in sorted(promo.iterdir())if p.is_file()]
    sums=''.join(sha(p)+'  '+str(p.relative_to(OUT))+'\n'for p in listing)
    value=dict(passed=True,version=VERSION,classification='含原版及衍生音畫，本機完整版與影片，不公開',
        packages=packages,architectures=architectures,source_files=len(inputs['source_files']),
        source_nonvendor=sum(not n.startswith('vendor/')for n in inputs['source_files']),common_files=len(inputs['common_files']),
        inputs=record(STAGE/'INPUTS.json'),normal=receipts,garrison=dict(**record(gfolder/'RESULT.json'),screenshots=19,saved=g['saved']),
        standards=dict(passed=16,no_tests=4,log=record(BATCH/'standard-final.log')),
        movie=dict(**record(promo/audit['file']),seconds=audit['seconds'],audit=record(promo/'AUDIT.json'),plan=record(promo/'PLAN.json'),
            retained_scenes=18,replaced_poster_scenes=14,story=plan['storyboard'],scope=plan['scope']),
        checksums=dict(path=str((OUT/'SHA256SUMS.txt').relative_to(ROOT)),sha256=hashlib.sha256(sums.encode()).hexdigest()),
        limits=['Linux採extract-and-run，未驗桌面FUSE','Windows使用Wine，macOS只核雙架構／依賴／ARM簽章，未真機','未新增DOS parity或尖端官方逐字聲明，未公開完整版／影片'])
    target=ROOT/'docs/internals/data/v163-full-delivery.json';full=OUT/'full/BUILD.json'
    if check:
        assert read(target)==read(full)==value and (OUT/'SHA256SUMS.txt').read_text()==sums
    else:
        assert not target.exists()and not full.exists()
        write(target,value);write(full,value);(OUT/'SHA256SUMS.txt').write_text(sums)
        (OUT/'full/SHA256SUMS.txt').write_text(''.join(r['sha256']+'  '+Path(r['path']).name+'\n'for r in packages))
    print('PASS full delivery',VERSION,[(Path(r['path']).name,r['bytes'])for r in packages],flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--check',action='store_true');a=p.parse_args();main(a.check)
