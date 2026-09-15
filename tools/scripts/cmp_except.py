"""比兩張 640×480 的 screen.bin，排除一個矩形（x,y,w,h）之後要逐位元組相同。
相同 exit 0，不同 exit 1 並印差異數。"""
import sys, pathlib
W = 640
a = pathlib.Path(sys.argv[1]).read_bytes()
b = pathlib.Path(sys.argv[2]).read_bytes()
x, y, w, h = (int(v) for v in sys.argv[3].split(','))
if len(a) != len(b):
    print('長度不同', len(a), len(b)); sys.exit(1)
bad = 0
for i in range(len(a)):
    if a[i] != b[i]:
        px, py = i % W, i // W
        if x <= px < x + w and y <= py < y + h:
            continue
        bad += 1
if bad:
    print('排除矩形之外仍有', bad, '個像素不同'); sys.exit(1)
sys.exit(0)
