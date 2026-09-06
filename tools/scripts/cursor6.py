import json
W=640
d=open('/work/workplace/oracle/curfill/screen.bin','rb').read()
# 滑鼠 (310,241) → 游標 (311,282)，那一帶是純視窗填色 9
x0,y0=311,282
pts=[(x,y) for y in range(y0-2,y0+18) for x in range(x0-2,x0+14) if d[y*W+x]!=9]
if not pts: print("該區全是填色，游標沒畫在那"); raise SystemExit
xs=[p[0] for p in pts]; ys=[p[1] for p in pts]
X0,X1,Y0,Y1=min(xs),max(xs),min(ys),max(ys)
print(f"游標 bbox ({X0},{Y0})-({X1},{Y1})  {X1-X0+1}×{Y1-Y0+1}  相對滑鼠 ({X0-310},{Y0-241})")
rows=[]
for y in range(Y0,Y1+1):
    r=''.join({7:'#',3:'o'}.get(d[y*W+x],'.') for x in range(X0,X1+1))
    rows.append(r); print('  '+r)
json.dump({'width':X1-X0+1,'height':Y1-Y0+1,
           'offset_from_mouse':[X0-310,Y0-241],
           'colors':{'#':7,'o':3},'transparent':'.',
           'note':'遊戲自己畫的箭頭游標：白色本體 ＋ 色號 3 描邊。在黑色背景上描邊看不見，所以只在純填色區才量得到完整形狀。',
           'rows':rows}, open('/work/docs/internals/data/cursor.json','w'), ensure_ascii=False, indent=2)
print("寫出 cursor.json")
