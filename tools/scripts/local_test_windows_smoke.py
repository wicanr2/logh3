"""#92：Windows實際ZIP解壓、版本及Wine正式GUI抽驗；不聲稱Windows實機。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def win(path):
    return 'Z:'+str(path.resolve()).replace('/', '\\')


def main(version, hd=False):
    assert Path.cwd() == Path('/out') and os.getuid() == 1000
    archive = Path('package92-'+version+'-windows-amd64.zip')
    out = Path('package92-smoke-'+version)
    out.mkdir()
    with zipfile.ZipFile(archive) as z:
        for entry in z.infolist():
            assert not Path(entry.filename).is_absolute() and '..' not in Path(entry.filename).parts
        assert z.testzip() is None
        z.extractall(out)
    package = out/('LOGH3-'+version+'-windows-amd64')
    manifest = json.loads((package/'MANIFEST.json').read_text())
    assert manifest['version'] == version
    for name, expected in manifest['files'].items():
        path = package/name
        assert path.stat().st_uid == os.getuid() and path.stat().st_size == expected['bytes'] and sha(path) == expected['sha256']
    readme = (package/'README.txt').read_bytes()
    assert readme.startswith(b'\xef\xbb\xbf') and b'\r\n' in readme and b'\n' not in readme.replace(b'\r\n', b'')
    binary = package/'gin3.exe'
    query = subprocess.run(['wine', str(binary.resolve()), '-version'], capture_output=True, text=True, timeout=45, check=True)
    assert query.stdout.strip() == version, query
    base = 'click 300 210;type 138;enter;rclick 300 240;rclick 300 240'
    scripts = {'main': 'move 300 240', 'input': base,
               'load': base+';click 544 125;click 560 145',
               'battle': 'click 300 270;type 138;enter;click 300 322;click 318 208;click 544 125;click 560 125;tick 1000'}
    if hd:
        scripts['month'] = 'click 300 210;type 138;enter'
        scripts['portrait'] = 'click 300 210;type 138;enter;click 448 370;click 448 370;click 448 125;click 464 165;click 400 90'
    rows = []
    for name, script in scripts.items():
        command = ['wine', str(binary.resolve()), '-orig', win(package/'original'), '-data', win(package/'data'),
                   '-music', '', '-save', win(out/('save-'+name)), '-ai', '-1', '-scale', '1.5',
                   '-script', script, '-shot', win(out/(name+'.png')), '-state-out', win(out/(name+'.txt')), '-shot-after', '4']
        if hd:
            command += ['-graphics', 'hd', '-art', win(package/'art')]
        with (out/(name+'.log')).open('x') as log:
            subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=60)
        state = (out/(name+'.txt')).read_text()
        if hd:
            import struct
            assert struct.unpack('>II', (out/(name+'.png')).read_bytes()[16:24]) == (1920, 1440)
            hd_count = len(list((package/'art/face').glob('face-*.png')))
            assert f'HD：{hd_count}/83' in (out/(name+'.log')).read_text()
        rows.append({'name': name, 'command': command, 'state': state, 'png_sha256': sha(out/(name+'.png'))})
        print('Windows封包Wine', name, state.splitlines()[0], flush=True)
    assert 'state=StateBattle' in (out/'battle.txt').read_text()
    result = {'version': version, 'classification': 'Windows ZIP完整逐檔核對及Wine正式GUI；非Windows實機／人耳',
              'archive_sha256': sha(archive), 'binary_sha256': sha(binary), 'manifest_files': len(manifest['files']),
              'wine_version': subprocess.check_output(['wine', '--version'], text=True).strip(), 'runs': rows}
    (out/'SMOKE.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('PASS Windows封包版本／UTF-8 BOM CRLF／逐檔身份及正常S3入口', flush=True)


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version', required=True)
    p.add_argument('--hd', action='store_true')
    args = p.parse_args()
    main(args.version, args.hd)
