"""列畫面 dump 裡非背景像素的連通塊外框，用來定位單位方塊。

    tools/py.sh tools/scripts/screen_blobs.py <批次> [x0 y0 x1 y1]

輸入是 `workplace/oracle/<批次>/screen.bin`（640×480 的色號陣列）。
"""
import pathlib
import sys

W = 640
batch = sys.argv[1]
x0, y0, x1, y1 = (int(v) for v in (sys.argv[2:6] or ["0", "0", "640", "480"]))
IGNORE = {int(v) for v in (sys.argv[6].split(",") if len(sys.argv) > 6 else ["0"])}
scr = pathlib.Path("/work/workplace/oracle/%s/screen.bin" % batch).read_bytes()

seen = set()
blobs = []
for y in range(y0, y1):
    for x in range(x0, x1):
        if scr[y * W + x] in IGNORE or (x, y) in seen:
            continue
        stack, pts = [(x, y)], []
        seen.add((x, y))
        while stack:
            cx, cy = stack.pop()
            pts.append((cx, cy))
            for nx, ny in ((cx+1, cy), (cx-1, cy), (cx, cy+1), (cx, cy-1)):
                if x0 <= nx < x1 and y0 <= ny < y1 and (nx, ny) not in seen \
                        and scr[ny * W + nx] not in IGNORE:
                    seen.add((nx, ny))
                    stack.append((nx, ny))
        xs = [p[0] for p in pts]; ys = [p[1] for p in pts]
        blobs.append((min(xs), min(ys), max(xs) - min(xs) + 1,
                      max(ys) - min(ys) + 1, len(pts)))

for b in sorted(blobs, key=lambda b: (b[1], b[0])):
    if b[4] < 20:
        continue
    print("(%3d,%3d) %2d×%2d  %d 個像素" % b)
