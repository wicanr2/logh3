"""把原版 AI 的「劇本進軍路線表」從記憶體傾印裡走出來。

    tools/py.sh tools/scripts/ai_script_lists.py <dump> [DS 線性位址]

`2D54:0000` 用 `2D54:0163()` 回的索引 n 去查 `ds:5EF2[n]`（帝國）或
`ds:5F22[n]`（同盟），拿到一組遠指標陣列（以 0:0 結尾），
每個元素指向一串 0xFF 結尾的星域編號（`docs/re/178`）。
"""
import json
import pathlib
import struct
import sys

mem = pathlib.Path(sys.argv[1]).read_bytes()
DS = int(sys.argv[2], 0) if len(sys.argv) > 2 else 0x54B30


def far(at):
    off, seg = struct.unpack_from('<HH', mem, at)
    return seg * 16 + off, seg, off


def region_list(lin):
    out = []
    while mem[lin] != 0xFF and len(out) < 64:
        out.append(mem[lin])
        lin += 1
    return out


def group(lin):
    """0:0 結尾的遠指標陣列 → 每個指到一串星域編號。"""
    lists = []
    while True:
        target, seg, off = far(lin)
        if seg == 0 and off == 0:
            return lists
        lists.append({'ptr': '%04X:%04X' % (seg, off), 'regions': region_list(target)})
        lin += 4
        if len(lists) > 64:
            return lists


out = {}
for name, base in (('帝國 ds:5EF2', 0x5EF2), ('同盟 ds:5F22', 0x5F22)):
    entries = []
    for n in range(12):
        target, seg, off = far(DS + base + n * 4)
        entries.append({'n': n, 'ptr': '%04X:%04X' % (seg, off), 'lists': group(target)})
    out[name] = entries
print(json.dumps(out, ensure_ascii=False, indent=1))
