# -*- coding: utf-8 -*-
"""把 165F:004D 的 29 個點算出來，看收據上那些位置是什麼色。"""
import math, collections
W=640
scr=open('workplace/oracle/s3cmdrange/screen.bin','rb').read()
CX,CY=298,409
TURN,DIST=15,240
SIN=[round(math.sin(i/256*2*math.pi)*1024) for i in range(256)]

def pts(facing, mul=1):
    out=[]
    i=0
    while (i*16)//TURN <= DIST:
        sx=sy=0; a=0; rem=i
        while rem>=TURN:
            h=(a+facing)&0xFF
            sx+=SIN[(h+0x40)&0xFF]<<3; sy+=SIN[h]<<3
            a+=TURN; rem-=TURN
        frac=(rem<<3)//TURN
        h=(i+facing)&0xFF
        sx+=frac*SIN[(h+0x40)&0xFF]; sy+=frac*SIN[h]
        st=DIST-(i*16)//TURN
        sx+=st*SIN[(h+0x40)&0xFF]; sy+=st*SIN[h]
        out.append((CX+(sx*mul>>10), CY+(sy*mul>>10)))
        i+=8
    return out

for facing in (0, 190):
    p=pts(facing)
    c=collections.Counter()
    for x,y in p:
        c[scr[y*W+x] if 0<=x<W and 40<=y<400 else -1]+=1
    print('朝向 %3d：%d 點，落點色號分佈 %s' % (facing, len(p), sorted(c.items())))
    print('        前六個點：%s' % p[:6])
