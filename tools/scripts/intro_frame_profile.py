"""逐格印開頭動畫每個 plane 的非零位元組數，並與長度表對照。

    tools/py.sh tools/scripts/intro_frame_profile.py [格數]

用來回答「原版記憶體裡那份 plane 內容是檔案裡的哪一格」——把 oracle 倒出來
的 plane 緩衝區的非零位元組數當簽章，來這張表裡找。**找不到就是它不是
單一格**（`docs/re/261`）。
"""
import sys
sys.path.insert(0, '/work/tools/scripts')
import intro_decode as I

n = int(sys.argv[1]) if len(sys.argv) > 1 else 40
fl = open(I.LENS, 'rb').read()
lens = [fl[2 * i] | fl[2 * i + 1] << 8 for i in range(len(fl) // 2)]
data = open(I.DATA, 'rb').read()
pos = 0
for k in range(min(n, len(lens))):
    bufs = [bytearray(I.PLANE) for _ in range(4)]
    start = pos
    for p in range(4):
        pos += I.decode_plane(data, pos, bufs[p])
        pos += 1                      # 每個 plane 後面一個分隔位元組
    nz = tuple(sum(1 for c in b if c) for b in bufs)
    mark = '' if pos - start == lens[k] else '  ← 與長度表不符'
    print('格 %3d 吃掉 %5d（長度表 %5d）非零 %s%s' % (k, pos - start, lens[k], nz, mark))
