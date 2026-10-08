"""由兩份傾印推出 GIN3PS 解包映像的重定位位置。

輸入：dosgolem 在「最終入口」的傾印（已套用重定位）與重定位前的傾印。
兩份在同一個載入位置，差異應該只有「加上載入段 0x0349 的 word」。
判斷依據：重定位 word 的低位 byte 一定改變（0x0349 的低位是 0x49，不為 0），
所以每一筆都從一個不同的 byte 開始。

用法（在 Docker 內執行，python3 -I）：
    derive_relocs.py FINAL.bin PRE.bin OUT.json [映像起點 映像長度]
"""
import hashlib
import json
import sys

LOAD_SEG = 0x0349
DELTA = LOAD_SEG
BASE = LOAD_SEG * 16
SIZE = 371594


def main():
    final = open(sys.argv[1], "rb").read()
    pre = open(sys.argv[2], "rb").read()
    out = sys.argv[3]
    base = int(sys.argv[4], 0) if len(sys.argv) > 4 else BASE
    size = int(sys.argv[5], 0) if len(sys.argv) > 5 else SIZE
    assert len(final) == len(pre) == 1 << 20

    relocs, other = [], []
    i = base
    end = base + size
    while i < end:
        if final[i] == pre[i]:
            i += 1
            continue
        if i + 1 < end:
            wf = final[i] | (final[i + 1] << 8)
            wp = pre[i] | (pre[i + 1] << 8)
            if (wf - wp) & 0xFFFF == DELTA:
                relocs.append(i - base)
                i += 2
                continue
        other.append(i - base)
        i += 1

    outside = sum(1 for k in range(len(final)) if (k < base or k >= end) and final[k] != pre[k])
    result = {
        "load_seg": LOAD_SEG,
        "image_linear_base": base,
        "image_size": size,
        "final_sha256": hashlib.sha256(final).hexdigest(),
        "pre_sha256": hashlib.sha256(pre).hexdigest(),
        "image_final_sha256": hashlib.sha256(final[base:end]).hexdigest(),
        "image_pre_sha256": hashlib.sha256(pre[base:end]).hexdigest(),
        "relocation_count": len(relocs),
        "other_diff_bytes": len(other),
        "other_diff_first": other[:16],
        "diff_bytes_outside_image": outside,
        "relocations": relocs,
    }
    with open(out, "w") as f:
        json.dump(result, f)
    print(f"重定位 {len(relocs)} 筆；映像內其他差異 {len(other)} bytes；映像外差異 {outside} bytes；"
          f"final 映像 sha256 {result['image_final_sha256'][:16]}… pre 映像 sha256 {result['image_pre_sha256'][:16]}…")


main()
