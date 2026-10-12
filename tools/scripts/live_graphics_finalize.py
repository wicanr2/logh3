"""第127輪：核准第三稿、六次F6實錄、AI實錄與跨平台封包的交付收據。Docker GUI入口。"""
import argparse,hashlib,json,os,shutil,zipfile,tarfile
from pathlib import Path
ROOT=Path('/src')
def sha(p):
 h=hashlib.sha256()
 with p.open('rb') as f:
  for b in iter(lambda:f.read(1048576),b''):h.update(b)
 return h.hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,v):
 assert p.parent.stat().st_uid==os.getuid()
 p.write_text(json.dumps(v,ensure_ascii=False,indent=2)+'\n')
def copy(a,b):
 assert a.is_file() and not a.is_symlink()
 b.parent.mkdir(parents=True,exist_ok=True)
 assert b.parent.stat().st_uid==os.getuid()
 if b.exists():assert sha(a)==sha(b),b
 else:shutil.copyfile(a,b)
def record(p):return dict(path=str(p.relative_to(ROOT)),bytes=p.stat().st_size,sha256=sha(p))
def receipts(a,b):
 for p in sorted(a.rglob('*')):
  if p.is_file() and p.suffix in ('.json','.jsonl','.png','.txt','.log') and '-attempt-' not in p.name and p.name!='WAIT.png':copy(p,b/p.relative_to(a))
