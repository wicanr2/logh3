"""用 MAP<nn>.CMP 算「指定待命行星」那張 128×128 小圖（docs/re/314），與收據比。
用法：minimap_probe.py 星域 收據/screen.bin [x y]（畫面左上角，預設 152 220）"""
import sys, pathlib
def decode(d):
    out=[]; i=0
    while i < len(d):
        if d[i] >= 0x80: out.append(d[i]&0x7F); i+=1; continue
        if i+1 >= len(d): out.append(d[i]); i+=1; continue
        out += [d[i]]*d[i+1]; i+=2
    return out
region=int(sys.argv[1]); scr=pathlib.Path(sys.argv[2]).read_bytes()
X=int(sys.argv[3]) if len(sys.argv)>3 else 152; Y=int(sys.argv[4]) if len(sys.argv)>4 else 220
t=decode(open(f'/orig/MAP{region:02d}.CMP','rb').read())
def bit(r,c):
    v=t[r*128+c]
    if r%12==0 and v==0: v=3
    if c%15==0 and v==0: v=4
    if c%15==0 and v==3: v=2
    return 0 if v in (0,1,5) else 1
bad=0; rows_bad=[]
for py in range(128):
    s=py-py//5
    if s>101: continue
    nb=0
    for px in range(128):
        want = 1 if bit(s,px) else 3
        got = scr[(Y+py)*640+X+px]
        if want!=got: nb+=1
    if nb: rows_bad.append((py,nb)); bad+=nb
print('差', bad, '列', rows_bad[:20])
