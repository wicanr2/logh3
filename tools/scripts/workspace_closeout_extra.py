"""168：續清舊工作區內已有ZIP的副本，保留實際權限、額外資料及AppImage連結。"""
import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import zipfile

import workspace_closeout as base

ROOT,WORK,DIST,BATCH=base.ROOT,base.WORK,base.DIST,base.BATCH
PLAN=BATCH/'EXTRA-PLAN.json';RESULT=BATCH/'EXTRA-RESULT.json'
RECEIPT=ROOT/'docs/internals/data/workspace-cleanup168-extra.json'
sha,read,write,relative=base.sha,base.read,base.write,base.relative
cache={}
def digest(path):
    s=path.stat();key=(str(path),s.st_size,s.st_mtime_ns)
    if key not in cache:cache[key]=sha(path)
    return cache[key]

def inventory(path,links=False):
    files={};symlinks={}
    for p in base.entries(path):
        if p.is_symlink():
            if not links or not p.resolve().is_relative_to(path.resolve()):return None,{}
            symlinks[str(p.relative_to(path))]=os.readlink(p)
        if p.is_file():
            s=p.stat();files[str(p.relative_to(path))]=dict(bytes=s.st_size,sha256=digest(p),mode=oct(s.st_mode&0o777))
    return files,symlinks

def nodes():
    for here,dirs,names in os.walk(WORK,followlinks=False):
        path=Path(here)
        if path.name in base.SKIP or(path/'.git').exists()or any(path.is_relative_to(p)for p in base.PROTECTED):dirs.clear();continue
        yield path,dirs,names

def plan():
    assert not PLAN.exists()and not RESULT.exists()
    before=base.usage();catalog=[];seen=set();rows=[];selected={};skipped=[]
    archives=sorted(DIST.glob('*/*/*.zip'))+sorted(p/n for p,_,names in nodes()for n in names if n.endswith('.zip'))
    for archive in archives:
        with zipfile.ZipFile(archive)as z:
            manifests=[n for n in z.namelist()if n.endswith('/MANIFEST.json')and n.count('/')==1]
            if len(manifests)!=1:continue
            prefix=manifests[0].removesuffix('MANIFEST.json');data=z.read(manifests[0]);m=base.json.loads(data)
            if not isinstance(m.get('files'),dict)or not all(isinstance(v,dict)and 'sha256'in v and 'mode'in v for v in m['files'].values()):continue
            key=hashlib.sha256(data).hexdigest()
            if key in seen:continue
            seen.add(key);files=dict(m['files']);files['MANIFEST.json']=dict(bytes=len(data),sha256=key,mode='0o644')
            catalog.append(dict(archive=relative(archive),prefix=prefix,version=m['version'],files=files))
    def add(path,kind,spec):
        base.safe(path)
        assert not any(path.is_relative_to(ROOT/r['path'])or(ROOT/r['path']).is_relative_to(path)for r in rows)
        rows.append(dict(id=len(rows),path=relative(path),kind=kind,restore=spec,**base.snapshot(path)))
        print('PLAN',len(rows)-1,kind,relative(path),flush=True)
    def match(path,item,prefix='',full=False):
        actual,links=inventory(path)
        if not actual:return False
        expected={n.removeprefix(prefix):m for n,m in item['files'].items()if n.startswith(prefix)}
        if full and set(actual)!=set(expected):return False
        if any(n not in expected or(m['bytes'],m['sha256'])!=(expected[n]['bytes'],expected[n]['sha256'])for n,m in actual.items()):return False
        selected[item['archive']]=item
        add(path,'package_copy'if full else'common_copy',dict(archive=item['archive'],prefix=item['prefix']+prefix,files=actual));return True
    for path,dirs,names in nodes():
        if any(path.is_relative_to(ROOT/r['path'])for r in rows):dirs.clear();continue
        if 'MANIFEST.json'in names:
            key=digest(path/'MANIFEST.json');item=next((x for x in catalog if x['files']['MANIFEST.json']['sha256']==key),None)
            if item:
                if match(path,item,full=True):dirs.clear();continue
                for name in ['art','music','original','data','soundfont','licenses']:
                    child=path/name
                    if child.is_dir():match(child,item,name+'/')
            skipped.append(relative(path));dirs.clear();continue
        if path.name=='common':
            version=base.re.search(r'v\.\d+\.\d+\.\d+-\d{8}',path.parent.name)
            if version:
                for item in catalog:
                    if item['version']==version[0]and match(path,item):dirs.clear();break
    images={}
    for p in DIST.glob('*/*/APPIMAGE.json'):
        m=read(p);artifact=p.parent.parent/m['package']['path'];assert digest(artifact)==m['package']['sha256']
        images[m['appdir_manifest_sha256']]=(artifact,m)
    for p in WORK.glob('*/APPDIR-MANIFEST.json'):
        if digest(p)not in images:continue
        artifact,meta=images[digest(p)];expected=read(p)['files']
        for path in [p.parent/'AppDir',p.parent/'extracted/squashfs-root']:
            if not path.exists()or any(path.is_relative_to(q)for q in base.PROTECTED):continue
            actual,links=inventory(path,True)
            if actual==expected:add(path,'appimage_copy',dict(archive=relative(artifact),sha256=digest(artifact),files=actual,links=links))
    def validate(item):
        path=ROOT/item['archive']
        with zipfile.ZipFile(path)as z:
            assert {n for n in z.namelist()if not n.endswith('/')}=={item['prefix']+n for n in item['files']}
            for n,m in item['files'].items():
                data=z.read(item['prefix']+n)
                assert len(data)==m['bytes']and hashlib.sha256(data).hexdigest()==m['sha256']
        return item['archive'],digest(path)
    # Reading every member verifies its CRC; two independent archives may be
    # checked in parallel within the container's fixed CPU/memory limits.
    with ThreadPoolExecutor(max_workers=2)as pool:hashes=dict(pool.map(validate,selected.values()))
    for row in rows:
        spec=row['restore']
        if spec['archive']in hashes:spec['sha256']=hashes[spec['archive']]
    value=dict(before_bytes=before,rows=rows,skipped=skipped,
        generator=dict(path=relative(Path(__file__).resolve()),sha256=sha(Path(__file__).resolve())),
        policy='既有ZIP內每個檔案CRC及SHA通過；保留候選實際權限。只刪可完整還原的副本或素材子目錄，額外日誌／存檔／來源保持；AppImage內連結另記。')
    write(PLAN,value);print('READY',len(rows),'items',sum(r['allocated']for r in rows),'bytes',flush=True)