def main(version):
 assert os.getuid()==1000
 batch=ROOT/'workplace/hd127'; out=ROOT/'dist-all'/version
 build=read(out/'BUILD.json')
 frozen=ROOT/f'workplace/package92-{version}-hd/source'
 for name,entry in build['source_files'].items():
  assert sha(frozen/name)==entry['sha256'],name
  if not name.startswith('vendor/'):assert sha(ROOT/name)==entry['sha256'],name
 assert len(build['packages'])==8
 manifests=[]
 for item in build['packages']:
  p=out/item['path'];assert sha(p)==item['sha256'] and p.stat().st_size==item['bytes']
  if p.suffix!='.zip':
   assert p.name.endswith('-engine-source.tar.gz')
   with tarfile.open(p,'r:gz') as tar:
    members={m.name:m for m in tar.getmembers() if m.isfile()}
    assert set(members)=={'logh3-source/'+f for f in build['source_files']}
    for f,e in build['source_files'].items():
     content=tar.extractfile(members['logh3-source/'+f]).read()
     assert len(content)==e['bytes'] and hashlib.sha256(content).hexdigest()==e['sha256'],f
   continue
  with zipfile.ZipFile(p) as z:
   assert z.testzip() is None
   name=p.stem;m=json.loads(z.read(name+'/MANIFEST.json'))
   assert m['version']==version
   assert set(z.namelist())=={name+'/MANIFEST.json',*(name+'/'+f for f in m['files'])}
   for f,e in m['files'].items():
    b=z.read(name+'/'+f)
    assert len(b)==e['bytes'] and hashlib.sha256(b).hexdigest()==e['sha256'],f
   if item['path'].startswith('release/'):
    forbidden={'.amg','.bmg','.img','.mds','.dat','.bit','.ogg','.sf2','.com','.rgn','.spr'}
    assert not any('/art/' in f or '/original/' in f or Path(f).suffix.lower() in forbidden for f in z.namelist())
   manifests.append(dict(path=item['path'],files=len(m['files']),passed=True))
 for platform,binary in [('linux-amd64','gin3'),('windows-amd64','gin3.exe'),('macos-universal','銀河英雄傳說III.app/Contents/MacOS/gin3')]:
  assert sha(out/'full'/f'LOGH3-{version}-{platform}'/binary)==sha(out/'release'/f'LOGH3-{version}-{platform}'/binary)
 art=out/'full'/f'LOGH3-{version}-linux-amd64/art'
 for p in (ROOT/'assets/hd').rglob('*'):
  if p.is_file() and p.name!='README.md':assert sha(p)==sha(art/p.relative_to(ROOT/'assets/hd')),p
 lines=(batch/'standard-tests-approved-r2.log').read_text().splitlines()
 assert not any(l.startswith(('FAIL','--- FAIL')) for l in lines)
 tests=dict(passed=sum(l.startswith('ok ') for l in lines),no_tests=sum(l.startswith('? ') for l in lines))
 assert tests==dict(passed=15,no_tests=4)
 copy(batch/'standard-tests-approved-r2.log',out/'smoke/standard-tests.log')
 copy(batch/'CLEANUP.json',out/'smoke/CLEANUP.json')
 receipts(batch/'music-timeout-diagnostic',out/'smoke/environment-diagnostic')
 maps=read(batch/'approved-map-r2/CHECK.json')
 assert len(maps)==2 and all(r['same_approved_pixels'] and r['different_pixels']==0 for r in maps)
 receipts(batch/'approved-map-r2',out/'smoke/approved-map')
 copy(batch/'MAP-NAMES-r3.json',out/'smoke/MAP-NAMES.json')
 copy(batch/'CLARITY-COMPARISON.html',out/'promo/CLARITY-COMPARISON.html')
 mac=read(out/'smoke/macos/MACOS.json')
 assert {s['architecture'] for s in mac['slices']}=={'amd64','arm64'}
 assert all(c['all_page_hashes_match'] for s in mac['slices'] for c in s['code_directory'])
 linux=read(ROOT/f'workplace/package92-smoke-{version}/SMOKE.json')
 wine=read(ROOT/f'workplace/music69-windows/package92-smoke-{version}/SMOKE.json')
 assert len(linux['runs'])==11 and len(wine['runs'])==6
 receipts(ROOT/f'workplace/package92-smoke-{version}',out/'smoke/full-linux')
 receipts(ROOT/f'workplace/music69-windows/package92-smoke-{version}',out/'smoke/full-windows-wine')
 receipts(ROOT/f'workplace/music69-windows/release-smoke-{version}',out/'smoke/release-windows-wine')
 receipts(batch/'scenarios',out/'smoke/scenarios')
 assert len(read(batch/'scenarios/SCENARIOS.json')['runs'])==4
 assert len(read(out/'smoke/release-linux/SMOKE.json')['runs'])==4
 assert len(read(out/'smoke/release-windows-wine/SMOKE.json')['runs'])==4
 ai=read(batch/'ai-model/AI.json');strategy=read(batch/'strategy-model/STRATEGY.json')
 assert ai['same_initial_tables'] and strategy['same_initial_tables']
 assert strategy['results']['original']['commands']!=strategy['results']['enhanced']['commands']
 recordings=[]
 binary=out/'full'/f'LOGH3-{version}-linux-amd64/gin3'
 for name,count in [('s1-live-r7',0),('ai-record-original',10),('ai-record-enhanced',10),('strategy-record-original',6),('strategy-record-enhanced',6)]:
  folder=batch/name;t=read(folder/'TRACE.json')
  assert t['normal_x11_input'] and not any(t[k] for k in ('script_flag','state_injection','seed_injection'))
  assert t['binary_sha256']==sha(binary) and t['raw_sha256']==sha(folder/'RAW.mp4')
  checks=[read(p) for p in folder.glob('*-comparison.json') if '-attempt-' not in p.name]
  assert len(checks)==count and all(c['different']==0 for c in checks),(name,len(checks))
  if name.startswith('ai-record'):
   assert sha(folder/'entry-off.png')==sha(folder/'entry-off-restored.png')
   assert sha(folder/'entry-on.png')!=sha(folder/'entry-off.png')
  copy(folder/'RAW.mp4',out/'promo'/('RAW-'+name+'.mp4'))
  receipts(folder,out/'smoke'/name)
  recordings.append(dict(name=name,comparisons=count,raw_sha256=t['raw_sha256'],binary_sha256=t['binary_sha256']))
 live=read(batch/'s1-live-r7/AUDIT.json')
 assert live['display_switch_passed'] and len(live['switches'])==6
 volumes=read(batch/'RECORDINGS-AUDIO.json')
 assert len(volumes)==5 and all(r['passed'] and -55<r['mean_db']<-5 for r in volumes)
 promo=batch/'promo-final';audit=read(promo/'AUDIT.json')
 movie=promo/f'LOGH3-{version}-promo.mp4'
 assert audit['passed'] and audit['movie_sha256']==sha(movie)
 assert audit['video']==[1920,1080,'30fps'] and not audit['black_segments']
 assert len(audit['source_frame_samples'])>=20
 visual=read(promo/'VISUAL-REVIEW.json');assert visual['passed'] and len(visual['samples'])>=12
 for p in promo.iterdir():
  if p.is_file() and (p.suffix in ('.json','.txt','.png') or p==movie):copy(p,out/'promo'/p.name)
 for p in [batch/'RECORDINGS-AUDIO.json',batch/'PROMO-PLAN.json',batch/'s1-live-plan-r3.json',ROOT/'workplace/hd126/FLAG-BACKGROUND-HD.json']:
  copy(p,out/'promo'/p.name)
 copy(batch/'ai-model/AI.json',out/'smoke/AI.json')
 copy(batch/'strategy-model/STRATEGY.json',out/'smoke/STRATEGY.json')
 check=dict(version=version,passed=True,manifests=manifests,tests=tests,maps=maps,macos=mac,
  normal_recordings=recordings,live_switch=live,strategic_ai=strategy,promo=audit,visual_review=visual,
  platform_smokes=dict(linux_full=11,windows_wine_full=6,linux_engine=4,windows_wine_engine=4,scenarios=4))
 write(out/'smoke/PACKAGE-CHECK.json',check)
 paths={ROOT/n for n in build['source_files'] if not n.startswith('vendor/')}
 paths.update(p for p in (ROOT/'assets/hd').rglob('*') if p.is_file())
 paths.update(ROOT/'tools/scripts'/n for n in ['live_graphics_finalize.py','live_graphics_audit.py','approved_map_check.go',
 'ai_gameplay_record.py','ai_promo_model.go','strategy_ai_promo_model.go','faction_story_promo.py',
 'local_test_package.py','local_test_macos_build.sh','local_test_macos_package.py','local_test_smoke.py',
 'local_test_windows_smoke.py','v1_release.py','v1_release_smoke.py','bundled_scenarios_smoke.py'])
 paths.update(ROOT/n for n in ['tools/gui.sh','tools/audio.sh','tools/macos.sh','tools/player_wine.sh',
 'tools/s1_record_display.sh','docs/spec/136-live-graphics-and-cadence.md','docs/spec/137-original-symbols-and-clear-map.md',
 'docs/playtest/116-live-graphics-review.md','docs/goal/2026-10-11-map-icons-cadence-round-127.md'])
 snapshot=batch/'delivery116-source'
 for p in sorted(paths):copy(p,snapshot/p.relative_to(ROOT))
 index=dict(version=version,status='PASS 核准第三稿、原版圖示、即時F6、原生動畫計時與新版實錄交付',
 source_head=build['head'],dirty_source_snapshot=True,frozen_build_source=str(frozen.relative_to(ROOT)),
 frozen_delivery_source=str(snapshot.relative_to(ROOT)),delivery=str(out.relative_to(ROOT)),
 packages=build['packages'],sources=[record(p) for p in sorted(paths)],
 evidence=[record(p) for d in (out/'smoke',out/'promo') for p in sorted(d.rglob('*')) if p.is_file()],
 verification=check,rights='完整版、美術及影片僅限本機或授權私有範圍；引擎包為RRSAL-1.0，不含原版素材',
 limits=['DOS全戰場same-state','完整登陸分段流程#48','Windows／macOS真機','人耳與實體音效卡','逐幀AI新繪'],
 remote_release=False,cleanup=read(batch/'CLEANUP.json'))
 write(ROOT/'docs/internals/data/live-graphics-review.json',index)
 write(out/'DELIVERY.json',index)
 write(out/'SHA256SUMS.json',dict(version=version,files=build['packages']+[
 dict(path=str(p.relative_to(out)),bytes=p.stat().st_size,sha256=sha(p)) for p in sorted((out/'promo').glob('*.mp4'))]))
 print('PASS 核准第三稿、六次正常F6、五段新版實錄、三平台八封包及HD推廣片',flush=True)
if __name__=='__main__':
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--version',required=True);main(p.parse_args().version)
