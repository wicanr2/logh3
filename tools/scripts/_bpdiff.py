import json,sys
W=640
base=open("/work/workplace/oracle/plan1/screen.bin","rb").read()
for name in ["s0","s1","s2","s3","t1","t2"]:
    try: d=open("/work/workplace/explore/%s.bin"%name,"rb").read()
    except FileNotFoundError: continue
    diff=[i for i in range(len(base)) if base[i]!=d[i]]
    if not diff:
        print(name,"與 plan1 相同"); continue
    xs=[i%W for i in diff]; ys=[i//W for i in diff]
    print(name,"差異",len(diff),"bbox x",min(xs),max(xs),"y",min(ys),max(ys))
    rows=sorted(set(ys))
    print("   列:",rows[:3],"…",rows[-3:] if len(rows)>3 else "")
