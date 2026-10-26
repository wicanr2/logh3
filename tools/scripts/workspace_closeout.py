"""168：核對後清理重複安裝／共用素材與系統快取，保留唯一證據及現行驗收輸入。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import stat
import subprocess
import tarfile
import zipfile

import workspace_cleanup as old

ROOT=old.ROOT;WORK=old.WORK;DIST=ROOT/'dist-all';BATCH=WORK/'cleanup168'
VERSION='v.1.6.3-20261026';RECEIPT=ROOT/'docs/internals/data/workspace-cleanup168.json'
sha,read,write,snapshot,entries,relative=old.sha,old.read,old.write,old.snapshot,old.entries,old.relative
PROTECTED=[BATCH,WORK/'cleanup148',DIST/VERSION,WORK/'poster166',WORK/'macos-release166',
    WORK/('package92-'+VERSION+'-release166'),WORK/'music69-windows/full166']
SKIP={'.git','oracle','decomp','ida','gomodcache','source','src','toolchains'}

def usage():return int(subprocess.check_output(['du','-sxB1',str(ROOT)],text=True,timeout=180).split()[0])

def safe(path):
    assert any(path.is_relative_to(root)and path!=root for root in [WORK,DIST])
    assert not path.is_symlink()and path.resolve().is_relative_to(ROOT)
    assert not any(path.is_relative_to(p)or p.is_relative_to(path)for p in PROTECTED)

def inventory(path):
    result={}
    for p in entries(path):
        s=p.lstat()
        if p.is_symlink():return None
        if stat.S_ISREG(s.st_mode):
            result[str(p.relative_to(path))]=dict(bytes=s.st_size,sha256=sha(p),mode=oct(s.st_mode&0o777))
        elif not stat.S_ISDIR(s.st_mode):return None
    return result

def plan():
    assert os.getuid()==os.getgid()==1000 and not BATCH.exists()
    for p in [WORK,DIST,RECEIPT.parent]:assert(p.stat().st_uid,p.stat().st_gid)==(1000,1000)
    assert not subprocess.check_output(['git','ls-files','-z','workplace','dist-all'])
    before=usage();BATCH.mkdir();rows=[];skipped=[];archives={};catalog={};versions={};validated={};appimages={}
    def digest(p):
        key=(str(p),p.stat().st_size,p.stat().st_mtime_ns)
        if key not in validated:validated[key]=sha(p)
        return validated[key]
    for p in sorted(DIST.glob('*/*/*')):
        if p.is_file()and p.name.endswith(('.zip','.tar.gz','.AppImage')):
            archives.setdefault(p.stat().st_size,[]).append(p)
        if p.is_dir()and(p/'MANIFEST.json').is_file()and p.with_name(p.name+'.zip').is_file():
            z=p.with_name(p.name+'.zip');value=read(p/'MANIFEST.json');files=value.get('files')
            if not isinstance(files,dict)or not all(isinstance(v,dict)and 'sha256'in v and 'mode'in v for v in files.values()):continue
            row=dict(folder=p,archive=z,manifest=digest(p/'MANIFEST.json'),version=value['version'],files=files)
            catalog.setdefault(row['manifest'],[]).append(row);versions.setdefault(row['version'],[]).append(row)
    def archive_spec(row):
        archive=row['archive'];key=str(archive)
        if key not in appimages:
            with zipfile.ZipFile(archive)as z:
                assert z.testzip()is None,archive
                prefix=row['folder'].name+'/'
                assert hashlib.sha256(z.read(prefix+'MANIFEST.json')).hexdigest()==row['manifest']
                assert set(z.namelist())=={prefix+n for n in row['files']}|{prefix+'MANIFEST.json'}
                for n,m in row['files'].items():
                    data=z.read(prefix+n);assert len(data)==m['bytes']and hashlib.sha256(data).hexdigest()==m['sha256']
                    assert(z.getinfo(prefix+n).external_attr>>16)&0o777==int(m['mode'],8)
            appimages[key]=dict(archive=relative(archive),sha256=digest(archive),prefix=row['folder'].name+'/')
        return appimages[key]
    def add(path,kind,restore=None):
        safe(path)
        assert not any(path.is_relative_to(ROOT/r['path'])or(ROOT/r['path']).is_relative_to(path)for r in rows)
        value=dict(id=len(rows),path=relative(path),kind=kind,**snapshot(path))
        if restore:value['restore']=restore
        rows.append(value);print('PLAN',value['id'],kind,value['path'],value['allocated'],flush=True)
    def package(path,options,subset=False):
        actual=inventory(path)
        if not actual:return False
        for item in options:
            expected=dict(item['files']);manifest=path/'MANIFEST.json'
            if not subset:
                if not manifest.is_file()or digest(manifest)!=item['manifest']:continue
                expected['MANIFEST.json']=actual['MANIFEST.json']
                if actual!=expected:continue
            elif any(n not in expected or m!=expected[n]for n,m in actual.items()):continue
            spec=archive_spec(item)
            add(path,'common_copy'if subset else'package_copy',dict(**spec,files=actual));return True
        return False
    # Old extracted delivery folders are redundant only after verifying the ZIP
    # itself, every member, permissions and the entire on-disk installation.
    for group in catalog.values():
        for item in group:
            path=item['folder']
            if not any(path.is_relative_to(p)for p in PROTECTED):assert package(path,[item])
    image_manifests={}
    for p in DIST.glob('*/*/APPIMAGE.json'):
        value=read(p);artifact=p.parent.parent/value['package']['path']
        if artifact.is_file()and digest(artifact)==value['package']['sha256']:
            image_manifests[value['appdir_manifest_sha256']]=(artifact,value)
    for p in WORK.glob('*/APPDIR-MANIFEST.json'):
        match=image_manifests.get(digest(p))
        if not match:continue
        artifact,meta=match;expected=read(p)['files']
        for path in [p.parent/'AppDir',p.parent/'extracted/squashfs-root']:
            if not path.is_dir()or any(path.is_relative_to(q)for q in PROTECTED):continue
            if inventory(path)==expected:
                add(path,'appimage_copy',dict(archive=relative(artifact),sha256=digest(artifact),files=expected))
    for here,dirs,names in os.walk(WORK,followlinks=False):
        path=Path(here)
        if path.name in SKIP or(path/'.git').exists()or any(path.is_relative_to(p)for p in PROTECTED)or any(path.is_relative_to(ROOT/r['path'])for r in rows):dirs.clear();continue
        if path.name=='gocache':add(path,'go_cache');dirs.clear();continue
        if 'system.reg'in names and(path/'drive_c/windows').is_dir():
            assert {p.name for p in path.iterdir()}<={'drive_c','dosdevices','.update-timestamp','system.reg','user.reg','userdef.reg','wineserver'}
            add(path,'wine_cache');dirs.clear();continue
        if 'MANIFEST.json'in names:
            options=catalog.get(digest(path/'MANIFEST.json'),[])
            if options and package(path,options):dirs.clear();continue
            skipped.append(dict(path=relative(path),reason='沒有完全相同正式ZIP，保留唯一歷史副本'))
            dirs.clear();continue
        if path.name=='common':
            version=re.search(r'v\.\d+\.\d+\.\d+-\d{8}',path.parent.name)
            if version and package(path,versions.get(version[0],[]),True):dirs.clear();continue
        for name in names:
            p=path/name
            if p.is_symlink()or not p.is_file()or not name.endswith(('.zip','.tar.gz','.AppImage')):continue
            match=next((q for q in archives.get(p.stat().st_size,[])if digest(p)==digest(q)),None)
            if match:add(p,'archive_copy',dict(archive=relative(match),sha256=digest(match)))
    value=dict(before_bytes=before,before_free=shutil.disk_usage(ROOT).free,rows=rows,skipped=skipped,
        protected=[relative(p)for p in PROTECTED],
        policy='保留正式壓縮交付、影片、原始資料、來源、唯一證據及166現行驗收輸入；只清可還原副本與已備份Wine／Go快取。',
        generator=dict(path=relative(Path(__file__).resolve()),sha256=sha(Path(__file__).resolve())))
    write(BATCH/'PLAN.json',value)
    print('READY',len(rows),'items',sum(r['allocated']for r in rows),'bytes',flush=True)

def wine_backup(rows):
    target=BATCH/'wine-state.tar.gz';files={}
    with tarfile.open(target,'x:gz',dereference=False)as tar:
        for row in rows:
            if row['kind']!='wine_cache':continue
            root=ROOT/row['path']
            def keep(member):
                name=Path(member.name)
                return None if name.is_relative_to(Path(row['path'])/'drive_c/windows')else member
            tar.add(root,arcname=row['path'],filter=keep)
            for p in entries(root):
                if p.is_file()and not p.is_symlink()and not p.is_relative_to(root/'drive_c/windows'):files[relative(p)]=sha(p)
    with tarfile.open(target,'r:gz')as tar:
        for name,h in files.items():
            with tar.extractfile(name)as f:assert hashlib.file_digest(f,'sha256').hexdigest()==h
    return dict(path=relative(target),bytes=target.stat().st_size,sha256=sha(target),files=files,
        scope='保留每個Wine prefix的全部非drive_c/windows內容，含registry、使用者資料、Program Files、ProgramData與連結；Windows系統檔由既有固定映像Wineboot重建。')

def apply():
    value=read(BATCH/'PLAN.json');assert not RECEIPT.exists()and not(BATCH/'RESULT.json').exists()
    assert value['generator']['sha256']==sha(Path(__file__).resolve())
    for row in value['rows']:
        path=ROOT/row['path'];safe(path)
        assert snapshot(path)=={k:row[k]for k in['fingerprint','allocated','logical','entries']},path
    backup=wine_backup(value['rows']);checked=set();deleted=[]
    for row in value['rows']:
        path=ROOT/row['path'];safe(path)
        assert snapshot(path)=={k:row[k]for k in['fingerprint','allocated','logical','entries']},path
        if row.get('restore'):
            spec=row['restore'];src=ROOT/spec['archive']
            if str(src)not in checked:assert sha(src)==spec['sha256'];checked.add(str(src))
        if path.is_dir():shutil.rmtree(path)
        else:path.unlink()
        deleted.append({k:row[k]for k in['id','path','kind','allocated']})
        write(BATCH/'PROGRESS.json',deleted);print('REMOVED',row['id'],row['kind'],flush=True)
    after=usage();report=dict(passed=True,before_bytes=value['before_bytes'],after_bytes=after,
        reclaimed=value['before_bytes']-after,after_free=shutil.disk_usage(ROOT).free,deleted=deleted,
        plan=dict(path=relative(BATCH/'PLAN.json'),sha256=sha(BATCH/'PLAN.json')),
        wine_backup={k:v for k,v in backup.items()if k!='files'},wine_saved_files=len(backup['files']),
        policy=value['policy'],protected=value['protected'],skipped=value['skipped'],
        restore='在Docker內執行python3 tools/scripts/workspace_closeout.py restore --id <ID>；Go／Wine系統由既有映像重建，Wine非系統資料保存在備份tar。')
    write(BATCH/'WINE.json',backup);write(BATCH/'RESULT.json',report);write(RECEIPT,report)
    print('PASS reclaimed',report['reclaimed'],'bytes',flush=True)

def restore(identity):
    row=read(BATCH/'PLAN.json')['rows'][identity];assert row['id']==identity and row.get('restore')
    target=ROOT/row['path'];safe(target);assert not target.exists()
    spec=row['restore'];src=ROOT/spec['archive'];assert sha(src)==spec['sha256']
    assert(target.parent.stat().st_uid,target.parent.stat().st_gid)==(1000,1000)
    if row['kind']=='archive_copy':shutil.copyfile(src,target)
    elif row['kind']=='appimage_copy':
        temporary=BATCH/('extract-'+str(identity));temporary.mkdir()
        try:
            subprocess.run([str(src),'--appimage-extract'],cwd=temporary,stdout=subprocess.DEVNULL,check=True,timeout=180)
            extracted=temporary/'squashfs-root';assert inventory(extracted)==spec['files'];extracted.rename(target)
        finally:shutil.rmtree(temporary)
    else:
        target.mkdir()
        with zipfile.ZipFile(src)as z:
            for n,meta in spec['files'].items():
                p=target/n;assert p.resolve().is_relative_to(target.resolve());p.parent.mkdir(parents=True,exist_ok=True)
                p.write_bytes(z.read(spec['prefix']+n));p.chmod(int(meta['mode'],8));assert sha(p)==meta['sha256']
        assert inventory(target)==spec['files']
    print('RESTORED',row['id'],row['path'],flush=True)

def check():
    plan=read(BATCH/'PLAN.json');result=read(RECEIPT)
    assert result==read(BATCH/'RESULT.json')and sha(BATCH/'PLAN.json')==result['plan']['sha256']
    empty_caches=[]
    for row in plan['rows']:
        path=ROOT/row['path']
        if path.exists():
            # tools/gui.sh creates GOCACHE before entering the container.
            # An empty directory is expected; any regenerated content fails.
            assert row['kind']=='go_cache'and path.is_dir()and not any(path.iterdir()),path
            empty_caches.append(row['path'])
    origins={r['restore']['archive']:r['restore']['sha256']for r in plan['rows']if r.get('restore')}
    for n,h in origins.items():assert sha(ROOT/n)==h
    backup=read(BATCH/'WINE.json');assert sha(ROOT/backup['path'])==backup['sha256']
    print('PASS cleanup',len(plan['rows']),'deleted;',len(origins),'restore archives unchanged;',result['reclaimed'],'bytes; empty cache dirs',empty_caches,flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['plan','apply','restore','check']);p.add_argument('--id',type=int);a=p.parse_args()
    assert os.getuid()==os.getgid()==1000
    if a.mode=='restore':assert a.id is not None;restore(a.id)
    else:globals()[a.mode]()