def apply():
    value=read(PLAN);assert not RESULT.exists()and not RECEIPT.exists()
    assert value['generator']['sha256']==sha(Path(__file__).resolve())
    origins={r['restore']['archive']:r['restore']['sha256']for r in value['rows']}
    for n,h in origins.items():assert sha(ROOT/n)==h
    for row in value['rows']:
        path=ROOT/row['path'];base.safe(path)
        assert base.snapshot(path)=={k:row[k]for k in ['fingerprint','allocated','logical','entries']}
    deleted=[]
    for row in value['rows']:
        path=ROOT/row['path'];shutil.rmtree(path)
        deleted.append({k:row[k]for k in ['id','path','kind','allocated']});write(BATCH/'EXTRA-PROGRESS.json',deleted)
        print('REMOVED',row['id'],row['path'],flush=True)
    after=base.usage();result=dict(passed=True,before_bytes=value['before_bytes'],after_bytes=after,
        reclaimed=value['before_bytes']-after,deleted=deleted,plan=dict(path=relative(PLAN),sha256=sha(PLAN)),
        policy=value['policy'],restore='Docker內執行python3 tools/scripts/workspace_closeout_extra.py restore --id <ID>')
    write(RESULT,result);write(RECEIPT,result);print('PASS reclaimed',result['reclaimed'],'bytes',flush=True)

def restore(identity):
    row=read(PLAN)['rows'][identity];assert row['id']==identity
    original_read,original_inventory=base.read,base.inventory
    base.read=lambda p:read(PLAN)if p==BATCH/'PLAN.json'else original_read(p)
    if row['kind']=='appimage_copy':base.inventory=lambda p:inventory(p,True)[0]
    try:
        base.restore(identity)
        if row['kind']=='appimage_copy':assert inventory(ROOT/row['path'],True)[1]==row['restore']['links']
    finally:base.read,base.inventory=original_read,original_inventory

def check():
    value=read(PLAN);result=read(RECEIPT);assert result==read(RESULT)and sha(PLAN)==result['plan']['sha256']
    assert all(not(ROOT/r['path']).exists()for r in value['rows'])
    for n,h in {r['restore']['archive']:r['restore']['sha256']for r in value['rows']}.items():assert sha(ROOT/n)==h
    print('PASS additional cleanup',len(value['rows']),result['reclaimed'],'bytes',flush=True)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=['plan','apply','restore','check']);p.add_argument('--id',type=int);a=p.parse_args()
    assert os.getuid()==os.getgid()==1000
    if a.mode=='restore':assert a.id is not None;restore(a.id)
    else:globals()[a.mode]()
