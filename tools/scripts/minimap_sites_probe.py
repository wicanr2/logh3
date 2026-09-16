import sys
def decode(d):
    out=[]; i=0
    while i < len(d):
        if d[i] >= 0x80: out.append(d[i]&0x7F); i+=1; continue
        if i+1 >= len(d): out.append(d[i]); i+=1; continue
        out += [d[i]]*d[i+1]; i+=2
    return out
region=int(sys.argv[1]); a=open(sys.argv[2],'rb').read()
t=decode(open(f'/orig/MAP{region:02d}.CMP','rb').read())
sites=[(i%128, i//128) for i,c in enumerate(t) if c==6]
for k,(col,row) in enumerate(sites):
    X=col*16+8; Y=row*20+12
    px=152+(X>>4); py=220+(Y>>4)
    print(k, 'col,row', col,row, 'px,py', px,py, 'pixel', a[py*640+px], 'shadow', a[(py-1)*640+px-1], 'button y', 230+20*k)
