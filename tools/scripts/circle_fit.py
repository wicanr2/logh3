"""量畫面上某個色號的像素外接矩形與推出來的圓心／半徑。

    tools/py.sh tools/scripts/circle_fit.py <批次> <色號...> [--rect x0,y0,x1,y1]
"""
import pathlib
import sys

W = 640
args = sys.argv[1:]
rect = (16, 60, 456, 420)
if "--rect" in args:
    i = args.index("--rect")
    rect = tuple(int(v) for v in args[i + 1].split(","))
    del args[i:i + 2]
batch, inks = args[0], [int(v) for v in args[1:]]
scr = pathlib.Path("/work/workplace/oracle/%s/screen.bin" % batch).read_bytes()

for ink in inks:
    pts = [(x, y) for y in range(rect[1], rect[3]) for x in range(rect[0], rect[2])
           if scr[y * W + x] == ink]
    if not pts:
        print("色 %d：沒有像素" % ink)
        continue
    xs = [p[0] for p in pts]; ys = [p[1] for p in pts]
    x0, x1, y0, y1 = min(xs), max(xs), min(ys), max(ys)
    print("色 %2d：%5d 個像素  外接 (%d,%d)–(%d,%d)  寬 %d 高 %d  中心 (%.1f,%.1f)  半徑 %.1f／%.1f"
          % (ink, len(pts), x0, y0, x1, y1, x1 - x0 + 1, y1 - y0 + 1,
             (x0 + x1) / 2, (y0 + y1) / 2, (x1 - x0 + 1) / 2, (y1 - y0 + 1) / 2))
