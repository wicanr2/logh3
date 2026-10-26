"""150：盤點可見介面文字來源及Big5字碼，不改原始資料。Docker內執行。"""
import ast
import json
from pathlib import Path
import re

ROOT = Path('/src')
OUT = ROOT/'workplace/poster150'
items = {}


def add(s, source):
    if isinstance(s, str) and re.search('[\u3400-\u9fff]', s):
        items.setdefault(s, set()).add(source)


def walk(v, source, key=''):
    if isinstance(v, dict):
        for k, value in v.items(): walk(value, source, k)
    elif isinstance(v, list):
        for value in v: walk(value, source, key)
    elif key in ['text', 'label', 'title', 'name', 'caption', 'message', 'lines']:
        add(v, source)


for f in (ROOT/'docs/internals/data').glob('screen-*.json'):
    walk(json.loads(f.read_text()), str(f.relative_to(ROOT)))
for folder in ['cmd/gin3', 'internal/screen', 'internal/game']:
    for f in (ROOT/folder).glob('*.go'):
        if f.name.endswith('_test.go'): continue
        for m in re.finditer(r'"(?:[^"\\\n]|\\.)*"', f.read_text()):
            try: value = json.loads(m[0])
            except (ValueError, TypeError): continue
            if len(value) < 150: add(value, str(f.relative_to(ROOT)))
rows = [dict(text=k, sources=sorted(v)) for k, v in sorted(items.items())]
(OUT/'TEXT-INVENTORY.json').write_text(json.dumps(rows, ensure_ascii=False, indent=2)+'\n')
print('文字項數', len(rows))
print('\n'.join(str(i)+' '+r['text'] for i, r in enumerate(rows)))
