"""印出一份收據上的色 7 長線段（框），與逐列的文字。"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
name=sys.argv[1]
x0,y0,x1,y1=(int(v) for v in sys.argv[2:6])
inks=[int(v) for v in sys.argv[6].split(',')] if len(sys.argv)>6 else [7]
rows,_=S.load_screen('/work/workplace/oracle/%s/screen.bin'%name)
tabs={b:S.font_lookup('/orig/GRAPH.IMG',bold=b) for b in (False,True)}
def big5(c):
    try: return bytes([c>>8,c&0xFF]).decode('big5')
    except Exception: return '?%04X'%c
print('== 色 7 水平線段（長 ≥ 60）')
for y in range(y0,y1+1):
    x=x0
    while x<=x1:
        if rows[y][x]==7:
            s=x
            while x<=x1 and rows[y][x]==7: x+=1
            if x-s>=60: print('  y=%3d x %3d–%3d 長 %d'%(y,s,x-1,x-s))
        else: x+=1
print('== 色 7 垂直線段（長 ≥ 60）')
for x in range(x0,x1+1):
    y=y0
    while y<=y1:
        if rows[y][x]==7:
            s=y
            while y<=y1 and rows[y][x]==7: y+=1
            if y-s>=60: print('  x=%3d y %3d–%3d 長 %d'%(x,s,y-1,y-s))
        else: y+=1
print('== 文字')
for y in range(y0,y1-14):
    for ink in inks:
        out,x=[],x0
        while x+8<=x1:
            hit=None
            for bold in (False,True):
                full,half=tabs[bold]
                b=S.cell_bits(rows,x,y,ink,16)
                if any(b) and b in full: hit=(big5(full[b]),16); break
            if not hit:
                for bold in (False,True):
                    full,half=tabs[bold]
                    b=S.cell_bits(rows,x,y,ink,8)
                    if any(b) and b in half: hit=(chr(half[b]),8); break
            if hit: out.append((x,hit[0])); x+=hit[1]
            else: x+=8
        good=[o for o in out if o[1] not in '─│┌┐└┘▔▁▏▕╴']
        if len(good)>=2:
            print('  y=%3d 色%d %s'%(y,ink,' '.join('%d:%s'%g for g in good)))
