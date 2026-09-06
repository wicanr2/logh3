W,H=640,480
import glob,os,collections
base=open('/work/workplace/re06/screen.bin','rb').read()
for f in sorted(glob.glob('/work/workplace/re08/y*.bin')):
    d=open(f,'rb').read()
    diff=[(i%W,i//W,base[i],d[i]) for i in range(W*H) if base[i]!=d[i]]
    if not diff:
        print(os.path.basename(f),"無變化"); continue
    ys=sorted({y for _,y,_,_ in diff}); xs=sorted({x for x,_,_,_ in diff})
    cols=collections.Counter(n for _,_,_,n in diff)
    print(f"{os.path.basename(f)}: y {ys[0]}–{ys[-1]}  x {xs[0]}–{xs[-1]}  新色號 {cols.most_common(4)}")
