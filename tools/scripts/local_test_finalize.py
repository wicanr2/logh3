"""#92：交付前補Go授權、核來源與封包，保存影片、抽驗及完整SHA-256。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import zipfile


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')


def main(version):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    root = Path('dist-all')/version
    build = json.loads((root/'BUILD.json').read_text())
    assert build['version'] == version
    for name, expected in build['source_files'].items():
        if name.startswith(('cmd/', 'internal/')) or name in ('go.mod', 'go.sum', 'LICENSE'):
            assert sha(Path(name)) == expected['sha256'], name+'目前來源不同'
    updates = []
    for platform in ('linux', 'windows'):
        folder = root/'full-local'/('LOGH3-'+version+'-'+platform+'-amd64')
        manifest = json.loads((folder/'MANIFEST.json').read_text())
        for name, expected in manifest['files'].items():
            path = folder/name
            assert path.stat().st_uid == os.getuid() and path.stat().st_size == expected['bytes'] and sha(path) == expected['sha256']
        license_path = folder/'licenses/Go-LICENSE'
        license_bytes = Path('/usr/local/go/LICENSE').read_bytes()
        if license_path.exists():
            assert license_path.read_bytes() == license_bytes
        else:
            license_path.write_bytes(license_bytes)
        manifest['files']['licenses/Go-LICENSE'] = {'bytes': license_path.stat().st_size, 'sha256': sha(license_path)}
        write(folder/'MANIFEST.json', manifest)
        archive = root/'full-local'/(folder.name+'.zip')
        before = sha(archive)
        temporary = archive.with_suffix('.zip.partial')
        assert not temporary.exists()
        with zipfile.ZipFile(temporary, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6) as z:
            for p in sorted(folder.rglob('*')):
                if p.is_file():
                    z.write(p, str(Path(folder.name)/p.relative_to(folder)))
        with zipfile.ZipFile(temporary) as z:
            assert z.testzip() is None
            for name, expected in manifest['files'].items():
                data = z.read(folder.name+'/'+name)
                assert len(data)==expected['bytes'] and hashlib.sha256(data).hexdigest()==expected['sha256']
            for item in z.infolist():
                assert item.filename.isascii() or item.flag_bits & 0x800
        temporary.replace(archive)  # 尚未交付的本機包；未建立或覆寫tag／Release。
        updates.append({'platform': platform, 'before_zip_sha256': before, 'after_zip_sha256': sha(archive),
                        'change': '只補Go授權與更新MANIFEST，程式／資料／配樂不變'})
    smoke = root/'smoke'
    sources = [Path('workplace/package92-smoke-'+version),
               Path('workplace/music69-windows/package92-smoke-'+version),
               Path('workplace/package92-r118-b'), Path('workplace/active-ai93-r118-b')]
    for i, source in enumerate(sources):
        assert source.is_dir()
        target = smoke/['linux', 'windows', 'audio', 'tactical-ai'][i]
        assert not target.exists()
        target.mkdir()
        for path in source.rglob('*'):
            if path.is_file() and path.suffix in ('.json', '.png', '.txt', '.log'):
                relative = path.relative_to(source)
                if i==1 and relative.parts[0].startswith('LOGH3-'):
                    continue
                q=target/relative
                q.parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(path, q)
    standard = Path('workplace/round118-delivery-standard.log')
    lines = standard.read_text().splitlines()
    assert len([line for line in lines if line.startswith('ok  ')]) == 13 and not any(line.startswith('FAIL') for line in lines)
    shutil.copyfile(standard, smoke/'standard.log')
    promo_source = Path('workplace/promo92-'+version)
    probe = json.loads((promo_source/'FFPROBE.json').read_text())
    video = next(s for s in probe['streams'] if s['codec_type']=='video')
    audio = next(s for s in probe['streams'] if s['codec_type']=='audio')
    assert (video['width'],video['height'],video['codec_name'])==(1920,1080,'h264')
    assert audio['codec_name']=='aac' and audio['channels']==2
    assert abs(float(probe['format']['duration'])-60)<0.1 and abs(float(video['duration'])-60)<0.1
    volume = (promo_source/'VOLUME.txt').read_text()
    assert 'mean_volume: -inf' not in volume and 'max_volume: -inf' not in volume
    for path in promo_source.iterdir():
        if path.name.startswith(('LOGH3-', 'sample-')) or path.name in ('contact-sheet.png','FFPROBE.json','VOLUME.txt','BLACK.txt','RIGHTS.txt','SOURCES.tsv','SHA256SUMS.txt'):
            shutil.copyfile(path, root/'promo'/path.name)
    write(root/'FINALIZE.json', {'version': version, 'license_updates': updates,
                               'scope': '兩模式戰略／戰術實作及本機測試交付；不宣稱原版戰場同態',
                               'package_input_integrity': True, 'standard_packages_pass': 13, 'standard_packages_no_tests': 5,
                               'windows': 'Wine抽驗的同一二進位及資料；最後補Go授權後ZIP全項核對',
                               'audio': '正式Linux ALSA虛擬PulseAudio有聲；人耳／實機待試玩',
                               'promo': '60秒1920×1080實際畫面靜態剪輯，有聲；不公開散布'})
    build['pre_finalize_packages'] = json.loads(json.dumps(build['packages']))
    for item in build['packages']:
        path = root/item['path']
        item['bytes'], item['sha256'] = path.stat().st_size, sha(path)
    write(root/'BUILD.json', build)
    files = []
    for path in sorted(root.rglob('*')):
        if path.is_file() and (path.suffix in ('.zip','.gz','.mp4') or path.parent==root/'promo' or path.parent==root):
            if path.name=='SHA256SUMS.json': continue
            files.append({'path': str(path.relative_to(root)), 'bytes': path.stat().st_size, 'sha256': sha(path)})
    write(root/'SHA256SUMS.json', {'version': version, 'files': files})
    print('PASS 完整版、來源、授權、13套件、兩平台抽驗及60秒音畫影片；',len(files),'份交付身份')


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version', required=True)
    main(p.parse_args().version)
