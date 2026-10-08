"""GIN3PS 函式進度台帳。

    ledger.py build  FUNCTIONS.json RELOCS.json UNRELOCATED.bin OUT_DIR
        由 IDA 匯出、重定位表與未重定位映像產生 OUT_DIR/functions.tsv（可重新產生，不手改）。
    ledger.py summary OUT_DIR
        讀 functions.tsv 與 status.tsv，印出各等級的函式數與 bytes。

functions.tsv 欄位：
    image_offset  映像內位移（hex）；IDA 線性位址 = +0x10000，dosgolem 線性位址 = +0x3490
    golem_linear  dosgolem 載入在 0349 時的線性位址（hex），與 docs/re 使用的 seg:off 同一個空間
    ida_name      IDA 自動名稱，只是座標，不是語意
    size          函式 bytes（IDA 的函式範圍）
    far           1 ＝ retf 結尾的 far 函式
    callers       IDA 找到的呼叫／跳轉交叉參照數
    relocs        函式範圍內的重定位 word 數（比對 bytes 時要遮罩）
    sha256_16     未重定位映像中該範圍 bytes 的 SHA-256 前 16 字元

status.tsv（手動維護，欄位以 tab 分隔，第一列是標題）：
    golem_linear  level  source  evidence
level 只有 byte-matching、diff-tested、reasoned；沒有列的函式視為 none。
"""
import csv
import hashlib
import json
import os
import sys
from bisect import bisect_left

IMAGE_SIZE = 371594
LEVELS = ("byte-matching", "diff-tested", "reasoned")


def build(functions_json, relocs_json, image_bin, out_dir):
    fns = json.load(open(functions_json))["functions"]
    relocs = sorted(json.load(open(relocs_json))["relocations"])
    img = open(image_bin, "rb").read()
    assert len(img) == IMAGE_SIZE, len(img)
    fns.sort(key=lambda f: f["image_offset"])
    os.makedirs(out_dir, exist_ok=True)
    path = os.path.join(out_dir, "functions.tsv")
    with open(path, "w", newline="") as fh:
        w = csv.writer(fh, delimiter="\t", lineterminator="\n")
        w.writerow(["image_offset", "golem_linear", "ida_name", "size", "far", "callers", "relocs", "sha256_16"])
        for f in fns:
            lo, hi = f["image_offset"], f["image_offset"] + f["size"]
            nrel = bisect_left(relocs, hi) - bisect_left(relocs, lo)
            digest = hashlib.sha256(img[lo:hi]).hexdigest()[:16]
            w.writerow([f"{lo:05x}", f"{f['golem_linear']:05x}", f["name"], f["size"],
                        int(f["far"]), f["callers"], nrel, digest])
    print(f"{len(fns)} 個函式 -> {path}")


def read_tsv(path):
    with open(path, newline="") as fh:
        return list(csv.DictReader(fh, delimiter="\t"))


def summary(out_dir):
    fns = read_tsv(os.path.join(out_dir, "functions.tsv"))
    status_path = os.path.join(out_dir, "status.tsv")
    status = {r["golem_linear"].lower(): r for r in read_tsv(status_path)} if os.path.exists(status_path) else {}
    known = {f["golem_linear"] for f in fns}
    bad = [k for k in status if k not in known]
    if bad:
        print("status.tsv 有不在 functions.tsv 的位址：", ", ".join(bad))
        sys.exit(1)
    total_n, total_b = len(fns), sum(int(f["size"]) for f in fns)
    print(f"函式 {total_n} 個，{total_b} bytes（映像 {IMAGE_SIZE} bytes）")
    for level in LEVELS + ("none",):
        rows = [f for f in fns if status.get(f["golem_linear"], {"level": "none"})["level"] == level]
        b = sum(int(f["size"]) for f in rows)
        print(f"  {level:14s} {len(rows):5d} 個 ({len(rows) * 100 / total_n:5.1f}%)  {b:7d} bytes ({b * 100 / total_b:5.1f}%)")


def main():
    cmd = sys.argv[1]
    if cmd == "build":
        build(*sys.argv[2:6])
    elif cmd == "summary":
        summary(sys.argv[2])
    else:
        sys.exit("用法: ledger.py build|summary ...")


main()
