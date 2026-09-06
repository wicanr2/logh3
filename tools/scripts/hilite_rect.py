import collections, json
W,H=640,480
d=open('/work/workplace/oracle/hilite/screen.bin','rb').read()
xs=[i%W for i in range(W*H) if d[i]==2]; ys=[i//W for i in range(W*H) if d[i]==2]
print(f"色號 2 的 bbox：({min(xs)},{min(ys)})-({max(xs)},{max(ys)})  {max(xs)-min(xs)+1}×{max(ys)-min(ys)+1}  共 {len(xs)} 像素")
solid = all(d[y*W+x]==2 for y in range(min(ys),max(ys)+1) for x in range(min(xs),max(xs)+1))
print("是否實心：", solid)
# 相對視窗 (273,181) 與第一格 (288,202)
print(f"相對視窗左上：({min(xs)-273},{min(ys)-181})   相對第一格：({min(xs)-288},{min(ys)-202})")
