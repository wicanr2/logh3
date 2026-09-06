# 把 workplace/explore 裡所有建造計劃 dump 的（船艦數, 部隊）疊起來，
# 反推每一列的「部隊定員」區間（docs/re/11）。
import json, sys, glob, os
W=640; GH=15; HW=8
spec=json.load(open('/work/docs/internals/data/screen-build-plan.json'))
INK=spec['windows']['report']['ink']
fnt=open('/orig/GRAPH.IMG','rb').read()
pats={}
for ch in "0123456789%":
    o=ord(ch)*GH
    pats[tuple(fnt[o:o+GH])]=ch
def read(img,x,y,w):
    out=""
    for cx in range(x,x+w,HW):
        rows=[]
        for gy in range(GH):
            v=0
            for gx in range(HW):
                if img[(y+gy)*W+cx+gx]==INK: v|=1<<(HW-1-gx)
            rows.append(v)
        out+=pats.get(tuple(rows),"?" if any(rows) else " ")
    return out
def num(s):
    s=s.strip()
    return int(s) if s.isdigit() else None
ROWY=[202+20*i for i in range(7)]
obs=[]
for f in sorted(glob.glob('/work/workplace/explore/*.bin'))+['/work/workplace/oracle/plan1/screen.bin']:
    img=open(f,'rb').read()
    if len(img)!=307200: continue
    # 只要建造計劃畫面：預算那一格要讀得出數字
    if num(read(img,280,162,40)) is None: continue
    for i,y in enumerate(ROWY):
        ships=num(read(img,336,y,48))
        units=num(read(img,392,y,48))
        if ships is None or units is None: continue
        obs.append((i,ships,units,os.path.basename(f)))
lo=[0]*7; hi=[10**9]*7
for i,ships,units,f in obs:
    # units = floor(ships / cap) → cap 的區間
    if units>0:
        hi[i]=min(hi[i], ships//units)
        lo[i]=max(lo[i], ships//(units+1))
    else:
        lo[i]=max(lo[i], ships)
for i in range(7):
    print("列%d 定員 ∈ (%d, %d]" % (i, lo[i], hi[i]))
print("觀測", len(obs), "筆")
