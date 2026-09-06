"""列出矩形內指定色號的連通分量 bbox（8 連通）。
用法：comp_rects.py <screen.bin> x0,y0,x1,y1 <色號,...> [最小面積]"""
import sys, pathlib
W = 640
scr = pathlib.Path(sys.argv[1]).read_bytes()
x0, y0, x1, y1 = (int(v) for v in sys.argv[2].split(','))
cols = set(int(v) for v in sys.argv[3].split(','))
minarea = int(sys.argv[4]) if len(sys.argv) > 4 else 20
pts = {(x, y) for y in range(y0, y1) for x in range(x0, x1) if scr[y*W+x] in cols}
seen, out = set(), []
for p in sorted(pts):
    if p in seen: continue
    st, comp = [p], []
    seen.add(p)
    while st:
        x, y = st.pop(); comp.append((x, y))
        for n in ((x-1,y),(x+1,y),(x,y-1),(x,y+1),(x-1,y-1),(x+1,y-1),(x-1,y+1),(x+1,y+1)):
            if n in pts and n not in seen:
                seen.add(n); st.append(n)
    xs=[a for a,_ in comp]; ys=[b for _,b in comp]
    if len(comp) >= minarea:
        out.append((min(xs),min(ys),max(xs),max(ys),len(comp)))
for r in sorted(out, key=lambda r:(r[1],r[0])):
    print('x %3d..%3d  y %3d..%3d  (%d×%d) 像素 %d' %
          (r[0],r[2],r[1],r[3],r[2]-r[0]+1,r[3]-r[1]+1,r[4]))
