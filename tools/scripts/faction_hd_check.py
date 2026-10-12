#!/usr/bin/env python3
"""核對已核准HTML與正式陣營呈現，保存完整差異並產生離線實機畫面頁。"""
import argparse
import base64
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import re

from PIL import Image, ImageChops


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--root', type=Path, default=Path('/work/hd126'))
    p.add_argument('--revision', default='r4')
    p.add_argument('--receipts', action='store_true', help='正常X11與標準測試完成後，輸出凍結來源索引')
    a = p.parse_args()
    root = a.root
    assert root.stat().st_uid == os.getuid()
    source = root/'FACTION-HD-CONCEPTS.html'
    assert sha(source) == '7e6cb3610d94a9bc8ad4b3a3370ef6b6ded3f4e50c8f81e3387abc1baf83a5a7'
    data = json.loads(re.search(r'<script id="concept-data" type="application/json">(.*?)</script>',
                               source.read_text(), re.S).group(1))
    spec = importlib.util.spec_from_file_location('concept', '/src/tools/scripts/faction_hd_concepts.py')
    concept = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(concept)
    results, previews = [], {}
    for faction, theme in [('empire', 'empire-b'), ('alliance', 'alliance-b')]:
        folder = root/f'theme-{faction}-{a.revision}'
        plan = json.loads((folder/'PLAN.json').read_text())
        previews[faction] = {}
        for scene, (label, stem, spec_name) in concept.SCENES.items():
            geometry = Path('/src/docs/internals/data')/(spec_name+'.json')
            mask, _, _ = concept.ui_geometry(json.loads(geometry.read_text()))
            approved = Image.open(io.BytesIO(base64.b64decode(data['images'][theme][scene].split(',')[1]))).convert('RGB')
            actual_path = folder/(stem+'.png')
            actual = Image.open(actual_path).convert('RGB')
            assert actual.size == approved.size == (1920, 1440)
            diff = ImageChops.difference(approved, actual)
            red, green, blue = diff.split()
            changed = ImageChops.lighter(ImageChops.lighter(red, green), blue).point(lambda n: 255 if n else 0)
            changed = ImageChops.multiply(changed, mask)
            count = sum(changed.histogram()[1:])
            event = next(e for e in plan['events'] if e.get('name') == stem)
            cursors = event.get('cursors', [])
            known = 0
            pairs = {}
            for i, (old, new, m) in enumerate(zip(approved.getdata(), actual.getdata(), mask.getdata())):
                if not m or old == new:
                    continue
                x, y = i % actual.width, i // actual.width
                key = str(old)+' -> '+str(new)
                pairs[key] = pairs.get(key, 0)+1
            assert count == known == 0, (faction, scene, count, known, pairs)
            difference_path = root/f'approved-{faction}-{scene}-difference-{a.revision}.png'
            Image.composite(diff, Image.new('RGB', actual.size), mask).save(difference_path)
            results.append({'faction': faction, 'scene': scene, 'label': label,
                            'ui_pixels': sum(mask.histogram()[1:]), 'different': count,
                            'known_cursor_neighbourhood': known, 'unclassified': count-known,
                            'bbox': changed.getbbox(), 'pairs': pairs, 'cursors': cursors,
                            'source_geometry_sha256': sha(geometry), 'actual_sha256': sha(actual_path),
                            'difference_sha256': sha(difference_path)})
            previews[faction][scene] = {'label': label, 'image': 'data:image/png;base64,'+base64.b64encode(actual_path.read_bytes()).decode()}
    report = {'scope': '核准概念與正式HD呈現；包含全部UI差異，不宣稱DOS原版parity',
              'source_html_sha256': sha(source), 'revision': a.revision, 'comparisons': results}
    report_path = root/f'APPROVED-UI-COMPARISON-{a.revision}.json'
    report_path.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    html = '''<!doctype html>
<html lang="zh-Hant"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>銀河英雄傳說 III SP｜陣營 HD 實際畫面</title>
<style>
:root{color-scheme:dark;font:16px/1.6 system-ui,sans-serif;background:#10141c;color:#f3ead6}
*{box-sizing:border-box}body{margin:0}main{max-width:1440px;margin:auto;padding:24px}
h1{font-size:clamp(23px,4vw,34px);margin:0 0 10px}p{margin:0 0 18px;color:#c5c6ca}
nav{display:flex;gap:8px;flex-wrap:wrap;margin:12px 0}button{font:inherit;cursor:pointer;padding:9px 15px;border:1px solid #627081;background:#202b3a;color:#f3ead6;border-radius:4px}
button[aria-pressed=true]{background:#365d87;border-color:#d6bf8e}button:focus-visible{outline:3px solid #f3d695;outline-offset:3px}
figure{margin:20px 0 0}img{display:block;width:100%;height:auto;aspect-ratio:4/3;background:black;border:1px solid #7b8596}
figcaption{padding:12px 0;color:#d8c99f}footer{margin-top:20px;font-size:14px;color:#a7afb9}
@media(max-width:600px){main{padding:16px}button{padding:8px 10px}}
</style>
<main><h1>帝國與同盟，介面和星圖一起換裝。</h1>
<p>帝國採亮藍與高雅黃金，同盟採赤焰戰線的深褐、土金與熱情紅。保留原版字形、立體框、按鈕位置與操作方式。</p>
<nav id="factions" aria-label="陣營"></nav><nav id="scenes" aria-label="畫面"></nav>
<figure><img id="preview" alt=""><figcaption id="caption"></figcaption></figure>
<footer>這是正式遊戲的 HD 畫面預覽。兩陣營各 34 個正常操作畫面已用實際視窗核對。此頁可離線開啟，圖片已內嵌。新版影片與封包另行製作。</footer></main>
<script id="preview-data" type="application/json">__DATA__</script>
<script>
const data=JSON.parse(document.getElementById('preview-data').textContent);
const labels={empire:'銀河帝國 · 亮藍與黃金',alliance:'自由行星同盟 · 赤焰戰線'};
let faction='empire',scene='strategy';
function render(){const item=data[faction][scene],img=document.getElementById('preview');img.src=item.image;img.alt=labels[faction]+' '+item.label;document.getElementById('caption').textContent=img.alt;document.querySelectorAll('[data-faction]').forEach(b=>b.setAttribute('aria-pressed',String(b.dataset.faction===faction)));document.querySelectorAll('[data-scene]').forEach(b=>b.setAttribute('aria-pressed',String(b.dataset.scene===scene)))}
for(const [id,label] of Object.entries(labels)){const b=document.createElement('button');b.textContent=label;b.dataset.faction=id;b.onclick=()=>{faction=id;render()};document.getElementById('factions').appendChild(b)}
for(const [id,item] of Object.entries(data.empire)){const b=document.createElement('button');b.textContent=item.label;b.dataset.scene=id;b.onclick=()=>{scene=id;render()};document.getElementById('scenes').appendChild(b)}
render();</script></html>'''
    (root/'FACTION-HD-GAME.html').write_text(html.replace('__DATA__', json.dumps(previews, ensure_ascii=False)))
    print(json.dumps({'comparisons': len(results), 'ui_pixels': sum(r['ui_pixels'] for r in results),
                      'different': sum(r['different'] for r in results),
                      'unclassified': sum(r['unclassified'] for r in results)}, ensure_ascii=False))
    if a.receipts:
        artifacts = []
        windows, pixels, attempts = 0, 0, []
        for faction in ('empire', 'alliance'):
            folder = root/f'theme-x11-{faction}-{a.revision}'
            trace = json.loads((folder/'TRACE.json').read_text())
            assert trace['normal_input'] and not trace['script_flag'] and trace['scale'] == 1.5
            comparisons = [p for p in folder.glob('*-x11-comparison.json') if '-attempt-' not in p.name]
            assert len(comparisons) == 34
            for path in comparisons:
                result = json.loads(path.read_text())
                assert result['different'] == result['max_channel_difference'] == 0 and result['pixels'] == 691200
                pixels += result['pixels']
                windows += 1
            for path in folder.glob('*-attempt-*-x11-comparison.json'):
                attempts.append({'path': str(path.relative_to('/work')), 'different': json.loads(path.read_text())['different']})
            artifacts += [p for p in folder.rglob('*') if p.is_file()]
            artifacts += [p for p in (root/f'theme-{faction}-{a.revision}').rglob('*') if p.is_file()]
        tests = (root/f'faction-standard-tests-{a.revision}.log').read_text()
        passed = sum(line.startswith('ok') for line in tests.splitlines())
        no_tests = sum(line.startswith('?') for line in tests.splitlines())
        assert passed == 15 and no_tests == 4 and 'FAIL' not in tests
        browser = json.loads((root/'FACTION-GAME-BROWSER-RESULTS.json').read_text())
        assert len(browser) == 2 and all(r['pass'] and r['checks'] == 34 for r in browser)
        artifacts += [source, root/'FACTION-HD-GAME.html', root/'FACTION-GAME-BROWSER-RESULTS.json',
                      root/f'gin3-faction-{a.revision}', report_path,
                      root/f'faction-standard-tests-{a.revision}.log', root/f'faction-benchmark-{a.revision}.log']
        artifacts += list(root.glob(f'approved-*-difference-{a.revision}.png'))
        artifacts += list(root.glob('faction-game-*.png'))
        artifacts += list(root.glob(f'faction-x11-*-{a.revision}.log'))
        sources = [
            'internal/gfx/framebuffer.go', 'internal/game/game.go', 'internal/game/enemyvisibility.go',
            'internal/screen/maingame.go', 'internal/ui/bevel.go',
            'internal/presentation/artwork.go', 'internal/presentation/classic.go',
            'internal/presentation/faction.go', 'internal/presentation/journal.go', 'internal/presentation/symbols.go',
            'internal/presentation/faction_test.go', 'internal/bootstrap/hd_journal_test.go',
            'tools/scripts/hd_art_model.go', 'tools/scripts/hd_art_gui.py', 'tools/scripts/faction_hd_check.py',
            'tools/scripts/faction_hd_concepts.py', 'tools/scripts/faction_hd_concept_check.py',
            'tools/audio.sh', 'tools/gui.sh', 'tools/s1_record_display.sh',
            'assets/hd/interface-themes.json', 'assets/hd/interface-empire-approved.png',
            'assets/hd/interface-alliance-approved.png', 'assets/hd/MANIFEST.json',
            'assets/hd/material/MANIFEST.json', 'assets/hd/font/MANIFEST.json',
            'docs/spec/135-original-style-hd-ui.md', 'docs/spec/132-hd-strategy-and-symbols.md',
            'docs/playtest/113-faction-themes.md', 'docs/internals/data/original-ui-concepts.json',
        ]
        sources += ['docs/internals/data/'+v[2]+'.json' for v in concept.SCENES.values()]
        artifacts += [Path('/src')/name for name in sources]
        original = Path('/src/org_game/银河英雄传说3SP.zip')
        assert sha(original) == 'de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673'
        frozen = []
        for path in sorted(set(artifacts)):
            base = '/src' if str(path).startswith('/src/') else '/work'
            frozen.append({'path': ('workplace/' if base == '/work' else '')+str(path.relative_to(base)),
                           'bytes': path.stat().st_size, 'sha256': sha(path),
                           'uid': path.stat().st_uid, 'gid': path.stat().st_gid})
        index = {'date': '2026-10-11', 'scope': '正式陣營介面及戰略星圖；影片與新版封包未完成',
                 'git_head': '523ced1346358b1d3a9982148c569a69af953a5f', 'worktree_dirty': True,
                 'original_bundle_sha256': sha(original), 'new_dos_oracle_receipts': False,
                 'images': {'logh3-ebiten:2': 'sha256:d7f4d3e04c75fc9e9b2b3606d9f015d8f1bcde806c25d6780780450ad40827ab',
                            'logh3-hd-preview:20261011-r1': 'sha256:932b5ad00e356b4252945c35938d9b30b939109abfaaca8b6673f1a17f532aff'},
                 'x11': {'windows': windows, 'pixels': pixels, 'different': 0, 'early_attempts': attempts},
                 'tests': {'pass': passed, 'no_tests': no_tests}, 'approved_ui': report,
                 'browser': browser, 'artifacts': frozen,
                 'rights': '原版衍生HD圖片與實機畫面只留本機及目前private儲存庫；沒有新增公開發行',
                 'pending': ['戰場真實X11敵軍開關', '戰略與戰術AI比較實錄', 'HD世界觀片頭與新版S1交火推廣片', '新封包']}
        destination = root/'FACTION-THEMES-SOURCES.json'
        destination.write_text(json.dumps(index, ensure_ascii=False, indent=2)+'\n')
        print('凍結來源', len(frozen), sha(destination))


if __name__ == '__main__':
    main()
