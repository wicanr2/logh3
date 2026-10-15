"""148：核對兩陣營正常設定驗收，匯出README原尺寸實錄及來源索引。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil

from PIL import Image

ROOT = Path('/src')
OUT = ROOT / 'docs/images'
BATCH = ROOT / 'workplace/settings140'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--check', action='store_true')
args = p.parse_args()
assert OUT.stat().st_uid == 1000
files, receipts = [], {}
for faction in (0, 1):
    folder = BATCH / ('f' + str(faction) + '-r2')
    result = read(folder / 'RESULT.json')
    assert result['passed'] and result['normal_entry'] and result['show_enemies_default'] is False
    assert result['ai_levels'] == [-1, 0, 1, 2, 3] and result['font_styles'] == ['serif', 'sans', 'original', 'classic']
    assert result['show_enemies_functional'] and result['units_unchanged'] and result['turn_unchanged']
    assert result['binary_sha256'] == sha(ROOT / 'workplace/settings140-gin3-r2')
    receipts[str((folder / 'RESULT.json').relative_to(ROOT))] = sha(folder / 'RESULT.json')
    captures = {e['name']: e for e in result['events'] if e.get('kind') == 'capture'}
    items = [('settings-on', 'empire' if faction == 0 else 'alliance')]
    if faction == 0:
        items.append(('ai-menu', 'ai'))
    for capture, label in items:
        event = captures[capture]
        assert event['observed']['show_enemies'] and event['observed']['graphics'] == 'hd'
        source = folder / (capture + '.png')
        target = OUT / ('readme-settings-clear-' + label + '.png')
        assert sha(source) == event['sha256']
        with Image.open(source) as image:
            assert image.size == (960, 720)
        if not args.check:
            assert not target.exists(), target
            shutil.copyfile(source, target)
        assert sha(target) == sha(source)
        files.append(dict(file=target.name, sha256=sha(target), bytes=target.stat().st_size,
                          source=str(source.relative_to(ROOT)), source_observation=event['observed']))
source_files = ['cmd/gin3/compact_settings.go', 'internal/font/settings.go',
                'internal/presentation/journal.go', 'internal/presentation/settings.go',
                'tools/scripts/settings_font.py', 'tools/scripts/settings_font_receipt.py']
manifest = dict(source='148原始碼更新，Linux正常片頭至S1，非已發布v.1.4.0 ZIP截圖',
                binary_sha256=sha(ROOT / 'workplace/settings140-gin3-r2'),
                font_source=read(ROOT / 'internal/font/settings_font.SOURCE.json'),
                source_sha256={n: sha(ROOT / n) for n in source_files}, gui_receipts=receipts,
                tests=dict(command='tools/gui.sh test ./internal/... ./cmd/...', passed=15, no_test_files=4),
                files=files, limits=['未覆寫v.1.4.0正式封包、tag或Release', '未新增Windows／macOS真機或DOS字形parity'])
target = OUT / 'readme-settings-clear.json'
if args.check:
    assert read(target) == manifest
else:
    assert not target.exists()
    target.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
print('PASS 兩陣營正常設定、三張完整README實錄及字型／來源雜湊' + ('獨立核對' if args.check else '已匯出'))
