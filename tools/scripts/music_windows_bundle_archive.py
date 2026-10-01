"""規格 076：合併技術驗證包補齊 Go notice、ZIP 與中文搬移驗證。

bash tools/music_cross.sh python3 /tools/scripts/music_windows_bundle_archive.py。
保留原 BUNDLE.json；補充檔案用 ANNEX.json，ZIP 與解壓清單完整逐檔驗證。
沒有正式版號或 dist-all 交付聲明。
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import sys
import zipfile

ROOT = Path('/out')


def sha(p):
    assert p.is_file() and not p.is_symlink() and (p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()), p
    with p.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


assert sys.argv[1:] in [[], ['--utf8']]
utf8 = sys.argv[1:] == ['--utf8']
package = ROOT/('combined-utf8-stage/封裝 測試' if utf8 else 'combined-stage/封裝 測試')
original = json.loads((package/'BUNDLE.json').read_text())
expected = set(original['files']) | {'BUNDLE.json'}
assert {str(p.relative_to(package)) for p in package.rglob('*') if p.is_file()} == expected
for name, record in original['files'].items():
    p = package/name
    assert sha(p) == record['sha256'] and p.stat().st_size == record['bytes'], name
notice_cache = ROOT/'go-notices'
names = {'Go-LICENSE', 'Go-PATENTS', 'oggvorbis-1.0.5-LICENSE', 'vorbis-1.0.2-LICENSE', 'dosgolem-LICENSE'}
assert {p.name for p in notice_cache.iterdir()} == names
assert sha(notice_cache/'dosgolem-LICENSE') == 'd70e70a35fe613d505ea57b97c986b58e2b16c980e37b985811cdf80cf7f8ede'
annex = {'purpose': '完整 Go／解碼套件／dosgolem 授權補充，不變更工具或音源',
         'stage_bundle_sha256': sha(package/'BUNDLE.json'), 'builder_sha256': sha(Path(__file__)), 'files': {}}
for name in sorted(names):
    source, target = notice_cache/name, package/'licenses/go'/name
    assert not target.exists() and not target.is_symlink()
    source_hash = sha(source)
    target.parent.mkdir(exist_ok=True)
    shutil.copyfile(source, target)
    assert sha(target) == source_hash
    annex['files'][str(target.relative_to(package))] = {'sha256': source_hash, 'bytes': target.stat().st_size}
(package/'ANNEX.json').write_text(json.dumps(annex, ensure_ascii=False, indent=2)+'\n')
readme = (package/'README.txt').read_bytes()
assert readme.startswith(b'\xef\xbb\xbf') and b'\n' not in readme.replace(b'\r\n', b'')
archive = ROOT/('combined-utf8-tools-technical.zip' if utf8 else 'combined-tools-technical.zip')
destination = ROOT/('combined-utf8-relocated/中文 搬移' if utf8 else 'combined-relocated/中文 搬移')
assert not archive.exists() and not destination.exists(), '拒絕覆蓋封包或搬移收據'
files = {str(p.relative_to(package)): {'sha256': sha(p), 'bytes': p.stat().st_size}
         for p in package.rglob('*') if p.is_file()}
assert set(files) == expected | set(annex['files']) | {'ANNEX.json'}
with zipfile.ZipFile(archive, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
    for name in sorted(files):
        z.write(package/name, '封裝 測試/'+name)
with zipfile.ZipFile(archive) as z:
    entries = z.infolist()
    assert len(entries) == len(files) and len({e.filename for e in entries}) == len(files)
    for entry in entries:
        assert entry.filename.startswith('封裝 測試/') and entry.flag_bits & 0x800
        parts = Path(entry.filename).parts
        assert not Path(entry.filename).is_absolute() and '..' not in parts
        assert not entry.is_dir() and entry.file_size == files[str(Path(*parts[1:]))]['bytes']
    assert z.testzip() is None
    z.extractall(destination)
relocated = destination/'封裝 測試'
assert {str(p.relative_to(relocated)) for p in relocated.rglob('*') if p.is_file()} == set(files)
for name, record in files.items():
    p = relocated/name
    assert sha(p) == record['sha256'] and p.stat().st_size == record['bytes'], name
result = {'scope': 'Windows 合併工具本機 ZIP／中文搬移技術驗證；正式交付未完成',
          'archive': archive.name, 'sha256': sha(archive), 'bytes': archive.stat().st_size,
          'files': files, 'utf8_flag_all': True, 'readme_bom_crlf': True,
          'relocated_path': str(relocated.relative_to(ROOT)), 'builder_sha256': sha(Path(__file__)),
          'formal_delivery_complete': False}
(ROOT/('COMBINED-UTF8-ARCHIVE.json' if utf8 else 'COMBINED-ARCHIVE.json')).write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print('Windows 技術 ZIP 與中文搬移逐檔驗證通過',len(files),'檔，SHA-256',result['sha256'],flush=True)
