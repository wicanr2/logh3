import collections
W, H = 640, 480
a = open('/work/workplace/re06/screen.bin','rb').read()
b = open('/work/workplace/remake/screen.bin','rb').read()
d = [(i%W, i//W, a[i], b[i]) for i in range(W*H) if a[i] != b[i]]
print("不同像素", len(d))
rows = collections.Counter(y for _,y,_,_ in d)
cols = collections.Counter(x for x,_,_,_ in d)
print("差異最多的列:", rows.most_common(6))
print("差異最多的行:", cols.most_common(6))
X0, Y0 = 273, 181
print("\noracle / remake（視窗左上 12×10，o=oracle r=remake）")
for y in range(Y0, Y0+10):
    o = ' '.join(f"{a[y*W+x]:2d}" for x in range(X0, X0+12))
    r = ' '.join(f"{b[y*W+x]:2d}" for x in range(X0, X0+12))
    mark = '   <<' if o != r else ''
    print(f"  y{y-Y0:3d} o|{o}|\n       r|{r}|{mark}")


print("\n逐點差異：")
for x,y,ao,bo in d:
    print(f"  ({x},{y}) 相對視窗({x-273},{y-181})  oracle={ao} remake={bo}")
