"""HD首批盤點及可操作試樣。素材只讀，正式遊戲程式保持獨立。"""
import argparse
import base64
from collections import Counter, defaultdict
import hashlib
import json
import os
from pathlib import Path
import struct

ROOT = Path('/src')
OUT = Path('/work/hd123-s1-20261011/hd')
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--refresh-preview', action='store_true')
args = parser.parse_args()
assert os.getuid() == 1000
if args.refresh_preview:
    assert (OUT/'INVENTORY.json').is_file()
else:
    assert not OUT.exists()
    OUT.mkdir()
assert OUT.stat().st_uid == os.getuid()
families = []
records = []
for family in ([] if args.refresh_preview else sorted((ROOT/'assets/original').iterdir())):
    if not family.is_dir():
        continue
    signatures = defaultdict(list)
    dimensions = Counter()
    for p in sorted(family.rglob('*.png')):
        data = p.read_bytes()
        assert data[:8] == b'\x89PNG\r\n\x1a\n'
        w, h = struct.unpack('>II', data[16:24])
        digest = hashlib.sha256(data).hexdigest()
        relative = str(p.relative_to(ROOT))
        signatures[digest].append(relative)
        dimensions[f'{w}x{h}'] += 1
        records.append(dict(path=relative, sha256=digest, bytes=len(data), width=w, height=h))
    families.append(dict(family=family.name, png_files=sum(map(len, signatures.values())),
                         unique_png_bytes=len(signatures), dimensions=dict(dimensions),
                         exact_duplicate_groups=[paths for paths in signatures.values() if len(paths)>1]))
report = dict(scope='PNG檔案內容去重；不同PNG編碼的同像素圖尚未合併。動畫幀不視作獨立畫作。',
              families=families, files=records, png_files=len(records),
              consumer_audit='下一步依bootstrap載入端確認共享圖塊及動畫段，不能由檔案數推算獨立重繪工量。',
              prototype='可丟棄HD操作介面；未接正式規則、存檔或renderer。')
if not args.refresh_preview:
    (OUT/'INVENTORY.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
map_path = ROOT/'dist-all/v.1.0.0-20261010/smoke/gui-0/12-command-x11.png'
assert struct.unpack('>II',map_path.read_bytes()[16:24]) == (960,720), '試樣星圖視窗需配合同一張來源'
url = 'data:image/png;base64,'+base64.b64encode(map_path.read_bytes()).decode()
template = ROOT/'tools/scripts/hd_interface_preview.html'
html = template.read_text().replace('__MAP__', url)
(OUT/'HD-INTERFACE.html').write_text(html)
(OUT/'SOURCES.json').write_text(json.dumps({'template_sha256':hashlib.sha256(template.read_bytes()).hexdigest(),
 'map_source':str(map_path.relative_to(ROOT)), 'map_sha256':hashlib.sha256(map_path.read_bytes()).hexdigest(),
 'map_viewport':[8,69,462,437], 'map_method':'HTML內SVG視窗只顯示星圖區，原始PNG保持',
 'rights':'試樣含原版衍生畫面，只留本機。字型使用瀏覽器可用字型，不打包字型檔。'},ensure_ascii=False,indent=2)+'\n')
print(json.dumps([dict(family=x['family'], files=x['png_files'], unique=x['unique_png_bytes']) for x in families],ensure_ascii=False))
