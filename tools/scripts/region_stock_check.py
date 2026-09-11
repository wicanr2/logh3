"""驗 `2B5F:02B2`：星域的飛彈／艦載機存量每回合回充（`docs/re/273`）。

兩種輸入，看檔案大小自動分：

    # 1 MB 線性記憶體：比 `+0x68`／`+0x70` 與「該星域行星的 +0x3C／+0x40 加總」
    tools/py.sh tools/scripts/region_stock_check.py workplace/monthblend.bin

    # 4032 bytes（32 × 126 的星域表，`-dump-mem-at …:lin:5F2CC:4032:…`）
    # 兩份一起給就是回充前後：逐格印出「有沒有被補、有沒有正好等於上限」
    tools/py.sh tools/scripts/region_stock_check.py \
        workplace/re273_rpre.bin workplace/re273_rpost.bin

位址（`docs/internals/22`、`docs/internals/29`）：

    星域表 線性 0x5F2CC，32 筆 × 126 bytes
    行星表 線性 0x53132，81 筆 × 82 bytes（沒有檔頭）
    每個星域有幾顆行星：ds:CFB8 那 32 個 word（＝ `PLADAT.DAT` 的檔頭）

規則（`docs/spec/039`）：擁有者不是 2、而且 `+0x13` bit0 沒設的星域，
每回合 `+0x68/+0x70 [擁有者]` 各加 `⌊(經濟力 ＋ 防衛力) ÷ 10⌋`，夾在
經濟力＋防衛力。
"""
import struct
import sys

REGION_BASE = 0x5F2CC
REGION_SIZE = 126
REGION_COUNT = 32
REGION_TABLE = REGION_SIZE * REGION_COUNT
PLANET_BASE = 0x53132
PLANET_SIZE = 82
COUNTS_BASE = 0x54B30 + 0xCFB8  # ds:CFB8，32 個 word


def u32(buf, off):
    return struct.unpack_from("<I", buf, off)[0]


def u16(buf, off):
    return struct.unpack_from("<H", buf, off)[0]


def regions(buf):
    """把輸入切成 32 筆星域記錄。1 MB 的傾印從 0x5F2CC 切，4032 的直接切。"""
    base = REGION_BASE if len(buf) > REGION_TABLE else 0
    return [buf[base + k * REGION_SIZE:base + (k + 1) * REGION_SIZE]
            for k in range(REGION_COUNT)]


def planet_sums(buf):
    """每個星域、每個擁有者的行星 `+0x3C`／`+0x40` 加總。要 1 MB 的傾印。"""
    counts = [u16(buf, COUNTS_BASE + 2 * k) for k in range(REGION_COUNT)]
    out, first = [], 0
    for k in range(REGION_COUNT):
        acc = {0: [0, 0], 1: [0, 0], 2: [0, 0]}
        for i in range(counts[k]):
            p = PLANET_BASE + (first + i) * PLANET_SIZE
            owner = buf[p + 0x11]
            if owner <= 2:
                acc[owner][0] += u32(buf, p + 0x3C)
                acc[owner][1] += u32(buf, p + 0x40)
        out.append(acc)
        first += counts[k]
    return counts, out


def against_planets(path):
    """1 MB 傾印：星域彙總對行星加總。兩者相等 ＝ 這一刻是 `2703:12EB` 說了算。"""
    buf = open(path, "rb").read()
    counts, sums = planet_sums(buf)
    print("== %s ==  行星數合計 %d" % (path, sum(counts)))
    bad = 0
    for k, rec in enumerate(regions(buf)):
        econ, defence = u32(rec, 0x1A), u16(rec, 0x20)
        for o in (0, 1):
            got = (u32(rec, 0x68 + o * 4), u32(rec, 0x70 + o * 4))
            want = tuple(sums[k][o])
            if got == want:
                continue
            bad += 1
            print("  R%-2d 擁%d 旗%02X 經濟%-6d 防%-5d 上限%-6d | 陣營%d "
                  "+68 %-8d(行星 %-8d) +70 %-8d(行星 %-8d)"
                  % (k, rec[0x12], rec[0x13], econ, defence, econ + defence,
                     o, got[0], want[0], got[1], want[1]))
    print("  對不上的 (星域,陣營) 組數：%d" % bad)


def before_after(pre_path, post_path):
    """兩份星域表：回充前後。有主的要補到上限，無主與另一方那一半不准動。"""
    pre, post = regions(open(pre_path, "rb").read()), \
        regions(open(post_path, "rb").read())
    print("== %s → %s ==" % (pre_path, post_path))
    ok_owned = ok_skipped = bad = 0
    for k in range(REGION_COUNT):
        a, b = pre[k], post[k]
        owner, flags = a[0x12], a[0x13]
        cap = u32(a, 0x1A) + u16(a, 0x20)
        skip = owner > 1 or (flags & 1)
        moved = [(o, f) for o in (0, 1) for f in (0x68, 0x70)
                 if u32(a, f + o * 4) != u32(b, f + o * 4)]
        if skip:
            if moved:
                bad += 1
                print("  R%-2d 擁%d 旗%02X **該跳過卻動了** %s" % (k, owner, flags, moved))
            else:
                ok_skipped += 1
            continue
        got = (u32(b, 0x68 + owner * 4), u32(b, 0x70 + owner * 4))
        other = 1 - owner
        untouched = all(u32(a, f + other * 4) == u32(b, f + other * 4)
                        for f in (0x68, 0x70))
        want = min(u32(a, 0x68 + owner * 4) + cap // 10, cap), \
            min(u32(a, 0x70 + owner * 4) + cap // 10, cap)
        mark = "OK" if got == want and untouched else "**不合**"
        if got == want and untouched:
            ok_owned += 1
        else:
            bad += 1
        print("  R%-2d 擁%d 旗%02X 上限%-7d 補%-6d | +68 %-8d→%-8d +70 %-8d→%-8d %s"
              % (k, owner, flags, cap, cap // 10,
                 u32(a, 0x68 + owner * 4), got[0],
                 u32(a, 0x70 + owner * 4), got[1], mark))
    print("  有主且算式相符 %d、依規則跳過 %d、不合 %d"
          % (ok_owned, ok_skipped, bad))


if __name__ == "__main__":
    args = sys.argv[1:]
    # ⚠ 沒給參數就印用法再走人。**不要安靜結束**——驗算工具不出聲和
    # 「跑過了、全部通過」長得一模一樣，而下一個人只會看到 exit 0。
    if not args:
        print(__doc__.strip(), file=sys.stderr)
        sys.exit(2)
    if len(args) == 2 and all(
            len(open(p, "rb").read()) == REGION_TABLE for p in args):
        before_after(*args)
    else:
        for p in args:
            against_planets(p)
