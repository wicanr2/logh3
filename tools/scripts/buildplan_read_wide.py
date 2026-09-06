# 建造計劃整列的數字：百分比、金額、船艦數、部隊、建造中部隊。
# 比 buildplan_read.py 寬，用來看右邊那幾欄。
import json, sys
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
for name in sys.argv[1:]:
    img=open(name,"rb").read()
    print(name.split('/')[-1])
    for y in [162]+[202+20*i for i in range(7)]+[342]:
        print("  y=%3d |%s|" % (y, read(img,152,y,400)))
