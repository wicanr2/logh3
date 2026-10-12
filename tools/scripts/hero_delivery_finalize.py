"""規格140：核對八封包、正常名將實錄與新片，集中本機交付並保存來源身份。"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import tarfile
import zipfile

from toolbar_delivery_finalize import ROOT, read, record, sha, write

version='v.1.2.0-20261012'
out=ROOT/'dist-all'/version
batch=ROOT/'workplace/hero130'
assert os.getuid()==1000 and out.stat().st_uid==1000
build=read(out/'BUILD.json')
assert build['version']==version and len(build['packages'])==8
frozen=ROOT/'workplace'/('package92-'+version+'-final')/'source'
for name,item in build['source_files'].items():
    assert sha(frozen/name)==item['sha256'],name
    if not name.startswith('vendor/'):
        assert sha(ROOT/name)==item['sha256'],('建置後產品來源改變',name)

manifests=[]
for package in build['packages']:
    path=out/package['path']
    assert sha(path)==package['sha256'] and path.stat().st_size==package['bytes']
    if path.suffix=='.zip':
        with zipfile.ZipFile(path) as archive:
            assert archive.testzip() is None
            manifest=read(path.parent/path.stem/'MANIFEST.json')
            assert manifest['version']==version
            assert set(archive.namelist())=={path.stem+'/'+n for n in manifest['files']}|{path.stem+'/MANIFEST.json'}
            for name,item in manifest['files'].items():
                assert not Path(name).is_absolute() and '..' not in Path(name).parts
                data=archive.read(path.stem+'/'+name)
                assert len(data)==item['bytes'] and hashlib.sha256(data).hexdigest()==item['sha256']
                assert sha(path.parent/path.stem/name)==item['sha256']
                if path.parent.name=='release':
                    assert not {'original','art','music','soundfont'}.intersection(Path(name).parts)
                if path.parent.name=='full' and name=='README.txt':
                    text=data.decode('utf-8-sig')
                    assert all(t in text for t in ('1.5','楊威利','萊因哈特','F1','F7','設定選單'))
                    assert 'F8' not in text
                if 'windows' not in path.name and (name.endswith(('.sh','.command','/MacOS/gin3','/MacOS/launch')) or name=='gin3'):
                    assert archive.getinfo(path.stem+'/'+name).external_attr>>16 & 0o111
            manifests.append({'path':package['path'],'files':len(manifest['files']),'passed':True})
    else:
        with tarfile.open(path,'r:gz') as archive:
            files=[m for m in archive.getmembers() if m.isfile()]
            assert len(files)==len(build['source_files'])
            for member in files:
                name=str(Path(member.name).relative_to('logh3-source'))
                assert hashlib.sha256(archive.extractfile(member).read()).hexdigest()==build['source_files'][name]['sha256']

lines=(batch/'TESTS-FINAL.log').read_text().splitlines()
assert not any(line.startswith(('FAIL','--- FAIL')) for line in lines)
tests=dict(passed=sum(line.startswith('ok ') for line in lines),no_tests=sum(line.startswith('? ') for line in lines))
assert tests==dict(passed=15,no_tests=4)
gameplay=read(batch/'GAMEPLAY-AUDIT.json')
assert gameplay['passed'] and len(gameplay['recordings'])==4
playback=[]
for name in ['full-record-f0','full-record-f1-r2','record-f1-ai-1']:
    report=read(batch/name/'PLAYBACK-AUDIT.json')
    assert report['passed'] and not report['skipped_frames']
    assert sha(batch/name/'FRAMES.jsonl')==report['frame_trace_sha256']
    playback.append(report)
assert {s['weapon'] for p in playback for s in p['scenes'] if s['kind']=='fleet'}=={0,1,2}

platforms=[
    (ROOT/'workplace'/('package92-smoke-'+version),out/'smoke/full-linux',11),
    (ROOT/'workplace/music69-windows'/('package92-smoke-'+version),out/'smoke/full-windows-wine',6),
    (ROOT/'workplace/music69-windows'/('release-smoke-'+version),out/'smoke/release-windows-wine',4),
]
for source,target,count in platforms:
    assert read(source/'SMOKE.json')['version']==version
    assert len(read(source/'SMOKE.json')['runs'])==count
    shutil.copytree(source,target,dirs_exist_ok=True,ignore=shutil.ignore_patterns('LOGH3-*','save-*'))
assert len(read(out/'smoke/release-linux/SMOKE.json')['runs'])==4
mac=read(out/'smoke/macos/MACOS.json')
assert {s['architecture'] for s in mac['slices']}=={'amd64','arm64'}
assert all(c['all_page_hashes_match'] for s in mac['slices'] for c in s['code_directory'])
shutil.copyfile(batch/'TESTS-FINAL.log',out/'smoke/standard-tests.log')
shutil.copyfile(batch/'GAMEPLAY-AUDIT.json',out/'smoke/GAMEPLAY-AUDIT.json')

promo=batch/'promo'
movie=promo/'LOGH3-HERO-PROMO.mp4'
audit=read(promo/'AUDIT.json')
assert audit['video_sha256']==sha(movie) and audit['non_silent'] and audit['black_segments']==0
assert audit['battle_animation_seconds']>=50 and audit['anime_seconds']==16
assert audit['visual_review']=='PASS 完整視窗、片頭、字幕、來源片段及正常戰鬥抽樣'
assert read(promo/'SOURCE-FRAMES.json')['passed']
assert read(promo/'FREEZE-REVIEW.json')['passed']
plan=read(promo/'PLAN.json')
assert all(sha(ROOT/'workplace'/Path(name).relative_to('/work') if name.startswith('/work/') else Path(name))==digest
           for name,digest in plan['inputs'].items())
for name in ['PLAN.json','AUDIT.json','TEXT-BOUNDS.json','SOURCE-FRAMES.json','FREEZE-REVIEW.json',
             'AUDIO-AUDIT.log','VIDEO-AUDIT.log','CONTACT.jpg']:
    shutil.copyfile(promo/name,out/'promo'/name)
destination=out/'promo'/('LOGH3-'+version+'-promo.mp4')
shutil.copyfile(movie,destination)
shutil.copyfile(batch/'ANIME-SOURCE-CLEAN.json',out/'promo/ANIME-SOURCE.json')
shutil.copyfile(batch/'anime-sources/SOURCES.json',out/'promo/POSTER-SOURCES.json')
for name in ['full-record-f0','full-record-f1-r2','record-f0-ai-1','record-f1-ai-1','record-f0-ai1']:
    source=batch/name
    shutil.copyfile(source/'RAW.mp4',out/'promo'/(name+'.mp4'))
    shutil.copytree(source,out/'smoke'/name,dirs_exist_ok=True,
                    ignore=shutil.ignore_patterns('RAW.mp4','pulse','save','player-save'))
(out/'promo/README.md').write_text(
    f'# {version} 推廣片\n\n成片：`{destination.name}`，1080p60、雙聲道，約125秒。\n\n'
    '雙旗HD片頭、官方1988／1993修復海報、使用者指定1988第1話16秒、正常S1讀檔與三個額外劇本、'
    '兩位名將與既有原版規則模式比較、F7原版／HD即時切換、52秒完整艦隊交火。\n\n'
    '指定動畫原生852×480，依原比例放大，保留來源水印與字幕；不稱原生HD。遊戲動畫保留普通速度與原聲。'
    '全部影片與完整遊戲含原版或衍生素材，僅本機使用。來源時碼、雜湊與聲畫核對見本目錄JSON；正常原始實錄另附。\n')

counts={'scenes':sum(len(p['scenes']) for p in playback),
        'pictures':sum(s['pictures'] for p in playback for s in p['scenes'])}
check=dict(version=version,passed=True,source_files=len(build['source_files']),tests=tests,
           manifests=manifests,normal_gameplay=gameplay,playback=counts,
           linux_full_runs=11,windows_full_wine_runs=6,linux_engine_runs=4,windows_engine_wine_runs=4,
           macos_architectures=['amd64','arm64'],promo=audit,
           limits=['Windows以Wine驗收，macOS僅架構、依賴及ARM簽章驗證，未真機',
                   '音訊技術核對不等於人耳；完整登陸及DOS同態保持原開放項'])
write(out/'smoke/PACKAGE-CHECK.json',check)
sources=[record(ROOT/n) for n in build['source_files'] if not n.startswith('vendor/')]
sources += [record(ROOT/'tools/scripts'/n) for n in ['hero_gameplay_model.go','hero_gameplay_audit.py',
    'hero_promo_plan.py','hero_story_promo.py','hero_promo_review.py','hero_anime_download.py','hero_promo_sources.py',
    'hero_delivery_finalize.py','battle_playback_audit.py','ai_gameplay_record.py',
    'local_test_package.py','local_test_macos_package.py','local_test_macos_build.sh','v1_release.py']]
evidence=[record(batch/n) for n in ['GAMEPLAY-AUDIT.json','TESTS-FINAL.log','ANIME-SOURCE-CLEAN.json','ANIME-CLEAN-DECODE.log']]
evidence += [record(batch/n/'PLAYBACK-AUDIT.json') for n in ['full-record-f0','full-record-f1-r2','record-f1-ai-1']]
evidence += [record(p) for p in (out/'promo').iterdir() if p.is_file()]
evidence += [record(out/'smoke/PACKAGE-CHECK.json')]
write(ROOT/'docs/internals/data/hero-ai-release-promo.json',dict(
    version=version,status='PASS 名將戰術、三平台本機完整版及有聲史詩推廣片',
    packages=build['packages'],checks=check,source_head=build['head'],dirty_source_snapshot=True,
    frozen_build_source=str(frozen.relative_to(ROOT)),sources=sources,evidence=evidence,
    rights='full與promo僅本機；release僅引擎與來源，不含原版素材',
    remote_release='尚未建立，需依主機gh例外範圍處理'))
sha_entries=build['packages']+[dict(path=str(p.relative_to(out)),bytes=p.stat().st_size,sha256=sha(p))
    for p in (out/'promo').iterdir() if p.is_file()]
write(out/'SHA256SUMS.json',dict(version=version,files=sha_entries))
print('PASS',len(build['source_files']),'凍結輸入、八封包、四路正常實錄、',counts,'及新版影片')
