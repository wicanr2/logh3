"""規格116：目前引擎與音樂工具來源包。Docker內、離線、拒絕覆寫。"""
import argparse
import hashlib
import io
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile

import music_source_kit as common

BASE = Path('/src/workplace/music69-delivery/source-kit-vcs-root/music-source-kit.tar')
BASE_SHA = '654a292e81afb0596b370337a00c36553786f77085873e2708decf16a07af611'
README = '''# 目前引擎與音樂工具來源

MANIFEST.json記錄乾淨來源HEAD、當輪封裝器附件、完整vendor及每檔SHA-256。
logh3採完整RRSAL-1.0；第三方及dosgolem維持各自授權，不授權原版媒體。
包內不含原版遊戲、解碼素材或生成曲庫。玩家須自備合法原版及另行準備畫面規格。

使用Go 1.24.13、GOPROXY=off、GOSUMDB=off及空GOMODCACHE，在logh3目錄建置：

    CGO_ENABLED=1 GOOS=linux go build -mod=vendor -trimpath -buildvcs=false -o gin3 ./cmd/gin3
    CGO_ENABLED=0 GOOS=windows go build -mod=vendor -trimpath -buildvcs=false -o gin3.exe ./cmd/gin3
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -buildvcs=false -o music-import ./cmd/music-import
    CGO_ENABLED=0 GOOS=windows go build -mod=vendor -trimpath -buildvcs=false -o music-import.exe ./cmd/music-import
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -buildvcs=false -o music-tool ./cmd/music-tool

Linux玩家程式需要ALSA、X11及OpenGL開發工具鏈，驗收使用專案tools/Dockerfile.gui的logh3-ebiten:2。
完整建置、真實Git身份與空快取核對入口為logh3/tools/scripts/remake_source_rebuild.py。
dosgolem來源保留固定提交及淺層Git物件；進入dosgolem目錄，明設GIT_DIR／GIT_WORK_TREE，
核git rev-parse --show-toplevel、fsck、reset index及clean，才能以-buildvcs=true建置cmd/probe。
唯一rich2.test是原固定提交的測試程式，僅保存身份，不執行、不作遊戲輸入。

合成器／編碼器及共享程式庫的完整來源仍見專案docs/playtest/033、035、036。
這是來源準備，尚非正式版號／dist-all交付，不代表macOS、實機、人耳或循環已驗收。
'''


