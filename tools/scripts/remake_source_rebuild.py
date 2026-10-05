"""規格116：來源與目前程式的離線完整二進位比較，含玩家入口啟動。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

import music_source_kit as common


def rebuild(revision):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    assert revision.isascii() and revision.isalnum()
    root = Path('workplace')/('source69-r103-'+revision)
    common.owned(root, True)
    source = (root/'extracted/logh3').resolve()
    manifest = json.loads((root/'extracted/MANIFEST.json').read_text())
    for name, expected in manifest['files'].items():
        f = root/'extracted'/name
        common.owned(f)
        assert f.stat().st_size == expected['bytes'] and common.digest(f.read_bytes()) == expected['sha256'], name
    out = (root/'rebuild').resolve()
    assert not out.exists() and not out.is_symlink(), '拒絕覆寫重建'
    out.mkdir()
    reference = out/'reference'
    reference.mkdir()
    for name in ('cmd', 'internal'):
        shutil.copytree(name, reference/name)
    for name in ('go.mod', 'go.sum', 'LICENSE'):
        shutil.copyfile(name, reference/name)
    lock = reference/'docs/internals/data/fm-import-lock.json'
    lock.parent.mkdir(parents=True)
    shutil.copyfile('docs/internals/data/fm-import-lock.json', lock)
    subprocess.run(['go', 'mod', 'vendor', '-o', str(reference/'vendor')],
                   cwd=reference, env=dict(os.environ, GOPROXY='off', GOSUMDB='off', GOWORK='off'), check=True)
    def files(folder):
        return {str(p.relative_to(folder)): common.digest(p.read_bytes()) for p in folder.rglob('*') if p.is_file()}
    assert files(reference/'vendor') == files(source/'vendor'), '獨立vendor不同'
    for folder in ('cmd', 'internal'):
        assert files(reference/folder) == files(source/folder), '目前程式與指定HEAD不同'
    records = []
    with tempfile.TemporaryDirectory(prefix='logh3-source-r103-') as cache:
        env = dict(os.environ, GOMODCACHE=cache, GOPROXY='off', GOSUMDB='off', GOWORK='off',
                   GOFLAGS='', GOARCH='amd64', CGO_ENABLED='0')
        assert not list(Path(cache).iterdir())
        (out/'go-version.txt').write_bytes(subprocess.check_output(['go', 'version'], env=env))
        for name, platform, cgo, target in [('gin3-linux', 'linux', '1', './cmd/gin3'),
                                           ('gin3-windows.exe', 'windows', '0', './cmd/gin3'),
                                           ('music-import-linux', 'linux', '0', './cmd/music-import'),
                                           ('music-import-windows.exe', 'windows', '0', './cmd/music-import'),
                                           ('music-tool-linux', 'linux', '0', './cmd/music-tool')]:
            binaries = []
            for side, directory in [('reference', reference), ('source', source)]:
                output = out/(name+'-'+side)
                command = ['go', 'build', '-mod=vendor', '-trimpath', '-buildvcs=false', '-o', str(output), target]
                subprocess.run(command, cwd=directory, env=dict(env, GOOS=platform, CGO_ENABLED=cgo), check=True)
                binaries.append(output.read_bytes())
            assert binaries[0] == binaries[1], name+'完整二進位不同'
            records.append({'name': name, 'platform': platform, 'cgo': cgo,
                            'bytes': len(binaries[0]), 'sha256': common.digest(binaries[0]), 'equal': True})
            print('來源重建相同', name, flush=True)
        dos = (root/'extracted/dosgolem').resolve()
        git_env = dict(env, GIT_DIR=str(dos/'.git'), GIT_WORK_TREE=str(dos))
        assert subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], cwd=dos, env=git_env, text=True).strip() == str(dos)
        subprocess.run(['git', 'fsck', '--full'], cwd=dos, env=git_env, check=True)
        subprocess.run(['git', 'reset', '--mixed', 'HEAD'], cwd=dos, env=git_env, check=True)
        assert subprocess.check_output(['git', 'rev-list', '--count', 'HEAD'], cwd=dos, env=git_env, text=True).strip() == '1'
        assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=dos, env=git_env, text=True).strip() == common.HEADS['dosgolem']
        assert not subprocess.check_output(['git', 'status', '--porcelain'], cwd=dos, env=git_env)
        for platform in ('linux', 'windows'):
            output = out/('dosgolem-probe-'+platform)
            subprocess.run(['go', 'build', '-mod=readonly', '-trimpath', '-buildvcs=true', '-o', str(output), './cmd/probe'],
                           cwd=dos, env=dict(git_env, GOOS=platform), check=True)
            buildinfo = subprocess.check_output(['go', 'version', '-m', str(output)], env=env, text=True)
            (out/('dosgolem-'+platform+'-buildinfo.txt')).write_text(buildinfo)
            assert 'vcs.revision='+common.HEADS['dosgolem'] in buildinfo and 'vcs.modified=false' in buildinfo
            baseline = Path('workplace/music69-delivery')/('dosgolem-probe-'+platform+'-vcs-reference')
            assert baseline.read_bytes() == output.read_bytes(), '固定dosgolem二進位不同'
            records.append({'name': 'dosgolem-probe-'+platform, 'sha256': common.digest(output.read_bytes()), 'equal': True})
        probe_dir = source/'workplace/music69-fm/tools'
        probe_dir.mkdir(parents=True)
        shutil.copyfile(out/'dosgolem-probe-linux', probe_dir/'dosgolem-probe')
        (probe_dir/'dosgolem-probe').chmod(0o755)
        with (out/'cli-tests.jsonl').open('w') as log:
            subprocess.run(['go', 'test', '-mod=vendor', '-count=1', '-json', './cmd/music-import', './cmd/music-tool'],
                           cwd=source, env=dict(env, GOOS='linux'), stdout=log, check=True)
        tests = [json.loads(line) for line in (out/'cli-tests.jsonl').read_text().splitlines()]
        assert any(t.get('Test') == 'TestFMProbeIdentity' and t['Action'] == 'pass' for t in tests)
        assert not any(t['Action'] in ('fail', 'skip') for t in tests)
        assert not list(Path(cache).iterdir()), '來源建置讀取了外部模組'
    # 正常片頭，不使用-script。原版及畫面規格明示為獨立本機輸入。
    for side in ('reference', 'source'):
        stem = out/('intro-'+side)
        command = [str(out/('gin3-linux-'+side)), '-orig', '/orig', '-data', '/src/docs/internals/data',
                   '-save', str(out/('player-save-'+side)), '-shot', str(stem)+'.png',
                   '-shot-bin', str(stem)+'.bin', '-state-out', str(stem)+'.txt', '-shot-after', '8']
        with (out/('intro-'+side+'.log')).open('w') as log:
            subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=60, check=True)
    assert (out/'intro-source.bin').read_bytes() == (out/'intro-reference.bin').read_bytes()
    assert (out/'intro-source.png').read_bytes() == (out/'intro-reference.png').read_bytes()
    assert len((out/'intro-source.bin').read_bytes()) == 640*480
    assert 'state=StateIntro' in (out/'intro-source.txt').read_text()
    (out/'REBUILD.json').write_text(json.dumps({'binaries': records, 'offline_empty_module_cache': True,
        'vendor_equal': True, 'cli_pass': sum(t['Action'] == 'pass' and 'Test' in t for t in tests),
        'cli_skip': 0, 'intro_pixels_equal': 640*480, 'script_flag_used': False,
        'scope': '來源重建及Linux正常片頭；Windows玩家程式僅交叉建置，實機與正式封包未驗'},
        ensure_ascii=False, indent=2)+'\n')
    print('來源入口及正常片頭通過', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--revision', required=True)
    rebuild(parser.parse_args().revision)
