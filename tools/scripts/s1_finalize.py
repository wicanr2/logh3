"""核對本輪凍結來源、三平台封包、S1實錄及HD試樣，集中本機交付。"""
import hashlib
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import zipfile
from local_test_macos_package import universal

ROOT=Path('/src')
VERSION='v.1.0.1-20261011'
OUT=ROOT/'dist-all'/VERSION
BATCH=ROOT/'workplace/hd123-s1-20261011'

def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as stream:
        for block in iter(lambda:stream.read(1024*1024),b''):h.update(block)
    return h.hexdigest()

def record(p,relative=ROOT):
    return dict(path=str(p.relative_to(relative)),bytes=p.stat().st_size,sha256=sha(p))

def write(p,v):
    p.write_text(json.dumps(v,ensure_ascii=False,indent=2)+'\n')

def copy(source,target):
    target.parent.mkdir(parents=True,exist_ok=True)
    assert target.parent.stat().st_uid==1000 and source.is_file()
    if target.exists():assert sha(source)==sha(target),str(target)
    else:shutil.copyfile(source,target)

def main(stage_promo=False):
    assert os.getuid()==1000 and OUT.stat().st_uid==1000
    build=json.loads((OUT/'BUILD.json').read_text())
    stage=ROOT/f'workplace/package92-{VERSION}-s1/source'
    for name,expected in build['source_files'].items():
        frozen=stage/name
        assert sha(frozen)==expected['sha256'] and frozen.stat().st_size==expected['bytes'],name
        if not name.startswith('vendor/'):
            assert sha(ROOT/name)==expected['sha256'],name+'正式來源已改變'
    assert len(build['packages'])==8
    manifests=[]
    for p in build['packages']:
        archive=OUT/p['path']
        assert record(archive,OUT)==p
        if archive.suffix!='.zip':continue
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            name=archive.stem
            manifest=json.loads(z.read(name+'/MANIFEST.json'))
            assert manifest['version']==VERSION
            assert set(z.namelist())=={name+'/MANIFEST.json',*(name+'/'+f for f in manifest['files'])}
            for f,expected in manifest['files'].items():
                data=z.read(name+'/'+f)
                assert len(data)==expected['bytes'] and hashlib.sha256(data).hexdigest()==expected['sha256'],f
            if p['path'].startswith('release/'):
                forbidden={'.amg','.bmg','.img','.mds','.dat','.bit','.ogg','.sf2','.com'}
                assert not any('/original/' in f or '/art/' in f or Path(f).suffix.lower() in forbidden for f in z.namelist())
            manifests.append(dict(package=p['path'],files=len(manifest['files']),valid=True))
    for platform,binary in [('linux-amd64','gin3'),('windows-amd64','gin3.exe'),('macos-universal','銀河英雄傳說III.app/Contents/MacOS/gin3')]:
        assert sha(OUT/'full'/f'LOGH3-{VERSION}-{platform}'/binary)==sha(OUT/'release'/f'LOGH3-{VERSION}-{platform}'/binary)
    macos=universal(OUT/'full'/f'LOGH3-{VERSION}-macos-universal/銀河英雄傳說III.app/Contents/MacOS/gin3')
    for src,label in [(ROOT/f'workplace/package92-smoke-{VERSION}','linux-full'),
                      (ROOT/f'workplace/music69-windows/package92-smoke-{VERSION}','windows-full'),
                      (ROOT/f'workplace/music69-windows/release-smoke-{VERSION}','windows-release')]:
        assert (src/'SMOKE.json').is_file()
        for p in sorted(src.iterdir()):
            if p.is_file() and p.suffix in ('.json','.png','.log','.txt','.bin'):copy(p,OUT/'smoke'/label/p.name)
    assert (OUT/'smoke/release-linux/SMOKE.json').is_file()
    for src,label in [(BATCH/'scenarios-linux','scenarios-linux'),(ROOT/f'workplace/music69-windows/scenarios-{VERSION}','scenarios-windows')]:
        receipt=json.loads((src/'SCENARIOS.json').read_text())
        assert [r['loaded_turn'] for r in receipt['runs']]==[346,93,346,380]
        assert all(r['reloaded_s3_turn']==380 for r in receipt['runs'])
        for p in sorted(src.rglob('*')):
            if p.is_file():copy(p,OUT/'smoke'/label/p.relative_to(src))
    log=BATCH/'standard-tests.log'
    lines=log.read_text().splitlines()
    assert sum(l.startswith('ok ') for l in lines)==14 and sum(l.startswith('? ') for l in lines)==5
    copy(log,OUT/'smoke/standard-tests.log')
    source=BATCH/'promo'
    audit=json.loads((source/'AUDIT.json').read_text())
    movie=source/f'LOGH3-{VERSION}-promo.mp4'
    assert audit['video_sha256']==sha(movie) and audit['frames']==2700 and audit['seconds']==90
    assert audit['visual_review']['passed'] and audit['non_silent'] and audit['black_segments']==0
    for p in sorted(source.iterdir()):
        if p.is_file() and (p.name in ('PLAN.json','AUDIT.json','FFPROBE.json','RIGHTS.txt','AUDIO-AUDIT.txt','VIDEO-AUDIT.txt','TEXT-BOUNDS.json','CONTACT.png','COMBAT-CONTACT.png','COMBAT-FRAMES.md5',movie.name) or p.name.startswith(('sample-','combat-'))):
            copy(p,OUT/'promo'/p.name)
    traces=[]
    for name,label in [('record-f','load'),('record-d','battle')]:
        raw=BATCH/name
        trace=json.loads((raw/'TRACE.json').read_text())
        assert trace['normal_x11_input'] and not trace['script_flag'] and not trace['state_injection']
        assert sha(raw/'RAW.mp4')==trace['raw_sha256']
        copy(raw/'RAW.mp4',OUT/'promo'/('RAW-S1-'+label+'.mp4'))
        for f in ('TRACE.json','EVENTS.json'):copy(raw/f,OUT/'promo'/('S1-'+label+'-'+f))
        traces.append(dict(record=name,trace=record(raw/'TRACE.json'),raw=record(raw/'RAW.mp4')))
    write(OUT/'smoke/PACKAGE-CHECK.json',dict(version=VERSION,source_files=len(build['source_files']),packages=build['packages'],manifests=manifests,macos=macos,promo=audit,traces=traces))
    if stage_promo:
        print('PASS 凍結來源／八封包／六ZIP逐檔／三劇本／S1影片及實錄已集中',flush=True)
        return
    hd=BATCH/'hd'
    browser=json.loads((hd/'BROWSER.json').read_text())
    assert all(c['pass'] for c in browser['checks']) and browser['visual_review']['passed']
    copy(hd/'HD-INTERFACE.html',OUT/'HD-INTERFACE.html')
    for p in hd.iterdir():
        if p.is_file() and p.suffix in ('.json','.png','.log'):copy(p,OUT/'smoke/hd-preview'/p.name)
    copy(BATCH/'FLAG-SOURCES.json',OUT/'promo/FLAG-SOURCES.json')
    plan=ROOT/'docs/goal/2026-10-11-hd-interface-and-s1-promo-round-123.md'
    def rebase(m):
        target=m.group(1)
        if '://' in target or target.startswith('#'):return m.group(0)
        resolved=(plan.parent/target).resolve();assert resolved.is_file(),str(resolved)
        return ']('+os.path.relpath(resolved,OUT)+')'
    (OUT/'HD-PLAN.md').write_text(re.sub(r'\]\(([^)]+)\)',rebase,plan.read_text()))
    build['classification']=VERSION+'正式本機完整版，含原版及衍生素材；禁止公開散布'
    build['frozen_source']=str(stage.relative_to(ROOT))
    write(OUT/'BUILD.json',build)
    scripts=['s1_finalize.py','s1_gameplay_record.py','s1_story_promo.py','bundled_scenarios_smoke.py',
             'hd_interface_inventory.py','hd_interface_preview.html','hd_interface_browser_check.py','hd_preview_deps.py',
             'local_test_package.py','local_test_macos_package.py','local_test_macos_build.sh','local_test_smoke.py',
             'local_test_windows_smoke.py','v1_release.py','v1_release_smoke.py']
    sources=[record(ROOT/'tools/scripts'/f) for f in scripts]
    sources += [record(ROOT/p) for p in ('tools/Dockerfile.hd-preview','tools/s1_record_display.sh','docs/spec/129-bundled-scenario-catalog.md',
                                       'docs/playtest/110-hd-interface-s1-promo.md',str(plan.relative_to(ROOT)))]
    sources += [record(stage/p) for p in build['source_files']]
    evidence=[record(p) for folder in (OUT/'smoke',OUT/'promo') for p in sorted(folder.rglob('*')) if p.is_file()]
    summary=dict(version=VERSION,source_head=build['head'],dirty_source_snapshot=True,frozen_source=build['frozen_source'],
                 source=sources,packages=build['packages'],manifests=manifests,macos=macos,traces=traces,promo=audit,
                 tests=dict(packages=19,passed=14,without_tests=5),hd=dict(inventory=record(hd/'INVENTORY.json'),
                 png_files=2611,interface='可丟棄試樣，待選型',animation='已納HD計畫；正式逐幀重製未完成',browser=browser),
                 evidence=evidence,rights=dict(full='本機限定',promo='本機限定',preview='本機限定',release='RRSAL-1.0引擎包，不含原版及HD素材'),
                 not_claimed=build['not_claimed'],published=False)
    write(OUT/'FINAL-VERIFICATION.json',summary)
    write(ROOT/'docs/internals/data/s1-gameplay-promo.json',summary)
    hashes=build['packages']+[record(p,OUT) for p in sorted((OUT/'promo').iterdir()) if p.is_file() and p.suffix=='.mp4']
    hashes += [record(OUT/f,OUT) for f in ('BUILD.json','FINAL-VERIFICATION.json','HD-PLAN.md','HD-INTERFACE.html')]
    write(OUT/'SHA256SUMS.json',dict(version=VERSION,files=hashes))
    print('PASS 1503凍結來源／六ZIP逐檔／八封包／三劇本／90秒實錄影片／HD操作試樣',flush=True)

if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--stage-promo',action='store_true')
    main(parser.parse_args().stage_promo)
