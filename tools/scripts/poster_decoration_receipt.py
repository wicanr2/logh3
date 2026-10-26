"""154：獨立驗證新裝飾、正式三平台封包與正常操作收據。"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import plistlib
import statistics
import struct
import subprocess
import zipfile

import numpy as np
from PIL import Image
import poster_decoration_release as build
from local_test_macos_package import universal

def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--check',action='store_true');args=parser.parse_args()
    ROOT,OUT,STAGE,BATCH,VERSION=build.ROOT,build.OUT,build.STAGE,build.BATCH,build.VERSION
    read,write,sha=build.read,build.write,build.sha
    record=lambda f:dict(path=str(f.relative_to(ROOT)),bytes=f.stat().st_size,sha256=sha(f))
    build.configure();inputs=build.app.previous.source_check()
    assert len(inputs['source_files'])==1620 and len(inputs['common_files'])==618
    packages=[];architectures={};asset_sets={}
    for name,platform in [('WINDOWS','windows-amd64'),('MACOS','macos-universal')]:
        result=read(STAGE/(name+'.json'));row=result['package'];archive=OUT/row['path'];folder=archive.with_suffix('')
        assert sha(archive)==row['sha256'] and archive.stat().st_size==row['bytes']
        manifest=read(folder/'MANIFEST.json');assert manifest['version']==VERSION
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            assert set(z.namelist())=={folder.name+'/'+n for n in [*manifest['files'],'MANIFEST.json']}
            for n,v in manifest['files'].items():
                entry=z.getinfo(folder.name+'/'+n);content=z.read(entry)
                assert not Path(entry.filename).is_absolute() and '..' not in Path(entry.filename).parts
                assert entry.filename.isascii() or entry.flag_bits&0x800
                assert len(content)==v['bytes'] and hashlib.sha256(content).hexdigest()==v['sha256']==sha(folder/n)
                assert (entry.external_attr>>16)&0o777==int(v['mode'],8)
        assert read(folder/'VERSION.json')['version']==VERSION
        notes=(folder/'README.txt').read_bytes();assert notes.startswith(b'\xef\xbb\xbf') and '每12秒' in notes.decode('utf-8-sig')
        assert not any(f.name=='preferences.json' for f in folder.rglob('*'))
        if name=='WINDOWS':
            binary=folder/'gin3.exe';assets=folder;raw=binary.read_bytes();offset=struct.unpack_from('<I',raw,0x3c)[0]
            assert raw[:2]==b'MZ' and raw[offset:offset+4]==b'PE\0\0' and struct.unpack_from('<H',raw,offset+4)[0]==0x8664
            installed=ROOT/'workplace/music69-windows/full154'/folder.name
            for n,v in manifest['files'].items():assert sha(installed/n)==v['sha256'],n
            assert b'\n' not in notes.replace(b'\r\n',b'')
            architectures[platform]='PE64 amd64；實際ZIP解出後Wine正常操作'
        else:
            contents=folder/'銀河英雄傳說III.app/Contents';assets=contents/'Resources';binary=contents/'MacOS/gin3'
            info=plistlib.loads((contents/'Info.plist').read_bytes())
            assert info['LOGH3FullVersion']==VERSION and info['CFBundleVersion']==info['CFBundleShortVersionString']=='1.6.2'
            slices=universal(binary);assert slices==result['slices'] and {s['architecture'] for s in slices}=={'amd64','arm64'}
            assert all(s['deployment']==[{'macos':'11.0.0','sdk':'15.5.0'}] for s in slices)
            assert next(s for s in slices if s['architecture']=='arm64')['code_directory'][0]['all_page_hashes_match']
            for f in [contents/'MacOS/launch',binary,folder/'start-original.command',folder/'start-enhanced.command']:assert os.access(f,os.X_OK)
            architectures[platform]=slices
        assert sha(binary)==result['binary_sha256'];asset_sets[platform]=assets
        packages.append(dict(**row,binary_sha256=sha(binary),manifest_sha256=sha(folder/'MANIFEST.json')))
    app=read(OUT/'full/APPIMAGE.json');appfile=OUT/app['package']['path']
    assert sha(appfile)==app['package']['sha256'] and appfile.stat().st_size==app['package']['bytes'] and os.access(appfile,os.X_OK)
    with appfile.open('rb') as f:prefix=f.read(app['offset'])
    assert prefix[:4]==b'\x7fELF' and prefix[8:11]==b'AI\x02' and hashlib.sha256(prefix).hexdigest()==app['runtime_sha256']
    appmanifest=read(BATCH/'APPDIR-MANIFEST.json');extracted=BATCH/'extracted/squashfs-root'
    assert sha(BATCH/'APPDIR-MANIFEST.json')==app['appdir_manifest_sha256']
    assert build.app.package.inventory(extracted)==build.app.package.inventory(BATCH/'AppDir')==appmanifest['files']
    assets=extracted/'usr/bin';binary=assets/'gin3';asset_sets['linux-x86_64']=assets
    assert sha(binary)==read(STAGE/'LINUX.json')['binary_sha256']
    raw=binary.read_bytes();assert raw[:6]==b'\x7fELF\x02\x01' and struct.unpack_from('<H',raw,18)[0]==62
    assert len(app['dependencies']['libraries'])==15
    assert max(tuple(map(int,v.split('.'))) for v in app['dependencies']['glibc_requirements'].values() if v)==(2,34)
    architectures['linux-x86_64']='Type-2 AppImage；ELF64 x86_64；最大GLIBC_2.34；15項相依；extract-and-run正常操作'
    packages.append(dict(**app['package'],binary_sha256=sha(binary),appdir_files=len(appmanifest['files']),runtime_sha256=app['runtime_sha256']))
    for platform,assets in asset_sets.items():
        for n,v in inputs['common_files'].items():assert sha(assets/n)==v['sha256'],(platform,n)
        assert not any(f.name=='preferences.json' for f in assets.rglob('*'))
    normal=[];saved=[]
    locations=[('linux-x86_64',BATCH/'player-linux',appfile),
               ('windows-amd64',ROOT/'workplace/music69-windows/full154/player-windows-r3',ROOT/'workplace/music69-windows/full154'/('LOGH3-'+VERSION+'-windows-amd64')/'gin3.exe'),
               ('linux-empire',BATCH/'player-empire-r2',appfile)]
    for platform,directory,binary in locations:
        result=read(directory/'RESULT.json');assert result['passed'] and result['normal_entry']
        assert not any(result[k] for k in ['state_injection','seed_injection','script_flag'])
        driver=BATCH/'poster_decoration_gui-r1.py' if platform=='linux-x86_64' else BATCH/'poster_decoration_gui-r2.py' if platform=='linux-empire' else ROOT/'tools/scripts/poster_decoration_gui.py'
        assert result['driver_sha256']==sha(driver) and result['binary_sha256']==sha(binary)
        assert len(result['events'])==23 and len(result['saved'])==6
        save_dir=directory/('save' if platform=='windows-amd64' else 'config/logh3/save')
        for n,h in result['saved'].items():assert sha(save_dir/n)==h
        if platform!='linux-empire':saved.append(result['saved'])
        if platform.startswith('linux'):
            assert result['appimage'] and result['automatic_resources'] and not {'-orig','-data','-art','-save'}.intersection(result['command'])
        events={v['name']:v for v in result['events']}
        for e in result['events']:
            f=directory/e['screenshot'];assert sha(f)==e['sha256']
            with Image.open(f) as im:assert im.size==tuple(e['window']);im.verify()
        for code in ['zh-TW','zh-CN','ja','en','ko']:assert events['settings-'+code]['observed']['locale']==code
        loop=result['loop'];assert loop['uploads_before']==loop['uploads_after']
        phases={v['decoration']['phase'] for v in loop['samples']};pages={v['decoration']['page'] for v in loop['samples']}
        assert len(phases)>20 and len(pages)>=2
        video=loop.get('video',True)
        if video:
            assert loop['sha256']==sha(directory/'loop.mp4')
            ff=json.loads(subprocess.check_output(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(directory/'loop.mp4')],text=True))
            assert len(ff['streams'])==1 and ff['streams'][0]['codec_type']=='video' and float(ff['format']['duration'])>=17.9
        pixel_checks=[]
        for off,on in [('map-off','decor-v1'),('original-off','original-restored')]:
            plain=np.asarray(Image.open(directory/(off+'.png')).convert('RGB'))
            full=np.asarray(Image.open(directory/(on+'.png')).convert('RGB'));h,w=plain.shape[:2];offset=(full.shape[1]-w)//2
            center=full[:,offset:offset+w];delta=np.any(plain!=center,axis=2)
            allowed=np.zeros(delta.shape,dtype=bool);bands=events[on]['observed']['poster_decoration']['bands']
            for x,y,x2,y2 in bands:allowed[math.floor(y*h/480):math.ceil(y2*h/480),math.floor(x*w/640):math.ceil(x2*w/640)]=True
            outside=int(np.count_nonzero(delta&~allowed));assert outside==0,(platform,off,on,outside)
            assert np.count_nonzero(delta)>0
            pixel_checks.append(dict(off=off,on=on,bands=bands,compared=int(np.count_nonzero(~allowed)),outside_different=outside))
        still=np.asarray(Image.open(directory/'animation-off.png').convert('RGB'))
        later=np.asarray(Image.open(directory/'animation-off-later.png').convert('RGB'));assert np.array_equal(still,later)
        original=np.asarray(Image.open(directory/'decor-v1.png').convert('RGB'))
        side_pixels=0
        for name in ['decor-after-loop','original-restored','decor-v1-cycled','decor-final']:
            current=np.asarray(Image.open(directory/(name+'.png')).convert('RGB'))
            for left,right in [(0,160),(1120,1280)]:
                assert np.array_equal(original[:,left:right],current[:,left:right]),(platform,name,'海報或名言改變')
                side_pixels+=original.shape[0]*(right-left)
        assert events['main-poster-on']['observed']['poster_decoration']['bands']==[[0,0,0,0],[0,0,0,0]]
        normal.append(dict(platform=platform,**record(directory/'RESULT.json'),screenshots=23,pixel_checks=pixel_checks,
            driver=record(driver),loop=record(directory/'loop.mp4') if video else None,pages=sorted(pages),phases=len(phases),background_uploads=[loop['uploads_before'],loop['uploads_after']],
            draw_median_ms=statistics.median(v['draw_ms'] for v in loop['samples']),static_pixels_identical=True,
            side_pixels=side_pixels,side_different=0,saved=result['saved']))
    baseline=read(ROOT/'workplace/pointer152/baseline-linux-r2/RESULT.json')['saved'];assert saved[0]==saved[1]==baseline
    standard=BATCH/'standard-final.log';lines=standard.read_text().splitlines()
    assert sum(l.startswith('ok ') for l in lines)==16 and sum('[no test files]' in l for l in lines)==4
    assert not any(l.startswith('FAIL') for l in lines)
    promo=OUT/'promo';preview=promo/('LOGH3-'+VERSION+'-poster-loop.mp4')
    preview_meta=read(promo/'PREVIEW.json');assert preview_meta['sha256']==sha(preview)
    ff=read(promo/'FFPROBE.json');assert len(ff['streams'])==1
    stream=ff['streams'][0]
    assert (stream['codec_name'],stream['width'],stream['height'],stream['r_frame_rate'],int(stream['nb_frames']))==('h264',1280,720,'15/1',270)
    assert float(ff['format']['duration'])==18
    media=(promo/'MEDIA-CHECK.txt').read_text()
    assert 'black_start:' not in media and 'freeze_start:' not in media
    data=dict(version=VERSION,classification='含原版與衍生素材，本機完整版，不公開',source_files=len(inputs['source_files']),
        source_nonvendor=sum(not n.startswith('vendor/') for n in inputs['source_files']),common_files=len(inputs['common_files']),
        frozen_inputs=record(STAGE/'INPUTS.json'),standards=dict(**record(standard),passed=16,no_tests=4),packages=packages,
        architectures=architectures,normal_player=normal,appimage_metadata=record(OUT/'full/APPIMAGE.json'),
        fighters=record(ROOT/'assets/hd/poster/fighters.png'),flagships=[16,12],
        preview=dict(**record(preview),ffprobe=record(promo/'FFPROBE.json'),metadata=record(promo/'PREVIEW.json'),
            media_check=record(promo/'MEDIA-CHECK.txt'),duration=18,frames=270,audio=False,black_segments_05s=0,freezes_2s=0),
        generators=[record(ROOT/'tools/scripts'/n) for n in ['poster_decoration_release.py','poster_decoration_receipt.py','poster_decoration_preview.py','poster_decoration_gui.py','appimage_build.sh']],
        limits=['動畫與CPU耗時採Xvfb軟體繪圖抽樣，未證明實機FPS或硬體輸入延遲。','Linux採AppImage extract-and-run，未驗桌面FUSE。','Windows採Wine，macOS雙架構／依賴／ARM簽章通過但未真機。','新增remake裝飾，不作DOS畫面parity聲明；v1.6.0推廣片保留，循環短片只是實錄預覽。'])
    target=ROOT/'docs/internals/data/poster-decoration154.json';buildfile=OUT/'full/BUILD.json'
    sums=''.join(r['sha256']+'  '+Path(r['path']).name+'\n' for r in packages)
    rootsums=''.join(r['sha256']+'  '+r['path']+'\n' for r in packages)+sha(preview)+'  '+str(preview.relative_to(OUT))+'\n'
    if args.check:
        assert read(target)==read(buildfile)==data
        assert (OUT/'full/SHA256SUMS.txt').read_text()==sums and (OUT/'SHA256SUMS.txt').read_text()==rootsums
    else:
        assert not target.exists() and not buildfile.exists()
        write(target,data);write(buildfile,data);(OUT/'full/SHA256SUMS.txt').write_text(sums);(OUT/'SHA256SUMS.txt').write_text(rootsums)
    print(json.dumps(dict(passed=True,packages=packages,normal_player=normal),ensure_ascii=False,indent=2))

if __name__=='__main__':main()
