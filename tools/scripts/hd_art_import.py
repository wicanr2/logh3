"""容器內將 image_gen 原始 PNG 複製進 HD 素材庫；不修改影像。"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import struct
import sys

root = Path('/work')
plan = json.loads((root/sys.argv[1]).read_text())
manifest_path = root/'assets/hd/MANIFEST.json'
manifest = json.loads(manifest_path.read_text())
records = {r['index']: r for r in manifest['files']}
names = {r['index']: r['name'] for r in json.loads((root/'docs/internals/data/keireki.json').read_text())['records']}

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

assert (root/'assets/hd/face').stat().st_uid == os.getuid()
for item in plan:
    index = item['index']
    assert 0 <= index < 82
    source = Path('/ref')/Path(item['source']).name
    assert source.is_file()
    header = source.read_bytes()[:24]
    assert header[:8] == b'\x89PNG\r\n\x1a\n'
    width, height = struct.unpack('>II', header[16:24])
    assert 240 <= width <= 4096 and 300 <= height <= 4096
    assert abs(width*100-height*80)*100 <= max(width*100,height*80), (index,width,height)
    name = f'face-{index:02d}'
    destination = root/'assets/hd/face'/f'{name}.png'
    prompt_path = destination.with_suffix('.prompt.txt')
    if destination.exists():
        assert sha(destination) == sha(source), '拒絕覆寫不同的定版圖'
    else:
        shutil.copyfile(source, destination)
        prompt_path.write_text(item['prompt']+'\n')
    assert destination.stat().st_uid == os.getuid()
    records[index] = dict(index=index,name=names[index],file=f'face/{name}.png',width=width,height=height,
        bytes=destination.stat().st_size,sha256=sha(destination),
        original_sha256=sha(root/'assets/original/face'/f'{name}.png'),
        tool='built-in image_gen',prompt=f'face/{name}.prompt.txt')
manifest['files'] = [records[i] for i in sorted(records)]
manifest['completed_portraits'] = len(records)
manifest['original_people'] = 82
manifest['blank_index'] = 82
manifest['status'] = 'portrait restoration in progress; blank index 82 preserved without invented person'
manifest_path.write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print('已保存',len(records),'張人物；PNG原檔與提示詞、來源SHA-256已核對')
