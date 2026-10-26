"""155：只在workplace複製的凍結來源套用視覺草稿，不修改產品程式。"""
import hashlib
import json
import os
from pathlib import Path
import shutil

ROOT=Path('/src');OUT=ROOT/'workplace/poster155'
BASE=ROOT/'workplace/package92-v.1.6.2-20261025-release154/source'
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
assert os.getuid()==os.getgid()==1000 and not OUT.exists()
assert BASE.is_dir() and (ROOT/'workplace').stat().st_uid==1000
OUT.mkdir();shutil.copytree(BASE,OUT/'source')
target=OUT/'source/cmd/gin3/poster_decoration.go'
original=target.read_text();marker='func (d *posterDecoration) drawBand('
assert original.count(marker)==1
snippet=ROOT/'tools/scripts/poster_elegant_draw.go.txt'
target.write_text(original[:original.index(marker)]+snippet.read_text())
(OUT/'SOURCE.json').write_text(json.dumps(dict(classification='可丟棄視覺原型，不是正式封包',
    base=str(BASE.relative_to(ROOT)),base_file_sha256=sha(BASE/'cmd/gin3/poster_decoration.go'),
    draw_source=str(snippet.relative_to(ROOT)),draw_sha256=sha(snippet),
    production_sha256=sha(ROOT/'cmd/gin3/poster_decoration.go')),ensure_ascii=False,indent=2)+'\n')
print('READY 155隔離原型；正式Go與v1.6.2未改',flush=True)
