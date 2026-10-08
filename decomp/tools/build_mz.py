"""把 GIN3PS 的解包映像（未重定位）與重定位表組成標準 MZ 檔，並驗證重定位後等於最終映像。

用法（Docker 內，python3 -I）：
    build_mz.py PRE.bin FINAL.bin RELOCS.json OUT_DIR

輸出（OUT_DIR）：
    gin3ps-unrelocated.bin   未重定位的映像，371,594 bytes
    gin3ps-final.bin         載入在 0349 並重定位後的映像（入口那一刻）
    GIN3PS_UNPACKED.EXE      以上映像加重定位表的 MZ 檔
    image.json               雜湊、入口與堆疊位置、重定位數

MZ 欄位的來源：入口 CS:IP 與 SS:SP 取自 dosgolem 在最終入口的暫存器
（CS:IP=0349:0079、SS:SP=63EB:0080，載入段 0349，減去載入段得到相對值）。
min／max alloc 不影響反組譯，沿用 0／0xFFFF。
"""
import hashlib
import json
import os
import struct
import sys

LOAD_SEG = 0x0349
BASE = LOAD_SEG * 16
SIZE = 371594
ENTRY_CS, ENTRY_IP = 0x0349, 0x0079
ENTRY_SS, ENTRY_SP = 0x63EB, 0x0080


def sha(b):
    return hashlib.sha256(b).hexdigest()


def main():
    pre = open(sys.argv[1], "rb").read()
    final = open(sys.argv[2], "rb").read()
    info = json.load(open(sys.argv[3]))
    out = sys.argv[4]
    os.makedirs(out, exist_ok=True)

    pre_img = pre[BASE:BASE + SIZE]
    fin_img = final[BASE:BASE + SIZE]
    relocs = info["relocations"]
    assert len(relocs) == 5991, len(relocs)

    # 驗證：未重定位映像 + 重定位表 = 最終映像
    rebuilt = bytearray(pre_img)
    for o in relocs:
        w = (rebuilt[o] | (rebuilt[o + 1] << 8)) + LOAD_SEG
        rebuilt[o] = w & 0xFF
        rebuilt[o + 1] = (w >> 8) & 0xFF
    assert bytes(rebuilt) == fin_img, "套用重定位後與最終映像不同"

    nreloc = len(relocs)
    header_bytes = (28 + 4 * nreloc + 15) // 16 * 16
    total = header_bytes + SIZE
    hdr = struct.pack(
        "<2s13H",
        b"MZ",
        total % 512,
        (total + 511) // 512,
        nreloc,
        header_bytes // 16,
        0,
        0xFFFF,
        ENTRY_SS - LOAD_SEG,
        ENTRY_SP,
        0,
        ENTRY_IP,
        ENTRY_CS - LOAD_SEG,
        28,
        0,
    )
    table = b"".join(struct.pack("<HH", o & 0xF, o >> 4) for o in relocs)
    exe = hdr + table
    exe += b"\0" * (header_bytes - len(exe))
    exe += pre_img

    open(os.path.join(out, "gin3ps-unrelocated.bin"), "wb").write(pre_img)
    open(os.path.join(out, "gin3ps-final.bin"), "wb").write(fin_img)
    open(os.path.join(out, "GIN3PS_UNPACKED.EXE"), "wb").write(exe)
    meta = {
        "load_seg": LOAD_SEG,
        "image_size": SIZE,
        "relocation_count": nreloc,
        "entry_cs_ip": [ENTRY_CS, ENTRY_IP],
        "entry_ss_sp": [ENTRY_SS, ENTRY_SP],
        "image_unrelocated_sha256": sha(pre_img),
        "image_final_sha256": sha(fin_img),
        "mz_exe_sha256": sha(exe),
        "mz_exe_size": len(exe),
        "header_bytes": header_bytes,
    }
    json.dump(meta, open(os.path.join(out, "image.json"), "w"), indent=2)
    print(json.dumps(meta, indent=2))


main()
