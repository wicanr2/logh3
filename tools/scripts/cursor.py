import json
W=640
o=open('/work/workplace/oracle/r0/screen.bin','rb').read()
r=open('/work/workplace/remake/screen.bin','rb').read()
d=[(i%W,i//W,o[i],r[i]) for i in range(W*480) if o[i]!=r[i]]
xs=[x for x,_,_,_ in d]; ys=[y for _,y,_,_ in d]
x0,x1,y0,y1=min(xs),max(xs),min(ys),max(ys)
print(f"游標 bbox ({x0},{y0})-({x1},{y1})  {x1-x0+1}×{y1-y0+1}  {len(d)} 像素")
print("色號:", sorted({c for _,_,c,_ in d}))
print("點陣（相對 bbox）:")
grid=[]
for y in range(y0,y1+1):
    row=''.join('#' if (x,y) in {(a,b) for a,b,_,_ in d} else '.' for x in range(x0,x1+1))
    grid.append(row); print('  '+row)
json.dump({'w':x1-x0+1,'h':y1-y0+1,'offset_from_mouse':[x0-300,y0-170],
           'color':sorted({c for _,_,c,_ in d}),'rows':grid},
          open('/work/docs/internals/data/cursor.json','w'), ensure_ascii=False, indent=2)
