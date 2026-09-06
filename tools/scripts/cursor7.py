import json
W=640
white = """#..........
###........
#####......
#######....
#########..
###########
######.....
#######....
###.###....
##...###...
#.....###..
......###..
.......###.
........#..""".split('\n')
h, w = len(white), max(len(r) for r in white)
solid = {(x,y) for y,r in enumerate(white) for x,c in enumerate(r) if c=='#'}
# 描邊 ＝ 白色像素的四鄰／八鄰中不是白色的位置
outline = set()
for (x,y) in solid:
    for dx in (-1,0,1):
        for dy in (-1,0,1):
            p=(x+dx,y+dy)
            if p not in solid:
                outline.add(p)
xs=[p[0] for p in solid|outline]; ys=[p[1] for p in solid|outline]
X0,Y0=min(xs),min(ys)
H,Wd=max(ys)-Y0+1, max(xs)-X0+1
rows=[]
for y in range(Y0,Y0+H):
    rows.append(''.join('#' if (x,y) in solid else ('o' if (x,y) in outline else '.') for x in range(X0,X0+Wd)))
print(f"推出的游標 {Wd}×{H}（白色本體相對它的位移 {(-X0,-Y0)}）")
for r in rows: print('  '+r)

# 用 curfill 驗證：滑鼠 (310,241)，白色本體左上角在 (311,282)
d=open('/work/workplace/oracle/curfill/screen.bin','rb').read()
bad=0; checked=0
for y,r in enumerate(rows):
    for x,c in enumerate(r):
        px,py = 311+X0+x, 282+Y0+y
        if py>291: continue          # 291 以下是視窗下緣框，不算
        got=d[py*W+px]; checked+=1
        want = 7 if c=='#' else (3 if c=='o' else 9)
        if got!=want: bad+=1
print(f"對照 curfill：檢查 {checked} 個像素，不符 {bad} 個")
json.dump({'width':Wd,'height':H,
           'offset_from_mouse':[1+X0,41+Y0],
           'colors':{'#':7,'o':3},'transparent':'.',
           'note':'白色本體量自純黑背景（cur300），描邊由本體的八鄰推出並用 curfill 驗證。',
           'rows':rows}, open('/work/docs/internals/data/cursor.json','w'), ensure_ascii=False, indent=2)
