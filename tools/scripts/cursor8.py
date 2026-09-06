import json
W=640
d=open('/work/workplace/oracle/curbar/screen.bin','rb').read()
# 反白條 row0：x 288..335, y 200..219，純色號 2；游標應在 (300,202) 一帶
pts=[(x,y) for y in range(200,220) for x in range(288,336) if d[y*W+x]!=2]
if not pts: print("反白條上沒有游標"); raise SystemExit
xs=[p[0] for p in pts]; ys=[p[1] for p in pts]
X0,X1,Y0,Y1=min(xs),max(xs),min(ys),max(ys)
print(f"游標 bbox ({X0},{Y0})-({X1},{Y1})  {X1-X0+1}×{Y1-Y0+1}  相對滑鼠 ({X0-300},{Y0-162})")
rows=[]
for y in range(Y0,Y1+1):
    r=''.join({7:'#',3:'o'}.get(d[y*W+x],'.') for x in range(X0,X1+1))
    rows.append(r); print('  '+r)
json.dump({'width':X1-X0+1,'height':Y1-Y0+1,
           'offset_from_mouse':[X0-300,Y0-162],
           'colors':{'#':7,'o':3},'transparent':'.',
           'note':'量自反白條（純色號 2）上的游標：那是畫面上唯一夠大又單色的區域，白色本體與色號 3 的描邊在上面都看得見。',
           'rows':rows}, open('/work/docs/internals/data/cursor.json','w'), ensure_ascii=False, indent=2)
print("寫出 cursor.json")
