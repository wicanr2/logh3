# -*- coding: utf-8 -*-
"""對艦隊記錄（`ds:C074`，122 bytes）做逐欄的值普查。

    tools/py.sh tools/scripts/fleet_fields.py <來源> [來源 ...]

來源可以是 1 MB 執行期傾印（艦隊表在線性 `0x60BA4`）或 3904 bytes 的
`FLEET.DT?` 存檔（從 0 起，版面相同，`docs/re/194`）——依檔案大小自動判斷。

每一格印：所有在用記錄（`+0x04 != 0xFF`）的相異值個數、最小、最大、
出現最多的幾個值，以及「這一格在不同來源之間有沒有變」。

**值的分布是線索不是證據**——一格看起來像旗標不代表有人讀它。
要下語意結論仍然要找到消費者（`loadbyte.py`／`field_xref.py`）。
"""
import collections
import pathlib
import sys

FLEET_LIN, N, SZ = 0x60BA4, 32, 122


def records(path):
    buf = pathlib.Path(path).read_bytes()
    base = FLEET_LIN if len(buf) > N * SZ else 0
    out = []
    for i in range(N):
        r = buf[base + i * SZ: base + (i + 1) * SZ]
        if len(r) == SZ and r[4] != 0xFF:
            out.append((i, r))
    return out


srcs = [(p, records(p)) for p in sys.argv[1:]]
for p, rs in srcs:
    print('%-34s 在用 %d 筆' % (p, len(rs)))
print()
print('位移  相異 最小 最大  最常見（值×次數）                      跨來源')
for off in range(SZ):
    allv = []
    per_src = []
    for p, rs in srcs:
        vals = [r[off] for _, r in rs]
        allv += vals
        per_src.append(tuple(sorted(set(vals))))
    c = collections.Counter(allv)
    top = ' '.join('%d×%d' % (v, n) for v, n in c.most_common(5))
    same = '相同' if len(set(per_src)) == 1 else '不同'
    print('+0x%02X  %4d %4d %4d  %-40s %s'
          % (off, len(c), min(allv), max(allv), top, same))

# 兩條可以自己開口的斷言：`+0x03` 應該等於這一筆的槽號、`+0x02` 是
# 畫面上的「第 N 艦隊」（1 起算、同陣營內不重複）。對不上就先查斷言。
print()
for p, rs in srcs:
    bad3 = [(i, r[3]) for i, r in rs if r[3] != i]
    nums = {}
    for i, r in rs:
        nums.setdefault(r[1], []).append(r[2])
    dup = {f: v for f, v in nums.items() if len(v) != len(set(v))}
    print('%-34s +0x03≠槽號 %d 筆%s；+0x02 同陣營重號 %s'
          % (p, len(bad3), (' %s' % bad3[:5]) if bad3 else '',
             dup if dup else '無'))
