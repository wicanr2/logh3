"""規格 077：從固定 Git archive 與離線 vendor 建立不含原版素材的來源套件。

LOGH3_EXTRA_RO 指向專案，tools/py.sh /work/tools/scripts/music_source_kit.py。
預設拒絕覆寫；--output 可在既有 music69-delivery 下指定新的獨立重跑目錄。
"""
import argparse
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import tarfile

ROOT = Path('/work/workplace/music69-delivery')
REPO = Path('/ref')
HEADS = {'logh3': '0f088320749fb791e5e156f85da3c424264b3185',
         'dosgolem': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c'}


def owned(path, directory=False):
    path = Path(path)
    s = path.lstat()
    if path.is_symlink() or (s.st_uid, s.st_gid) != (os.getuid(), os.getgid()):
        raise ValueError(f'連結或擁有權錯誤：{path}')
    if not (path.is_dir() if directory else path.is_file()):
        raise ValueError(f'檔案形態錯誤：{path}')


def parent_owned(path):
    path = Path(path)
    while path != Path('/work/workplace'):
        if path.exists() or path.is_symlink():
            owned(path, True)
        if Path('/work/workplace') not in path.parents:
            raise ValueError('輸出必須在 workplace 下')
        path = path.parent
    owned(path, True)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def safe_name(name):
    parts = PurePosixPath(name).parts
    if not name or name == '.' or '\\' in name or '\x00' in name or name.startswith('/') or '..' in parts or str(PurePosixPath(name)) != name:
        raise ValueError(f'不安全的來源路徑：{name!r}')
    return name


def read_git(path, head, project):
    owned(path)
    result, seen = {}, set()
    with tarfile.open(path, 'r:') as archive:
        if archive.pax_headers.get('comment') != head:
            raise ValueError('Git archive 的來源提交不符')
        for item in archive.getmembers():
            name = safe_name(item.name.rstrip('/') if item.isdir() else item.name)
            if name in seen:
                raise ValueError('來源 member 重複')
            seen.add(name)
            allowed = (name in ('go.mod', 'go.sum', 'docs/internals/data/fm-import-lock.json') or name.split('/')[0] in ('cmd', 'internal', 'tools') or (item.isdir() and name in ('docs', 'docs/internals', 'docs/internals/data'))) if project == 'logh3' else (name in ('.gitignore', 'go.mod', 'go.sum', 'LICENSE', 'rich2.test') or Path(name).suffix in ('.go', '.md', '.json', '.sh') or item.isdir())
            if not allowed or (not item.isdir() and not item.isfile()) or item.size > 64*1024*1024:
                raise ValueError(f'來源形態或範圍不符：{name}')
            if item.isdir():
                continue
            data = archive.extractfile(item).read()
            if len(data) != item.size:
                raise ValueError('來源 member 長度錯誤')
            result[project+'/'+name] = (data, 0o755 if item.mode & 0o111 else 0o644)
    for name in ['go.mod'] + (['go.sum', 'docs/internals/data/fm-import-lock.json'] if project == 'logh3' else ['LICENSE']):
        if project+'/'+name not in result:
            raise ValueError(f'必要來源缺件：{project}/{name}')
    return result


def license_mapping(template, license_text):
    pieces = re.split(r'(@[A-Z_]+@)', template)
    tokens, pattern = {}, ''
    for piece in pieces:
        if re.fullmatch(r'@[A-Z_]+@', piece):
            if piece not in tokens:
                tokens[piece] = 'p'+str(len(tokens))
                pattern += '(?P<'+tokens[piece]+'>[\\s\\S]+?)'
            else:
                pattern += '(?P='+tokens[piece]+')'
        else:
            pattern += re.escape(piece)
    match = re.fullmatch(pattern, license_text)
    if not match or re.search(r'@[A-Z_]+@', license_text):
        raise ValueError('授權範本只有占位符可以替換')
    values = {token: match.group(group) for token, group in tokens.items()}
    if values['@REPO@'] != 'https://github.com/wicanr2/logh3' or values['@YEAR@'] != '2026':
        raise ValueError('授權專案身份不符')
    return values


README = '''# 音樂工具來源套件

本套件包含 logh3 的 Go 程式與工具來源、固定 dosgolem Go 來源，以及鎖版 Go 依賴。
根目錄 MANIFEST.json 記錄來源提交、額外授權／封裝器附件及逐檔 SHA-256。
logh3 的授權為完整 RRSAL-1.0，第三方元件依自身 LICENSE／PATENTS。
不含原版遊戲、解碼素材或生成曲庫；玩家須自備合法原版。

以下使用 Go 1.24.13；實際驗收在無網路 Docker 容器執行，空模組快取及 -mod=vendor。
進入 logh3/ 後，以 CGO_ENABLED=0、GOARCH=amd64、GOPROXY=off 建置：

    GOOS=linux go build -mod=vendor -trimpath -buildvcs=false -o music-import ./cmd/music-import
    GOOS=windows go build -mod=vendor -trimpath -buildvcs=false -o music-import.exe ./cmd/music-import
    GOOS=linux go build -mod=vendor -trimpath -buildvcs=false -o music-tool ./cmd/music-tool

進入 dosgolem/ 後，先檢查來源物件並重建 index：

    test "$(git rev-parse --show-toplevel)" = "$PWD"
    git fsck --full
    git reset --mixed HEAD
    git status --porcelain

status 必須空白，HEAD 必須為 MANIFEST.json 的固定 f7523f9；僅此一個提交，不含先前歷史。
唯一 rich2.test 為固定提交既有 Go 測試程式，保留以核對來源樹，請勿執行。
用 GOOS=windows（或 linux）、GOARCH=amd64、CGO_ENABLED=0 建置：

    go build -mod=readonly -trimpath -buildvcs=true -o dosgolem-probe.exe ./cmd/probe

dosgolem 必須保留真實 Git 建置身份，-buildvcs=false 產物會被正式 FM 匯入器拒絕。

這份套件可重建指定 Go 入口。合成器／編碼器及平台共享程式庫的完整來源另由固定來源包提供，
不能用 vendor 子集合冒稱已重建全部執行期。Windows 固定來源 tar 與 Linux Debian 來源 tar
的入口、固定版本與雜湊見專案 docs/playtest/033、035、036，以及工作項目：
https://github.com/wicanr2/logh3/issues/69。
本套件內的 tools/ 仍遵守原專案 Docker 包裝器；重生原版收據需要另外備妥合法原版與固定來源。

這是交付準備的技術來源套件。正式版號／dist-all、macOS、Windows 實機、聆聽與循環仍待驗收。
'''


def build(output):
    output = Path(output)
    if output.parent != ROOT or output.exists() or output.is_symlink():
        raise ValueError('只接受新的來源套件目錄，不覆寫歷史收據')
    parent_owned(output.parent)
    files, inputs = {}, {}
    for project, head in HEADS.items():
        base = ROOT/('logh3-source-reviewed.tar' if project == 'logh3' else 'dosgolem-source-full.tar')
        files.update(read_git(base, head, project))
        inputs[project] = {'head': head, 'archive_sha256': digest(base.read_bytes())}
    git_source = ROOT/'dosgolem-source-git'
    owned(git_source, True)
    pack = git_source/'objects/pack'
    owned(pack, True)
    names = list(pack.iterdir())
    if {p.suffix for p in names} not in ({'.pack', '.idx'}, {'.pack', '.idx', '.rev'}) or len({p.stem for p in names}) != 1 or len(names) not in (2, 3):
        raise ValueError('淺層 Git 物件必須為一組完整 pack／index，可附反向索引')
    for p in names:
        owned(p)
        files['dosgolem/.git/objects/pack/'+safe_name(p.name)] = (p.read_bytes(), 0o644)
    shallow = git_source/'shallow'
    owned(shallow)
    if shallow.read_text() != HEADS['dosgolem']+'\n':
        raise ValueError('Git 歷史必須只保留固定提交')
    for name, data in {'HEAD': b'ref: refs/heads/logh3-oracle\n',
                       'refs/heads/logh3-oracle': (HEADS['dosgolem']+'\n').encode(),
                       'shallow': shallow.read_bytes(),
                       'config': b'[core]\n\trepositoryformatversion = 0\n\tbare = false\n\tfilemode = true\n'}.items():
        files['dosgolem/.git/'+name] = (data, 0o644)
    template_path, license_path = ROOT/'RRSAL-TEMPLATE.txt', REPO/'LICENSE'
    for p in [template_path, license_path, Path(__file__)]:
        owned(p)
    template, license_data = template_path.read_bytes(), license_path.read_bytes()
    substitutions = license_mapping(template.decode(), license_data.decode())
    for name, data in [('logh3/LICENSE', license_data), ('LICENSE', license_data),
                       ('README.md', README.encode()), ('logh3/tools/scripts/music_source_kit.py', Path(__file__).read_bytes())]:
        if name in files:
            raise ValueError('附件與乾淨來源撞名')
        files[name] = (data, 0o644)
    vendor = ROOT/'go-vendor-candidate'
    owned(vendor, True)
    for folder, dirs, names in os.walk(vendor):
        for name in dirs:
            owned(Path(folder)/name, True)
        for name in names:
            p = Path(folder)/name
            owned(p)
            relative = safe_name(p.relative_to(vendor).as_posix())
            files['logh3/vendor/'+relative] = (p.read_bytes(), 0o755 if p.stat().st_mode & 0o111 else 0o644)
    modules = re.findall(r'^# (\S+) (v\S+)$', files['logh3/vendor/modules.txt'][0].decode(), re.M)
    if len(modules) != 10:
        raise ValueError('Go 依賴數不符')
    for module, version in modules:
        if 'logh3/vendor/'+module+'/LICENSE' not in files:
            raise ValueError(f'第三方授權缺件：{module}')
    notices = ROOT.parent/'music69-windows/go-notices'
    for name in ['Go-LICENSE', 'Go-PATENTS']:
        p = notices/name
        owned(p)
        files['licenses/'+name] = (p.read_bytes(), 0o644)
    for name in files:
        safe_name(name)
        if name.startswith(('org_game/', 'assets/', 'workplace/', 'docs/images/', 'docs/lore/')):
            raise ValueError('禁止收錄原版素材或執行產物')
    manifest = {'schema': 'logh3.music-source-kit.v1', 'date': '2026-10-02', 'inputs': inputs,
                'additional_sources_dirty': True, 'builder_sha256': digest(Path(__file__).read_bytes()),
                'license_template_sha256': digest(template), 'license_sha256': digest(license_data),
                'license_substitutions': substitutions, 'vendor_modules': dict(modules),
                'files': {n: {'sha256': digest(data), 'bytes': len(data), 'mode': mode} for n, (data, mode) in sorted(files.items())},
                'scope': '指定音樂 Go 入口的來源與授權交付準備；正式發行未完成'}
    files['MANIFEST.json'] = ((json.dumps(manifest, ensure_ascii=False, indent=2)+'\n').encode(), 0o644)
    output.mkdir()
    target = output/'music-source-kit.tar'
    with tarfile.open(target, 'x', format=tarfile.PAX_FORMAT) as archive:
        for name, (data, mode) in sorted(files.items()):
            entry = tarfile.TarInfo(name)
            entry.size, entry.mode, entry.uid, entry.gid, entry.mtime = len(data), mode, 1000, 1000, 0
            entry.uname = entry.gname = 'source'
            archive.addfile(entry, io.BytesIO(data))
    unpacked = output/'extracted'
    unpacked.mkdir()
    with tarfile.open(target, 'r:') as archive:
        entries = archive.getmembers()
        if len(entries) != len(files) or {p.name for p in entries} != set(files):
            raise ValueError('來源包 member 不完整')
        for entry in entries:
            if not entry.isfile() or entry.name != safe_name(entry.name):
                raise ValueError('來源包形態錯誤')
            data, mode = files[entry.name]
            if entry.size != len(data) or digest(archive.extractfile(entry).read()) != digest(data):
                raise ValueError('來源包逐檔重讀不符')
            p = unpacked/entry.name
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_bytes(data)
            p.chmod(mode)
            owned(p)
    record = {'archive': target.name, 'archive_sha256': digest(target.read_bytes()), 'bytes': target.stat().st_size,
              'file_count': len(files), 'manifest_sha256': digest(files['MANIFEST.json'][0]),
              'manifest': manifest, 'all_members_verified': True}
    (output/'BUNDLE.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
    print('音樂來源套件逐檔重讀通過', len(files), '檔', record['archive_sha256'], flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=ROOT/'source-kit-vcs-root')
    args = parser.parse_args()
    build(args.output)
