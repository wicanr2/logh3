"""116來源重建的正常標題及FM入口抽驗，Docker內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def run(revision, mode, attempt):
    assert os.getuid() == 1000 and revision.isascii() and revision.isalnum()
    workspace = Path('/src') if Path.cwd() == Path('/src') else Path('/work')
    root = workspace/'workplace'/('source69-r103-'+revision)
    assert root.is_dir() and root.stat().st_uid == os.getuid()
    rebuild = root/'rebuild'
    record = json.loads((rebuild/'REBUILD.json').read_text())
    for row in record['binaries']:
        if not row['name'].startswith('dosgolem'):
            assert sha(rebuild/(row['name']+'-source')) == row['sha256']
    assert not attempt or (attempt.isascii() and attempt.isalnum())
    out = root/('runtime-'+mode+('-'+attempt if attempt else ''))
    assert not out.exists() and not out.is_symlink()
    out.mkdir()
    if mode == 'title':
        assert workspace == Path('/src'), '正常GUI須走tools/gui.sh'
        # 用目前正式常數及實際decoder幀數選標題等待的中點。
        # 此程式只算觀測點，不改GUI的Tick、狀態或輸入。
        planner = out/'title-plan.go'
        planner.write_text('''//go:build ignore
package main
import (
 "encoding/json"
 "os"
 "time"
 "github.com/wicanr2/logh3/internal/bootstrap"
 "github.com/wicanr2/logh3/internal/game"
 "github.com/wicanr2/logh3/internal/screen"
)
func main() {
 r,e:=bootstrap.Load("/orig","/src/docs/internals/data");if e!=nil {panic(e)}
 if e=bootstrap.LoadIntro(r,"/orig");e!=nil {panic(e)}
 if r.Intro==nil||r.Title==nil {panic("完整動畫或標題缺件")}
 var start time.Duration
 for frame:=1;frame<r.Intro.Frames();frame++ {start+=screen.IntroFrameIntervalAt(frame)}
 start+=game.IntroTitleDelay
 step:=time.Second/60
 updates:=int((start+game.TitleWaitDuration/2+step-1)/step)
 if e=json.NewEncoder(os.Stdout).Encode(map[string]any{"frames":r.Intro.Frames(),"title_start_ns":int64(start),"title_wait_ns":int64(game.TitleWaitDuration),"updates":updates});e!=nil {panic(e)}
}
''')
        timing = json.loads(subprocess.check_output(['go', 'run', str(planner)], text=True))
        (out/'title-plan.json').write_text(json.dumps(timing, indent=2)+'\n')
        print('正式時程觀測點', timing, flush=True)
        commands = []
        for side in ('reference', 'source'):
            stem = out/side
            command = [str(rebuild/('gin3-linux-'+side)), '-orig', '/orig', '-data', '/src/docs/internals/data',
                       '-save', str(out/('save-'+side)), '-shot', str(stem)+'.png', '-shot-bin', str(stem)+'.bin',
                       '-state-out', str(stem)+'.txt', '-shot-after', str(timing['updates'])]
            commands.append(command)
            with (out/(side+'.log')).open('w') as log:
                subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=120, check=True)
            assert 'state=StateTitle' in (out/(side+'.txt')).read_text(), '正常片頭未進標題'
            print('正常片頭至標題', side, flush=True)
        pix = (out/'source.bin').read_bytes()
        assert len(pix) == 307200 and pix == (out/'reference.bin').read_bytes()
        assert (out/'source.png').read_bytes() == (out/'reference.png').read_bytes()
        assert sum(v != 0 for v in pix) > 1024, '只有黑幀不能當完整動畫啟動抽驗'
        result = {'commands': commands, 'source_binary_sha256': sha(rebuild/'gin3-linux-source'),
                  'updates': timing['updates'], 'timing': timing, 'state': 'StateTitle', 'pixels': len(pix), 'different': 0,
                  'png_sha256': sha(out/'source.png'), 'script_used': False, 'state_injection': False}
    else:
        assert workspace == Path('/work') and not Path('/usr/bin/ffmpeg').exists()
        base = workspace/'workplace/music69-linux/封裝 測試'
        manifest = json.loads((base/'BUILD.json').read_text())
        for name, want in manifest['files'].items():
            assert sha(base/name) == want['sha256'] and (base/name).stat().st_size == want['bytes'], name
        package = out/'重建 工具'
        package.mkdir()
        shutil.copytree(base/'tools', package/'tools')
        for name in ('ffmpeg', 'fluidsynth', 'opl2wav'):
            shutil.copyfile(rebuild/'music-tool-linux-source', package/'tools'/name)
            (package/'tools'/name).chmod(0o755)
        for source, dest in [(rebuild/'music-import-linux-source', package/'music-import'),
                             (rebuild/'dosgolem-probe-linux', package/'tools/dosgolem-probe')]:
            shutil.copyfile(source, dest)
            dest.chmod(0o755)
        command = [str(package/'music-import'), '-source', '/orig', '-output', str(out/'匯入 曲庫'),
                   '-fm-only', '-tracks', 'BGM_0F']
        environment = {'PATH': '/none', 'HOME': '/tmp', 'LC_ALL': 'C'}
        start = {'command': command, 'environment': environment, 'base_manifest_sha256': sha(base/'BUILD.json'),
                 'tools': {str(p.relative_to(package)): sha(p) for p in sorted(package.rglob('*')) if p.is_file()},
                 'source_archive_sha256': sha(root/'source.tar')}
        (out/'START.json').write_text(json.dumps(start, ensure_ascii=False, indent=2)+'\n')
        with (out/'import.log').open('w') as log:
            subprocess.run(command, env=environment, stdout=log, stderr=subprocess.STDOUT, timeout=1200, check=True)
        data = out/'匯入 曲庫'
        fm = json.loads((data/'FM-SOURCES.json').read_text())
        assert fm['status'] == 'complete-fm-stage' and len(fm['tracks']) == 1
        track = fm['tracks'][0]
        assert track['track'] == 'BGM_0F'
        lock = json.loads((workspace/'docs/internals/data/fm-import-lock.json').read_text())['tracks']['BGM_0F']
        for key in ('vgm_sha256', 'wav_sha256', 'pcm_sha256'):
            assert track[key] == lock[key], key
        assert sha(data/'BGM_0F.vgm') == lock['vgm_sha256'] and sha(data/'BGM_0F.ogg') == lock['ogg_sha256']
        result = {**start, 'status': 'complete-source-built-fm-sample', 'fm': fm,
                  'outputs': {p.name: sha(p) for p in sorted(data.iterdir()) if p.is_file()},
                  'limit': '一首正式FM入口技術抽驗，非人耳、實機或循環驗收'}
        print('目前來源重建FM入口通過', flush=True)
    (out/'CHECK.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--revision', required=True)
    parser.add_argument('--mode', required=True, choices=('title', 'fm'))
    parser.add_argument('--attempt', default='')
    args = parser.parse_args()
    run(args.revision, args.mode, args.attempt)
