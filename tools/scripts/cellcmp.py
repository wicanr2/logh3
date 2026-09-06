W=640
a=open('/work/workplace/oracle/hov170/screen.bin','rb').read()
b=open('/work/workplace/remake/screen.bin','rb').read()
sym={9:'.',7:'#',3:'o',2:'R',15:'+',8:'*'}
for name,d in (('oracle',a),('remake',b)):
    print(f"\n{name} (318..337, 236..262)")
    for y in range(236,262):
        print(f"  y{y} " + ''.join(sym.get(d[y*W+x],'?') for x in range(318,338)))
