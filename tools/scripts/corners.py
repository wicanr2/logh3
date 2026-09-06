W, H = 640, 480
a = open('/work/workplace/re06/screen.bin','rb').read()
X0, Y0, WD, HT = 273, 181, 78, 118
def blk(name, xs, ys):
    print(f"\n{name}（列＝y，欄＝x）")
    print("      " + ' '.join(f"{x-X0:2d}" for x in xs))
    for y in ys:
        print(f"  y{y-Y0:3d} " + ' '.join(f"{a[y*W+x]:2d}" for x in xs))
blk("左上", range(X0, X0+8), range(Y0, Y0+8))
blk("右上", range(X0+WD-8, X0+WD), range(Y0, Y0+8))
blk("左下", range(X0, X0+8), range(Y0+HT-8, Y0+HT))
blk("右下", range(X0+WD-8, X0+WD), range(Y0+HT-8, Y0+HT))
