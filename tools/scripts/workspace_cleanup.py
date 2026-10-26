"""148：只清本專案快取、無玩家資料的Wine環境及已核對可還原的交付副本。Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import zipfile

ROOT = Path('/src')
WORK = ROOT/'workplace'
BATCH = WORK/'cleanup148'


def sha(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for block in iter(lambda: f.read(1024*1024), b''): h.update(block)
    return h.hexdigest()


def read(path): return json.loads(Path(path).read_text())
def write(path, value): Path(path).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')
def relative(path): return str(path.relative_to(ROOT))
def usage(): return int(subprocess.check_output(['du', '-sxB1', str(WORK)]).split()[0])


def entries(path):
    if not path.is_dir(): return [path]
    result = [path]
    for here, dirs, files in os.walk(path, followlinks=False):
        result.extend(Path(here)/n for n in dirs+files)
    return sorted(set(result))


def snapshot(path):
    h = hashlib.sha256(); allocated = 0; logical = 0; count = 0
    for p in entries(path):
        st = p.lstat()
        assert st.st_uid == st.st_gid == os.getuid(), ('擁有權不符', p)
        name = '.' if p == path else str(p.relative_to(path))
        h.update(json.dumps([name, st.st_mode, st.st_ino, st.st_size, st.st_mtime_ns, os.readlink(p) if p.is_symlink() else '']).encode())
        allocated += st.st_blocks*512; logical += st.st_size; count += 1
    return dict(fingerprint=h.hexdigest(), allocated=allocated, logical=logical, entries=count)


def safe(path):
    assert path.is_relative_to(WORK) and path != WORK and not path.is_symlink()
    assert path.resolve().is_relative_to(WORK.resolve()) and not path.is_relative_to(BATCH)


def plan():
    assert os.getuid() == os.getgid() == 1000 and not BATCH.exists()
    assert WORK.stat().st_uid == os.getuid()
    tracked = subprocess.check_output(['git', 'ls-files', '-z', 'workplace', 'dist-all']).split(b'\0')
    assert not any(tracked), '受版控檔案必須另外處理'
    BATCH.mkdir()
    before = usage(); free = shutil.disk_usage(ROOT).free
    canonical, archive_sizes, archive_cache, hash_cache = {}, {}, {}, {}
    for p in sorted((ROOT/'dist-all').glob('*/*/LOGH3-*')):
        if p.is_file() and (p.name.endswith('.zip') or p.name.endswith('.tar.gz')):
            archive_sizes.setdefault(p.stat().st_size, []).append(p)
        elif p.is_dir() and (p/'MANIFEST.json').is_file():
            z = p.with_name(p.name+'.zip')
            if z.is_file(): canonical.setdefault(p.name, []).append(p)

    def digest(path):
        key = str(path)
        if key not in hash_cache: hash_cache[key] = sha(path)
        return hash_cache[key]

    def source(folder):
        archive = folder.with_name(folder.name+'.zip')
        key = str(archive)
        if key not in archive_cache:
            with zipfile.ZipFile(archive) as z:
                assert z.testzip() is None, archive
                assert hashlib.sha256(z.read(folder.name+'/MANIFEST.json')).hexdigest() == digest(folder/'MANIFEST.json')
            archive_cache[key] = dict(archive=relative(archive), sha256=digest(archive))
        return archive_cache[key]

    rows, skipped = [], []
    def add(path, kind, restore=None):
        safe(path)
        row = dict(id=len(rows), path=relative(path), kind=kind, **snapshot(path))
        if restore: row['restore'] = restore
        rows.append(row); print('PLAN', kind, row['path'], row['allocated'], flush=True)

    def duplicate_tree(path, folder, common=False):
        files = {str(p.relative_to(path)): p for p in entries(path) if p.is_file() and not p.is_symlink()}
        if any(p.is_symlink() for p in entries(path)): return False
        expected = read(folder/'MANIFEST.json')['files']
        if not common and set(files) != set(expected) | {'MANIFEST.json'}: return False
        if common and not files: return False
        for n, p in files.items():
            target = folder/n
            if not target.is_file() or p.stat().st_size != target.stat().st_size or digest(p) != digest(target): return False
        origin = source(folder)
        add(path, 'common_copy' if common else 'package_copy', dict(**origin, member_prefix=folder.name+'/', files={n: dict(sha256=digest(p), mode=p.stat().st_mode & 0o777) for n, p in files.items()}))
        return True

    for here, dirs, names in os.walk(WORK, followlinks=False):
        path = Path(here)
        if path == BATCH or path.name in ['oracle', 'decomp', 'gomodcache']:
            dirs.clear(); continue
        if path.name == 'gocache':
            add(path, 'go_cache'); dirs.clear(); continue
        if 'system.reg' in names and (path/'drive_c').is_dir():
            user = path/'drive_c/users'
            player = [p for p in user.rglob('*') if p.is_file() and ('logh3' in str(p).lower() or re.search(r'\.dt[0-9]$', p.name, re.I) or p.name == 'preferences.json')]
            if player: skipped.append(dict(path=relative(path), reason='Wine含玩家資料，保留'))
            else: add(path, 'wine_cache')
            dirs.clear(); continue
        if path.name.startswith('LOGH3-') and 'MANIFEST.json' in names:
            matched = False
            for folder in canonical.get(path.name, []):
                if digest(path/'MANIFEST.json') == digest(folder/'MANIFEST.json') and duplicate_tree(path, folder): matched = True; break
            if not matched: skipped.append(dict(path=relative(path), reason='沒有相同正式ZIP，保留'))
            dirs.clear(); continue
        if path.name == 'common':
            match = re.fullmatch(r'package92-(v\.\d+\.\d+\.\d+-\d{8})(?:-[a-z0-9-]+)?', path.parent.name)
            if match:
                for platform in ['windows-amd64', 'linux-amd64']:
                    options = canonical.get('LOGH3-'+match[1]+'-'+platform, [])
                    folder = next((p for p in options if p.parent.name == 'full'), None)
                    if folder and duplicate_tree(path, folder, common=True): dirs.clear(); break
                if not dirs: continue
        for name in names:
            p = path/name
            if p.is_symlink() or not p.is_file() or not name.endswith(('.zip', '.tar.gz')): continue
            options = archive_sizes.get(p.stat().st_size, [])
            if not options: continue
            value = digest(p)
            match = next((q for q in options if digest(q) == value), None)
            if match: add(p, 'archive_copy', dict(archive=relative(match), sha256=value))
    result = dict(before_workplace=before, before_free=free, rows=rows, skipped=skipped,
                  policy='保留dist-all、org_game、assets、來源、原版oracle、反編譯與獨有音畫；只清快取及已逐檔核對的正式包副本。')
    write(BATCH/'PLAN.json', result)
    print(json.dumps(dict(items=len(rows), allocated=sum(r['allocated'] for r in rows), by_kind={k: sum(r['allocated'] for r in rows if r['kind']==k) for k in sorted({r['kind'] for r in rows})}), indent=2), flush=True)


def apply():
    info = read(BATCH/'PLAN.json')
    assert not (BATCH/'RESULT.json').exists()
    checked = set(); deleted = []
    backup = BATCH/'wine-state.tar.gz'
    with tarfile.open(backup, 'x:gz', dereference=False) as tar:
        for row in info['rows']:
            if row['kind'] != 'wine_cache': continue
            p = ROOT/row['path']
            for f in sorted(p.glob('*.reg')): tar.add(f, arcname=str(f.relative_to(WORK)))
            if (p/'drive_c/users').is_dir(): tar.add(p/'drive_c/users', arcname=str((p/'drive_c/users').relative_to(WORK)))
    for row in info['rows']:
        p = ROOT/row['path']; safe(p)
        assert snapshot(p) == {k: row[k] for k in ['fingerprint', 'allocated', 'logical', 'entries']}, ('計畫後內容改動，停止', p)
        if row.get('restore'):
            src = ROOT/row['restore']['archive']
            if str(src) not in checked:
                assert sha(src) == row['restore']['sha256'], src
                checked.add(str(src))
        if p.is_dir(): shutil.rmtree(p)
        else: p.unlink()
        deleted.append({k: row[k] for k in ['id', 'path', 'kind', 'allocated']})
        write(BATCH/'PROGRESS.json', deleted)
        print('REMOVED', row['kind'], row['path'], flush=True)
    after = usage()
    report = dict(passed=True, before_workplace=info['before_workplace'], after_workplace=after,
                  reclaimed=info['before_workplace']-after, after_free=shutil.disk_usage(ROOT).free,
                  deleted=deleted, plan=dict(path=relative(BATCH/'PLAN.json'), sha256=sha(BATCH/'PLAN.json')),
                  wine_state=dict(path=relative(backup), bytes=backup.stat().st_size, sha256=sha(backup)),
                  restore='tools/scripts/workspace_cleanup.py restore --id <項目ID>；Go／Wine環境由既有工具映像重建。',
                  retained=info['policy'], sources=sorted({r['restore']['archive'] for r in info['rows'] if r.get('restore')}))
    write(BATCH/'RESULT.json', report)
    write(ROOT/'docs/internals/data/workspace-cleanup148.json', report)
    print(json.dumps({k: report[k] for k in ['passed', 'before_workplace', 'after_workplace', 'reclaimed', 'after_free']}, indent=2), flush=True)


def restore(identity):
    row = read(BATCH/'PLAN.json')['rows'][identity]
    assert row['id'] == identity and row.get('restore'), '此項由既有Go／Wine工具環境重建'
    spec = row['restore']; target = ROOT/row['path']; safe(target)
    assert not target.exists() and target.parent.stat().st_uid == os.getuid()
    archive = ROOT/spec['archive']; assert sha(archive) == spec['sha256']
    if row['kind'] == 'archive_copy': shutil.copyfile(archive, target)
    else:
        target.mkdir()
        with zipfile.ZipFile(archive) as z:
            for n, meta in spec['files'].items():
                p = target/n; assert p.resolve().is_relative_to(target.resolve())
                p.parent.mkdir(parents=True, exist_ok=True)
                p.write_bytes(z.read(spec['member_prefix']+n)); p.chmod(meta['mode'])
                assert sha(p) == meta['sha256']
    print('RESTORED', row['path'])


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode', choices=['plan', 'apply', 'restore']); p.add_argument('--id', type=int)
    a = p.parse_args()
    if a.mode == 'restore': assert a.id is not None; restore(a.id)
    else: globals()[a.mode]()
