W=640
o=open('/work/workplace/oracle/r0/screen.bin','rb').read()
sym={9:'.',7:'#',3:'o',2:'R',15:'+',8:'*'}
print("oracle r0，(296..320, 216..244)：")
for y in range(216,245):
    print(f"  y{y} " + ''.join(sym.get(o[y*W+x],'?') for x in range(296,321)))
