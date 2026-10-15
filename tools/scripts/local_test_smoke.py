"""#92：完整版正式入口抽驗及推廣片實際畫面，沿公開CLI輸入。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main(version, full_directory=None, output_override=None):
    assert Path.cwd() == Path('/src') and os.getuid() == 1000
    root = Path('dist-all')/version
    build = json.loads((root/'BUILD.json').read_text())
    package = root/(full_directory or build.get('full_directory', 'full-local'))/('LOGH3-'+version+'-linux-amd64')
    binary = (package/'gin3').resolve()
    out = Path(output_override) if output_override else Path('workplace')/('package92-smoke-'+version)
    out.mkdir()
    assert subprocess.check_output([str(binary), '-version'], text=True).strip() == version
    base = 'click 300 210;type 138;enter;rclick 300 240;rclick 300 240'
    scripts = {
        'main': 'move 300 240',
        'copy': 'click 300 210',
        'month': 'click 300 210;type 138;enter',
        'ship': 'click 300 210;type 138;enter;rclick 300 240',
        'input': base,
        'fleet': base+';click 352 125;click 400 125;click 400 85;click 488 330;click 368 350',
        'load': base+';click 544 125;click 560 145',
        'battle': 'click 300 270;type 138;enter;click 300 322;click 318 208;click 544 125;click 560 125;tick 1000',
    }
    runs = []
    def run(name, script, level=-1, ticks=4, music=''):
        cmd = [str(binary), '-orig', str((package/'original').resolve()), '-data', str((package/'data').resolve()),
               '-music', music, '-ai', str(level), '-save', str((out/('save-'+name)).resolve()),
               '-script', script, '-shot', str((out/(name+'.png')).resolve()),
               '-shot-bin', str((out/(name+'.bin')).resolve()), '-state-out', str((out/(name+'.txt')).resolve()),
               '-shot-after', str(ticks)]
        with (out/(name+'.log')).open('x') as log:
            subprocess.run(cmd, stdout=log, stderr=subprocess.STDOUT, timeout=60, check=True)
        state = (out/(name+'.txt')).read_text()
        assert (out/(name+'.bin')).stat().st_size == 640*480
        runs.append({'name': name, 'command': cmd, 'state': state, 'png_sha256': sha(out/(name+'.png')),
                     'index_sha256': sha(out/(name+'.bin'))})
        print('封包啟動', name, state.splitlines()[0], flush=True)
    for name, script in scripts.items():
        run(name, script)
    assert 'state=StateBattle' in (out/'battle.txt').read_text()
    # 既有068的公開新局、編隊、實行、存讀與續回合序列，不覆寫歷史收據。
    match = re.search(r"^script='([^']+)'", Path('tools/enhanced_ai_check.sh').read_text(), re.M)
    assert match
    script = match.group(1).replace('click 300 122;', 'click 300 122;click 318 208;')
    for mode, level in [('original', -1), ('enhanced', 1)]:
        run('ai-'+mode, script, level)
        folder = out/('save-ai-'+mode)
        for stem in ('FLEET', 'SHIP', 'STELLAR', 'PLANET'):
            assert (folder/(stem+'.DT0')).read_bytes() == (folder/(stem+'.DT1')).read_bytes(), (mode, stem)
    ai = {}
    for mode in ('original', 'enhanced'):
        rows = []
        folder = out/('save-ai-'+mode)
        for slot in '02':
            data = (folder/('FLEET.DT'+slot)).read_bytes()
            counts = [0, 0]
            commands = []
            for i in range(32):
                r = data[i*122:(i+1)*122]
                if r[4] != 255:
                    counts[r[1]] += 1
                    if r[1] == 1:
                        commands.append([i, r[6], r[7], r[8]])
            assert counts[0] == 1 and counts[1] >= 3
            rows.append({'slot': slot, 'counts': counts, 'commands': commands})
        ai[mode] = rows
    for i in (0, 1):
        assert ai['original'][i]['commands'] != ai['enhanced'][i]['commands']
    run('music', base, -1, 600, str((package/'music').resolve()))
    result = {'version': version, 'binary_sha256': sha(binary), 'classification': '本機Linux封包正式GUI／公開CLI輸入抽驗；無新增原版對拍',
              'seed': '未注入；沿正常載入器與既有正式AI種子', 'runs': runs, 'ai': ai,
              'save_roundtrip': ['FLEET', 'SHIP', 'STELLAR', 'PLANET'], 'not_claimed': ['實體音效卡', '人耳', '原版戰場同態']}
    (out/'SMOKE.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('PASS 正式兩模式建隊／命令差異／存讀與續回合；封包音訊入口已啟動', flush=True)


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--version', required=True)
    p.add_argument('--full-directory',choices=['full','full-local'],help='既有Windows BUILD未列跨平台目錄時明確指定')
    p.add_argument('--out',help='獨立驗收輸出，避免覆寫先前失敗批次')
    args=p.parse_args()
    main(args.version,args.full_directory,args.out)
