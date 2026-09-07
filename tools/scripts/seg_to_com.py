"""把記憶體傾印裡的一個段切出來，補成 IDA 吃得下的 `.COM`。

    seg_to_com.py <傾印.bin> <傾印起始線性位址> <段值> <長度> <輸出.com>

參數都是十六進位。`GIN3PS.EXE` 是 GLUE 包過的，程式碼只能從執行期倒
（`docs/re/43`）；倒出來的段前面補 `0x100` bytes 存成 `.COM`，IDA 就會
當 16 位元載入，段內位移 X 落在「段起點 ＋ 0x100 ＋ X」。
"""
import sys

dump = open(sys.argv[1], "rb").read()
base = int(sys.argv[2], 16)
seg = int(sys.argv[3], 16)
length = int(sys.argv[4], 16)
out = sys.argv[5]

start = seg * 16 - base
if start < 0 or start + length > len(dump):
    raise SystemExit("段 %04X 的 %X bytes 不在傾印範圍內（傾印 %X–%X）"
                     % (seg, length, base, base + len(dump)))
open(out, "wb").write(b"\x00" * 0x100 + dump[start:start + length])
print("段 %04X（線性 %05X）%d bytes → %s" % (seg, seg * 16, length, out))
