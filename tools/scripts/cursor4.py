import json
W=640
a=open('/work/workplace/oracle/r0/screen.bin','rb').read()   # 滑鼠 x=300
b=open('/work/workplace/oracle/r0c/screen.bin','rb').read()  # 滑鼠 x=340
# 游標在 a 的位置：x 從 300 起；在 b 的位置：x 從 340 起。互不重疊。
sym={9:'.',7:'#',3:'o',2:'R',15:'+',8:'*'}
print("a（滑鼠 x=300）x 296..320:")
for y in range(218,240):
    print(f"  y{y} " + ''.join(sym.get(a[y*W+x],'?') for x in range(296,321)))
print("\nb（滑鼠 x=340）x 336..360（該處原本是視窗外／填色）:")
for y in range(218,240):
    print(f"  y{y} " + ''.join(sym.get(b[y*W+x],'?') for x in range(336,361)))
