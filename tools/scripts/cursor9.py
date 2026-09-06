import json
W=640
a=open('/work/workplace/oracle/curfill/screen.bin','rb').read()  # 滑鼠 (310,241)
b=open('/work/workplace/oracle/cur300/screen.bin','rb').read()   # 滑鼠 (450,300)
# 兩份都沒有反白，畫面其餘部分相同；差異只有兩個位置的游標。
pts=[(i%W,i//W) for i in range(W*480) if a[i]!=b[i] and 300<=i%W<340 and 270<=i//W<310]
xs=[p[0] for p in pts]; ys=[p[1] for p in pts]
X0,X1,Y0,Y1=min(xs),max(xs),min(ys),max(ys)
print(f"游標 bbox ({X0},{Y0})-({X1},{Y1})  {X1-X0+1}×{Y1-Y0+1}  相對滑鼠 ({X0-310},{Y0-241})")
S={p:True for p in pts}
rows=[]
for y in range(Y0,Y1+1):
    r=''.join({7:'#',3:'o'}.get(a[y*W+x],'?') if (x,y) in S else '.' for x in range(X0,X1+1))
    rows.append(r); print('  '+r)
bad=[c for r in rows for c in r if c=='?']
print("未知色號的像素:", len(bad))
json.dump({'width':X1-X0+1,'height':Y1-Y0+1,
           'offset_from_mouse':[X0-310,Y0-241],
           'colors':{'#':7,'o':3},'transparent':'.',
           'note':'用兩份「都沒有反白」的收據互相對消取得：畫面其餘部分相同，差異只有游標本身，所以連壓在外框上的像素都讀得到。',
           'rows':rows}, open('/work/docs/internals/data/cursor.json','w'), ensure_ascii=False, indent=2)
print("寫出 cursor.json")
