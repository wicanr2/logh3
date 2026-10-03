"""089：移除舊月初清單量測混入框線的25個游標像素，保留動態游標。"""
import hashlib
import json
from pathlib import Path


def strip_monthly_cursor(spec):
    # 舊量測的mouse(464,330)在畫布(464,370)；此處位於兩列框線內部。
    # 新原版保持雙跑重畫框線後沒有這25個像素，研究396／089已審查。
    if (spec.get('mouse', {}).get('x'), spec.get('mouse', {}).get('y')) != (464, 330):
        return []
    removed = []
    for box in spec['windows']['overlay']['boxes']:
        if (box['x'], box['y'], box['w'], box['h']) not in [
                (336, 360, 280, 20), (336, 380, 280, 20)]:
            continue
        kept = []
        for patch in box.get('patches', []):
            x, y = box['x'] + patch['dx'], box['y'] + patch['dy']
            if 464 <= x < 480 and 370 <= y < 386:
                removed.append({'x': x, 'y': y, 'color': patch['color']})
            else:
                kept.append(patch)
        box['patches'] = kept
    return removed


if __name__ == '__main__':
    root = Path('/src')
    path = root / 'docs/internals/data/screen-admiral-list.json'
    spec = json.loads(path.read_text())
    before = hashlib.sha256(path.read_bytes()).hexdigest()
    raw_a = (root / 'workplace/oracle/hold60-r58-month-clean-a/screen.bin').read_bytes()
    raw_b = (root / 'workplace/oracle/hold60-r58-month-clean-b/screen.bin').read_bytes()
    assert raw_a == raw_b
    assert hashlib.sha256(raw_a).hexdigest() == '45c79a94d9a81231198081a05041c81e3bb22a65dddbd31202729e81882a532f'
    removed = strip_monthly_cursor(spec)
    assert len(removed) == 25, '拒絕未識別輸入或重複套用'
    for patch in removed:
        assert raw_a[patch['y'] * 640 + patch['x']] != patch['color']
    path.write_text(json.dumps(spec, ensure_ascii=False, indent=1) + '\n')
    print(json.dumps({'source_before_sha256': before, 'removed': removed,
                      'source_after_sha256': hashlib.sha256(path.read_bytes()).hexdigest()},
                     ensure_ascii=False))
