import sys
sys.path.insert(0, '/work/tools/scripts')
import screenspec as S
a, _ = S.load_screen('/work/workplace/oracle/s3data/screen.bin')
b, _ = S.load_screen('/work/workplace/oracle/s3data0/screen.bin')
# 被蓋掉的整塊：逐列／逐行找差異的極值
cols = [x for x in range(S.W) if any(a[y][x]!=b[y][x] for y in range(60,420)) and x < 463]
print('左側被蓋掉的欄 x %d–%d（連續? %s）' % (min(cols), max(cols), len(cols)==max(cols)-min(cols)+1))
rws = [y for y in range(S.H) if any(a[y][x]!=b[y][x] for x in range(0,463))]
print('左側被蓋掉的列 y %d–%d（連續? %s）' % (min(rws), max(rws), len(rws)==max(rws)-min(rws)+1))

print('\n整塊往內剝（sample 取表頭那一列）：')
lay, alt = S.measure_frame(a, 16, 60, 462, 419, fill=9, sample=(300, 110), layers_max=12)
for i,(l,t) in enumerate(zip(lay,alt)):
    print('  第%2d 層 上%2d 右%2d 下%2d 左%2d  alt %s' % (i, l[0],l[1],l[2],l[3], t))

print('\n上緣 x=300 的逐列色（y 56–72）：', ''.join('%X'%a[y][300] for y in range(56,73)))
print('左緣 y=110 的逐行色（x 12–40）：', ''.join('%X'%a[110][x] for x in range(12,41)))
print('右緣 y=110 的逐行色（x 448–470）：', ''.join('%X'%a[110][x] for x in range(448,471)))
print('下緣 x=300 的逐列色（y 404–424）：', ''.join('%X'%a[y][300] for y in range(404,425)))
