import sys, os
sys.path.insert(0, '/work/tools/scripts')
import png
img = open('/orig/GRAPH.IMG', 'rb').read()
head = img[:430*30]
print(f"前段 {len(head)} bytes")
for ref in ('/ref/ascfont.15', '/ref/usrfont.15m'):
    if os.path.exists(ref):
        r = open(ref, 'rb').read()
        n = min(len(r), len(head))
        same = sum(1 for i in range(n) if r[i] == head[i])
        print(f"  與 {os.path.basename(ref)}（{len(r)} bytes）前 {n} bytes 相同率 {same*100.0/n:.1f}%")
        o = head.find(r[:60])
        print(f"  {os.path.basename(ref)} 的前 60 bytes 在前段的位置：{o}")
        o2 = img.find(r[:60])
        print(f"  ……在整個 GRAPH.IMG 的位置：{o2}")

# 當成 8×15 的半形字（每字 15 bytes）畫出來
n8 = len(head) // 15
per = 64
nrow = (n8 + per - 1) // per
sheet = [[0]*(per*8) for _ in range(nrow*15)]
for g in range(n8):
    gx, gy = (g % per)*8, (g // per)*15
    for y in range(15):
        b = head[g*15 + y]
        for x in range(8):
            sheet[gy+y][gx+x] = (b >> (7-x)) & 1
os.makedirs('/work/workplace/assets/font', exist_ok=True)
png.write_mono('/work/workplace/assets/font/graph-img-head-8x15.png', per*8, nrow*15, sheet)
print(f"寫出 graph-img-head-8x15.png（當成 {n8} 個 8×15 半形字）")

print("\n當成 8×15 半形字，前 6 個：")
for g in range(6):
    for y in range(15):
        b = head[g*15+y]
        print('   ' + ''.join('#' if b >> (7-x) & 1 else '.' for x in range(8)))
    print('   --------')
