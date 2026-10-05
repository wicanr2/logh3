"""第104輪全圖RGB統計。只在具有ImageMagick的Wine Docker image內執行。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

assert Path.cwd() == Path('/out') and os.getuid() == 1000
parser = argparse.ArgumentParser()
parser.add_argument('--output', default='player69-r104-PIXELS4.json')
args = parser.parse_args()
assert Path(args.output).name == args.output and args.output.endswith('.json')
target = Path(args.output)
assert not target.exists()


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def compare(left, right, scale_left=False, logical=None):
    width, height = map(int, subprocess.check_output(
        ['identify', '-format', '%w %h', str(right)], text=True).split())
    command = ['convert', str(left)]
    if scale_left:
        command += ['-filter', 'point', '-resize', f'{width}x{height}!']
    a = subprocess.check_output(command + ['-depth', '8', 'RGB:-'])
    b = subprocess.check_output(['convert', str(right), '-depth', '8', 'RGB:-'])
    assert len(a) == len(b) == width * height * 3
    count, maximum = 0, 0
    xmin, ymin, xmax, ymax = width, height, -1, -1
    for pos in range(0, len(a), 3):
        av, bv = a[pos:pos+3], b[pos:pos+3]
        if av == bv:
            continue
        count += 1
        x, y = (pos // 3) % width, (pos // 3) // width
        xmin, ymin, xmax, ymax = min(xmin, x), min(ymin, y), max(xmax, x), max(ymax, y)
        maximum = max(maximum, *(abs(av[n] - bv[n]) for n in range(3)))
    result = {'left': str(left), 'left_sha256': sha(left), 'right': str(right), 'right_sha256': sha(right),
            'canvas': [width, height], 'pixels': width * height, 'same': width * height - count,
            'different': count, 'max_channel_difference': maximum,
            'bbox': [xmin, ymin, xmax-xmin+1, ymax-ymin+1] if count else None,
            'transform': '完整對照圖依point濾鏡放大；不裁切' if scale_left else '原尺寸；不裁切',
            'scope': '同來源remake跨平台渲染；非新增原版同態'}
    if logical is not None:
        assert (width, height) == (960, 720)
        source = subprocess.check_output(['convert', str(logical), '-depth', '8', 'RGB:-'])
        assert len(source) == 640 * 480 * 3
        outside, outside_ties = [0, 0], 0
        for y in range(height):
            ys = [2*(y//3), 2*(y//3)+1] if y % 3 == 1 else [(2*y+1)//3]
            for x in range(width):
                xs = [2*(x//3), 2*(x//3)+1] if x % 3 == 1 else [(2*x+1)//3]
                candidates = [source[(sy*640+sx)*3:(sy*640+sx)*3+3] for sy in ys for sx in xs]
                # 固定Ebitengine ScreenShaderSource方框濾鏡：1.5倍邊界等權混合。
                # 分別核每個通道的下／上整數界，保留跨平台原始差異數。
                bounds = [(sum(c[n] for c in candidates)//len(candidates),
                           (sum(c[n] for c in candidates)+len(candidates)-1)//len(candidates))
                          for n in range(3)]
                pos = (y*width+x)*3
                outside[0] += any(not bounds[n][0] <= a[pos+n] <= bounds[n][1] for n in range(3))
                outside[1] += any(not bounds[n][0] <= b[pos+n] <= bounds[n][1] for n in range(3))
                outside_ties += a[pos:pos+3] != b[pos:pos+3] and x % 3 != 1 and y % 3 != 1
        result['sampling'] = {'logical': str(logical), 'logical_sha256': sha(logical),
                              'formula': '固定ScreenShaderSource方框濾鏡；p%3==1取相鄰兩點等權平均，其餘取單點；每通道核平均值下／上整數界',
                              'left_outside_box_rounding_bounds': outside[0],
                              'right_outside_box_rounding_bounds': outside[1],
                              'differences_outside_tie_rows_or_columns': outside_ties,
                              'full_window_exact': count == 0}
    return result


reports = []
base = Path('player69-r104-startup')
for name in ['startup', 'repeat']:
    root = Path('player69-r104-' + name)
    assert (root / 'title.bin').read_bytes() == (base / 'linux-title.bin').read_bytes()
    reports.append(compare(base / 'linux-title.png', root / 'title.png'))
    process = json.loads((root / 'PROCESS.json').read_text())
    reports.append(compare(root / 'title.png', root / process['screenshots'][-1]['path']))
for name in ['load1b', 'load15b', 'fleet1']:
    root = Path('player69-r104-' + name)
    check = json.loads((root / 'CHECK.json').read_text())
    chosen = {}
    for row in check['frames']:
        if row['exit'] == 0:
            chosen[row['name']] = row
    for stage, row in chosen.items():
        reference = Path('player69-r104-reference2') / (stage + '.png')
        reports.append(compare(reference, root / (row['path'] + '.png')))
        scaled_root = Path('player69-r104-reference15b')
        scaled = json.loads((scaled_root / 'REFERENCES.json').read_text())
        raw_reference = scaled_root / next(r['window'] for r in scaled['runs'] if r['name'] == stage) if name == 'load15b' else reference
        reports.append(compare(raw_reference, root / (row['path'] + '-window.png'),
                               logical=reference if name == 'load15b' else None))
result = {'schema': 'logh3.windows-player-pixels.v1', 'reports': reports,
          'total_pixels': sum(r['pixels'] for r in reports), 'different': sum(r['different'] for r in reports),
          'status': '一倍視窗及原生邏輯畫布一致；1.5倍完整實體像素差異另列，不稱全圖零差'}
with target.open('x') as stream:
    json.dump(result, stream, ensure_ascii=False, indent=2)
    stream.write('\n')
print(len(reports), 'full-image comparisons', result['total_pixels'], 'pixels', result['different'], 'different')
assert all(r['different'] == 0 for r in reports if 'sampling' not in r)
for r in reports:
    if 'sampling' in r:
        s = r['sampling']
        assert s['left_outside_box_rounding_bounds'] == s['right_outside_box_rounding_bounds'] == 0
        assert s['differences_outside_tie_rows_or_columns'] == 0
print('一倍與原生畫布零差；1.5倍差異限已逐像素核對的方框濾鏡量化邊界，不放寬全視窗一致門檻')
