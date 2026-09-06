W=640
a=open('/work/workplace/oracle/r0/screen.bin','rb').read()
b=open('/work/workplace/remake/screen.bin','rb').read()
sym={9:'.',7:'#',3:'o',2:'R',15:'+',8:'*'}
for n,d in (('oracle',a),('remake',b)):
    print(f"{n} x296..316 y216..232")
    for y in range(216,233):
        print(f"  y{y} " + ''.join(sym.get(d[y*W+x],'?') for x in range(296,317)))