def build(head, revision):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    if not re.fullmatch('[0-9a-f]{40}', head) or not revision.isascii() or not revision.isalnum():
        raise ValueError('須明示完整Git HEAD及英數批次')
    out = Path('workplace')/('source69-r103-'+revision)
    common.owned(out.parent, True)
    if out.exists() or out.is_symlink():
        raise ValueError('拒絕覆寫既有來源批次')
    if hashlib.sha256(BASE.read_bytes()).hexdigest() != BASE_SHA:
        raise ValueError('037來源包身份不符')
    out.mkdir()
    raw = subprocess.check_output(['git', 'archive', '--format=tar', head,
                                  'go.mod', 'go.sum', 'cmd', 'internal', 'tools',
                                  'docs/internals/data/fm-import-lock.json'])
    (out/'logh3-input.tar').write_bytes(raw)
    files = common.read_git(out/'logh3-input.tar', head, 'logh3')
    for name, (data, _) in files.items():
        current = Path(name.removeprefix('logh3/'))
        common.owned(current)
        if current.read_bytes() != data:
            raise ValueError('指定HEAD與目前來源不同：'+name)
    # go mod vendor會掃整個模組。研究工具及workplace有外部oracle的Go探針，
    # 不是玩家程式依賴，不能把它們帶入正式cmd／internal的依賴圖。
    module = out/'vendor-module'
    module.mkdir()
    for name, (data, mode) in files.items():
        relative = name.removeprefix('logh3/')
        if relative in ('go.mod', 'go.sum') or relative.startswith(('cmd/', 'internal/')):
            dest = module/relative
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_bytes(data)
            dest.chmod(mode)
    vendor = (out/'vendor-input').resolve()
    subprocess.run(['go', 'mod', 'vendor', '-o', str(vendor)],
                   cwd=module, env=dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off'), check=True)
    for f in vendor.rglob('*'):
        common.owned(f, f.is_dir())
        if f.is_file():
            name = 'logh3/vendor/'+common.safe_name(f.relative_to(vendor).as_posix())
            files[name] = (f.read_bytes(), 0o644)
    with tarfile.open(BASE, 'r:') as archive:
        baseline_manifest = json.loads(archive.extractfile('MANIFEST.json').read())
        assert baseline_manifest['inputs']['dosgolem']['head'] == common.HEADS['dosgolem']
        for entry in archive.getmembers():
            if entry.name.startswith(('dosgolem/', 'licenses/')):
                common.safe_name(entry.name)
                if not entry.isfile() or entry.name in files:
                    raise ValueError('固定來源包形態或重複項錯誤')
                files[entry.name] = (archive.extractfile(entry).read(), entry.mode)
    license_data = subprocess.check_output(['git', 'show', head+':LICENSE'])
    assert license_data == Path('LICENSE').read_bytes()
    template = Path('workplace/music69-delivery/RRSAL-TEMPLATE.txt').read_bytes()
    substitutions = common.license_mapping(template.decode(), license_data.decode())
    files['LICENSE'] = files['logh3/LICENSE'] = (license_data, 0o644)
    files['README.md'] = (README.encode(), 0o644)
    attachments = {}
    for name in ['tools/scripts/remake_source_kit.py', 'tools/scripts/remake_source_rebuild.py']:
        f = Path(name)
        common.owned(f)
        data = f.read_bytes()
        old = files.get('logh3/'+name)
        if old is not None and old[0] != data:
            raise ValueError('附件與指定HEAD撞名且內容不同')
        files['logh3/'+name] = (data, 0o644)
        attachments[name] = {'sha256': common.digest(data), 'already_in_head': old is not None}
    modules = dict(re.findall(r'^# (\S+) (v\S+)$', files['logh3/vendor/modules.txt'][0].decode(), re.M))
    assert len(modules) == 10
    for module in modules:
        assert 'logh3/vendor/'+module+'/LICENSE' in files, module
    manifest = {'schema': 'logh3.music-source-kit.v1', 'date': '2026-10-05',
                'inputs': {'logh3': {'head': head, 'archive_sha256': common.digest(raw)},
                           'dosgolem': {**baseline_manifest['inputs']['dosgolem'], 'base_bundle_sha256': BASE_SHA}},
                'additional_sources_dirty': any(not v['already_in_head'] for v in attachments.values()),
                'builder_sha256': common.digest(Path(__file__).read_bytes()),
                'attachments': attachments, 'license_sha256': common.digest(license_data),
                'license_template_sha256': common.digest(template), 'license_substitutions': substitutions,
                'vendor_modules': modules,
                'files': {n: {'bytes': len(data), 'sha256': common.digest(data), 'mode': mode}
                          for n, (data, mode) in sorted(files.items())},
                'scope': '目前玩家及音樂Go入口來源準備；正式封包及macOS未完成'}
    files['MANIFEST.json'] = ((json.dumps(manifest, ensure_ascii=False, indent=2)+'\n').encode(), 0o644)
    target = out/'source.tar'
    with tarfile.open(target, 'x', format=tarfile.PAX_FORMAT) as archive:
        for name, (data, mode) in sorted(files.items()):
            common.safe_name(name)
            entry = tarfile.TarInfo(name)
            entry.size, entry.mode, entry.uid, entry.gid, entry.mtime = len(data), mode, 1000, 1000, 0
            entry.uname = entry.gname = 'source'
            archive.addfile(entry, io.BytesIO(data))
    extracted = out/'extracted'
    extracted.mkdir()
    with tarfile.open(target, 'r:') as archive:
        entries = archive.getmembers()
        assert len(entries) == len(files) and len({e.name for e in entries}) == len(entries)
        for e in entries:
            assert e.isfile() and common.safe_name(e.name) == e.name
            data = archive.extractfile(e).read()
            assert (data, e.mode) == files[e.name]
            dest = extracted/e.name
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_bytes(data)
            dest.chmod(e.mode)
            common.owned(dest)
    record = {'archive_sha256': common.digest(target.read_bytes()), 'bytes': target.stat().st_size,
              'file_count': len(files), 'all_members_verified': True, 'manifest': manifest}
    (out/'BUNDLE.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
    print('目前來源逐檔重讀', len(files), record['archive_sha256'], flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--head', required=True)
    parser.add_argument('--revision', required=True)
    args = parser.parse_args()
    build(args.head, args.revision)
