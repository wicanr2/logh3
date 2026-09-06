import json
W=640
d=open('/work/workplace/oracle/cursor/screen.bin','rb').read()
# 滑鼠 (450,300) → 游標畫在 (450,350)，那一帶背景是純色號 3
xs=[i%W for i in range(W*480) if d[i]!=3 and 440<=i%W<490 and 340<=i//W<390]
ys=[i//W for i in range(W*480) if d[i]!=3 and 440<=i%W<490 and 340<=i//W<390]
if not xs:
    print("那一帶沒有非背景像素"); raise SystemExit
x0,x1,y0,y1=min(xs),max(xs),min(ys),max(ys)
print(f"游標 bbox ({x0},{y0})-({x1},{y1})  {x1-x0+1}×{y1-y0+1}")
print(f"相對滑鼠 (450,300) 的位移：({x0-450},{y0-300})")
rows=[]
sym={0:'.',3:'.',7:'#',9:'o',15:'+',8:'*',2:'R'}
for y in range(y0,y1+1):
    r=''.join(sym.get(d[y*W+x],'?') for x in range(x0,x1+1))
    rows.append([d[y*W+x] for x in range(x0,x1+1)])
    print('  '+r)
json.dump({'w':x1-x0+1,'h':y1-y0+1,
           'offset_from_mouse':[x0-450,y0-300],
           'transparent':3,
           'pixels':rows}, open('/work/docs/internals/data/cursor.json','w'), indent=1)
print("寫出 docs/internals/data/cursor.json")
