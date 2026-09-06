W=640
a=open('/work/workplace/oracle/r0/screen.bin','rb').read()   # 滑鼠 y=170
b=open('/work/workplace/oracle/r0b/screen.bin','rb').read()  # 滑鼠 y=178
d=[(i%W,i//W) for i in range(W*480) if a[i]!=b[i]]
print("差異像素", len(d))
if d:
    xs=[x for x,_ in d]; ys=[y for _,y in d]
    print(f"bbox ({min(xs)},{min(ys)})-({max(xs)},{max(ys)})")
    sym={9:'.',7:'#',3:'o',2:'R',15:'+',8:'*'}
    for name,src in (('y=170',a),('y=178',b)):
        print(f"\n{name}（x {min(xs)}..{max(xs)}）")
        for y in range(min(ys),max(ys)+1):
            print(f"  y{y} " + ''.join(sym.get(src[y*W+x],'?') for x in range(min(xs),max(xs)+1)))
