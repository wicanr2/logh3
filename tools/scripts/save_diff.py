"""比較同一種存檔檔案的幾個版本，找出記錄長度與會變的欄位。

    save_diff.py <stride> <檔案> <檔案> [...]

`stride` ＝ 0 時只印「不同的位元組落在哪些位移」；非零時把位移對 stride
取餘數分組，看得出記錄裡哪幾格會變。存檔是執行期的表原樣寫出去
（`docs/re/41`），所以這裡看到的欄位就是記憶體裡的欄位。
"""
import sys

stride = int(sys.argv[1])
paths = sys.argv[2:]
data = [open(p, "rb").read() for p in paths]
n = min(len(d) for d in data)
print("檔案 %d 份，長度 %s" % (len(data), [len(d) for d in data]))

diff = [i for i in range(n) if len({d[i] for d in data}) > 1]
print("不同的位元組 %d / %d" % (len(diff), n))
if stride:
    print("記錄長度 %d ＝ %d 筆" % (stride, n // stride))
    off = {}
    for i in diff:
        off.setdefault(i % stride, []).append(i // stride)
    print("會變的位移（位移: 幾筆）：")
    line = []
    for o in sorted(off):
        line.append("+0x%02X:%d" % (o, len(off[o])))
    print("  " + "  ".join(line))
    same = [o for o in range(stride) if o not in off]
    print("完全不變的位移 %d 個： %s" % (
        len(same), " ".join("+0x%02X" % o for o in same)))
else:
    runs = []
    for i in diff:
        if runs and i == runs[-1][1] + 1:
            runs[-1][1] = i
        else:
            runs.append([i, i])
    print("  " + "  ".join("0x%X-0x%X" % (a, b) for a, b in runs[:60]))
