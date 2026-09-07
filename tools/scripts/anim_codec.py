"""開頭動畫那套 XOR-RLE 的解析器，順便拿去試 `COMANI.BMG`。

語法（`docs/re/36`）：

    XX YY（XX ≠ YY）      把 XX XOR 進一格，只吃掉 XX
    XX XX n（n 有 bit7）  把 XX XOR 進 n & 0x7F 格
    XX XX n1 n2           格數 ＝ (n1 << 8) | n2
    00 00 …               一樣的兩種計數，但只前進不寫

一格畫面 ＝ 四個 plane，**每個 plane 後面接一個分隔位元組**。

    anim_codec.py intro                 驗開頭動畫（374 格逐格對長度表）
    anim_codec.py comani                拿同一套語法試 COMANI.BMG 的每一段
"""
import struct
import sys


def parse_plane(buf, p, cells):
    """從 p 解一個 plane，走滿 cells 格就停。回 (吃掉到哪, 是否走完)。"""
    n = 0
    while n < cells:
        if p + 1 >= len(buf):
            return p, False
        a, b = buf[p], buf[p + 1]
        if a != b:
            p += 1
            n += 1
            continue
        p += 2
        if p >= len(buf):
            return p, False
        c = buf[p]
        p += 1
        if c & 0x80:
            run = c & 0x7F
        else:
            if p >= len(buf):
                return p, False
            run = (c << 8) | buf[p]
            p += 1
        if run == 0:
            return p, False
        n += run
    return p, n == cells


def parse_frame(buf, p, cells):
    """四個 plane ＋ 四個分隔位元組。"""
    for _ in range(4):
        p, ok = parse_plane(buf, p, cells)
        if not ok:
            return p, False
        p += 1  # 分隔位元組
    return p, True


def intro():
    lens = list(struct.unpack_from('<374H', open('/orig/IBMFLS16.350', 'rb').read(), 0))
    d = open('/orig/IBM16.350', 'rb').read()
    p, bad = 0, 0
    for i, want in enumerate(lens):
        q, ok = parse_frame(d, p, 80 * 227)
        if not ok or q - p != want:
            bad += 1
            if bad <= 3:
                print('第 %d 格：吃掉 %d，長度表說 %d，走完 %s' % (i, q - p, want, ok))
        p = q
    print('開頭動畫：374 格裡對不上的 %d，吃到 %d／%d bytes' % (bad, p, len(d)))


def comani():
    d = open('/orig/COMANI.BMG', 'rb').read()
    first = struct.unpack_from('<H', d, 0)[0]
    ent, p = [], 2
    while p + 10 <= first:
        ent.append((d[p:p + 4], struct.unpack_from('<H', d, p + 4)[0],
                    struct.unpack_from('<I', d, p + 6)[0]))
        p += 10
    hit = 0
    sizes = {}
    for name, ln, off in ent:
        seg = d[off:off + ln]
        found = None
        for cells in range(1, 4001):
            q, ok = parse_frame(seg, 0, cells)
            if ok and q == ln:
                found = cells
                break
        if found:
            hit += 1
            sizes[found] = sizes.get(found, 0) + 1
    print('COMANI：%d／%d 段能用同一套語法吃光' % (hit, len(ent)))
    for k in sorted(sizes, key=lambda k: -sizes[k])[:8]:
        print('   每個 plane %d 格 → %d 段' % (k, sizes[k]))


{'intro': intro, 'comani': comani}[sys.argv[1]]()
