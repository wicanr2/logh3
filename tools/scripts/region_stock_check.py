"""比對星域記錄的 `+0x68`／`+0x70`（飛彈／艦載機的星域彙總）與行星加總。

    tools/py.sh tools/scripts/region_stock_check.py workplace/<傾印>.bin [...]

傾印是 1 MB 的線性記憶體（`tools/oracle.sh ... -dump-mem lin:0:1048576:...`）。
用途是回答「`2B5F:02B2` 每回合加進星域 `+0x68`／`+0x70` 的那一筆，
會不會被回合尾巴的 `2703:12EB`（從行星重算）蓋掉」。

位址（`docs/internals/22`、`docs/internals/29`）：
    星域表 線性 0x5F2CC，32 筆 × 126 bytes
    行星表 線性 0x53132，81 筆 × 82 bytes（沒有檔頭）
    每個星域有幾顆行星：`PLADAT.DAT` 檔頭那 32 個 word（這裡用 ds:CFB8 的複本）
"""
import struct
import sys

REGION_BASE = 0x5F2CC
REGION_SIZE = 126
PLANET_BASE = 0x53132
PLANET_SIZE = 82
COUNTS_BASE = 0x54B30 + 0xCFB8  # ds:CFB8，32 個 word


def u32(buf, off):
    return struct.unpack_from("<I", buf, off)[0]


def u16(buf, off):
    return struct.unpack_from("<H", buf, off)[0]


def main(paths):
    for path in paths:
        buf = open(path, "rb").read()
        counts = [u16(buf, COUNTS_BASE + 2 * k) for k in range(32)]
        print("== %s ==  行星數合計 %d" % (path, sum(counts)))
        first = 0
        starts = []
        for k in range(32):
            starts.append(first)
            first += counts[k]
        bad = 0
        for k in range(32):
            r = REGION_BASE + k * REGION_SIZE
            owner = buf[r + 0x12]
            flags = buf[r + 0x13]
            econ = u32(buf, r + 0x1A)
            defence = u16(buf, r + 0x20)
            # 行星加總，依行星擁有者分組
            psum = {0: [0, 0], 1: [0, 0], 2: [0, 0]}
            for i in range(counts[k]):
                p = PLANET_BASE + (starts[k] + i) * PLANET_SIZE
                po = buf[p + 0x11]
                if po > 2:
                    continue
                psum[po][0] += u32(buf, p + 0x3C)
                psum[po][1] += u32(buf, p + 0x40)
            for o in (0, 1):
                reg68 = u32(buf, r + 0x68 + o * 4)
                reg70 = u32(buf, r + 0x70 + o * 4)
                if reg68 == psum[o][0] and reg70 == psum[o][1]:
                    continue
                bad += 1
                print("  R%-2d 擁%d 旗%02X 經濟%-6d 防%-5d 上限%-6d | 陣營%d "
                      "+68 %-8d(行星 %-8d) +70 %-8d(行星 %-8d)"
                      % (k, owner, flags, econ, defence, econ + defence, o,
                         reg68, psum[o][0], reg70, psum[o][1]))
        print("  對不上的 (星域,陣營) 組數：%d" % bad)


if __name__ == "__main__":
    main(sys.argv[1:])
