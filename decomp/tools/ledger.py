"""GIN3PS 函式進度台帳。

    ledger.py build   FUNCTIONS.json RELOCS.json UNRELOCATED.bin OUT_DIR [FINAL.bin]
        由 IDA 匯出（v2）、重定位表與未重定位映像產生 OUT_DIR/functions.tsv 與 callgraph.tsv。
        兩者都可重新產生，不手改。
    ledger.py summary OUT_DIR
        按分類與等級印出函式數與 bytes，並列出各分類的模組數。
    ledger.py show    OUT_DIR SEG:OFF
        印出一個函式的特徵與呼叫關係（SEG:OFF 用 docs/re 的記法，如 203A:0424）。
    ledger.py mark    OUT_DIR SEG:OFF LEVEL SOURCE EVIDENCE
        寫 OUT_DIR/status.d/SEG_OFF.tsv（一個函式一個檔）。
    ledger.py order   OUT_DIR [CLASS ...]
        由下而上的反編譯順序（被呼叫者先），預設 game-pure 與 game-logic。

模組：far call／far jmp 的段值（重定位的 word）就是被呼叫函式所屬原始模組的程式碼段。
Large model 下一個模組是一個程式碼段，所以 129 個不同的段值就是 129 個模組。
運行時段值 = 未重定位段值 + 0x0349，與 docs/re 的 SEG:OFF 同一個空間。

分類規則（機械，由特徵與呼叫閉包決定，不含語意推論）：
    runtime       模組是啟動碼（0349）或位於映像尾端的函式庫區（運行時段 >= 3C5B）
    video-dos     自己有 int、in／out、視訊段立即值（A000／A800／B000／B800）或 cli／sti／pushf／popf
    hw-dependent  自己沒有，但呼叫閉包裡有 video-dos 或有硬體特徵的 runtime 函式
    unknown       呼叫閉包（含自己）有間接呼叫，目標無法靜態決定
    game-pure     以上皆非，且呼叫閉包內沒有任何非區域變數的記憶體寫入
    game-logic    以上皆非，但閉包內有記憶體寫入
注意：視訊記憶體若透過變數指定段（不是立即值）就不會被偵測到，video-dos 是下限。
"""
import csv
import hashlib
import json
import os
import sys
from bisect import bisect_left, bisect_right

IMAGE_SIZE = 371594
RT_BASE = 0x0349
LEVELS = ("byte-matching", "diff-tested", "reasoned")
RUNTIME_STARTUP_SEG = 0x0349
RUNTIME_TAIL_SEG = 0x3C5B
COLUMNS = ["image_offset", "golem_linear", "seg_off", "ida_name", "size", "far", "callers", "relocs",
           "class", "ints", "io", "vram", "ind", "stores", "callees", "sha256_16", "sha256_final_16"]


def read_tsv(path):
    with open(path, newline="") as fh:
        return list(csv.DictReader(fh, delimiter="\t"))


def load_modules(img, relocs):
    segs = set()
    for r in relocs:
        if r >= 3 and img[r - 3] in (0x9A, 0xEA):
            segs.add(img[r] | (img[r + 1] << 8))
    return sorted(segs)


def build(functions_json, relocs_json, image_bin, out_dir, final_bin=None):
    fns = json.load(open(functions_json))["functions"]
    relocs = sorted(json.load(open(relocs_json))["relocations"])
    img = open(image_bin, "rb").read()
    assert len(img) == IMAGE_SIZE, len(img)
    fin = open(final_bin, "rb").read() if final_bin else None
    assert fin is None or len(fin) == IMAGE_SIZE
    fns.sort(key=lambda f: f["image_offset"])
    segs = load_modules(img, relocs)
    bases = [s * 16 for s in segs]
    byea = {f["ida_ea"]: f for f in fns}

    def mod_index(off):
        return bisect_right(bases, off) - 1

    for f in fns:
        i = mod_index(f["image_offset"])
        f["_seg"] = segs[i] if i >= 0 else None
        f["_seg_rt"] = (segs[i] + RT_BASE) if i >= 0 else None
        f["_off"] = f["image_offset"] - bases[i] if i >= 0 else f["image_offset"]
        f["_callees"] = sorted({t for t in f["callees_far"] + f["callees_near"] if t in byea and t != f["ida_ea"]})
        f["_own_hw"] = bool(f["ints"] or f["io"] or f["vram_imm"] or f["flag_ops"])
        f["_runtime"] = f["_seg_rt"] is not None and (f["_seg_rt"] == RUNTIME_STARTUP_SEG or f["_seg_rt"] >= RUNTIME_TAIL_SEG)

    # 呼叫閉包
    reach = {}

    def closure(ea):
        if ea in reach:
            return reach[ea]
        reach[ea] = set()           # 先放空集合處理遞迴
        acc = set()
        for t in byea[ea]["_callees"]:
            acc.add(t)
            acc |= closure(t)
        reach[ea] = acc
        return acc

    sys.setrecursionlimit(10000)
    for f in fns:
        closure(f["ida_ea"])

    for f in fns:
        cl = reach[f["ida_ea"]] | {f["ida_ea"]}
        if f["_runtime"]:
            cls = "runtime"
        elif f["_own_hw"]:
            cls = "video-dos"
        elif any(byea[x]["_own_hw"] for x in cl):
            cls = "hw-dependent"
        elif any(byea[x]["indirect_calls"] for x in cl):
            cls = "unknown"
        elif not any(byea[x]["stores_nonlocal"] or byea[x]["string_stores"] for x in cl):
            cls = "game-pure"
        else:
            cls = "game-logic"
        f["_class"] = cls

    os.makedirs(out_dir, exist_ok=True)
    with open(os.path.join(out_dir, "functions.tsv"), "w", newline="") as fh:
        w = csv.writer(fh, delimiter="\t", lineterminator="\n")
        w.writerow(COLUMNS)
        for f in fns:
            lo, hi = f["image_offset"], f["image_offset"] + f["size"]
            nrel = bisect_left(relocs, hi) - bisect_left(relocs, lo)
            digest = hashlib.sha256(img[lo:hi]).hexdigest()[:16]
            digest_final = hashlib.sha256(fin[lo:hi]).hexdigest()[:16] if fin else "-"
            seg_off = "%04X:%04X" % (f["_seg_rt"], f["_off"]) if f["_seg_rt"] is not None else "----:%05X" % lo
            w.writerow([f"{lo:05x}", f"{f['golem_linear']:05x}", seg_off, f["name"], f["size"], int(f["far"]),
                        f["callers"], nrel, f["_class"], ",".join(str(i) for i in f["ints"]) or "-",
                        f["io"], f["vram_imm"], f["indirect_calls"], f["stores_nonlocal"] + f["string_stores"],
                        len(f["_callees"]), digest, digest_final])
    with open(os.path.join(out_dir, "insns.tsv"), "w", newline="") as fh:
        for f in fns:
            fh.write("%05x\t%s\n" % (f["golem_linear"], ",".join("%x" % o for o in f["insn_offsets"] if o >= 0)))
    with open(os.path.join(out_dir, "callgraph.tsv"), "w", newline="") as fh:
        w = csv.writer(fh, delimiter="\t", lineterminator="\n")
        w.writerow(["caller_golem_linear", "callee_golem_linear"])
        for f in fns:
            for t in f["_callees"]:
                w.writerow([f"{f['golem_linear']:05x}", f"{byea[t]['golem_linear']:05x}"])
    n_mod = len({f["_seg"] for f in fns})
    print(f"{len(fns)} 個函式、{n_mod} 個模組 -> {out_dir}")


def load(out_dir):
    fns = read_tsv(os.path.join(out_dir, "functions.tsv"))
    status = {}
    status_path = os.path.join(out_dir, "status.tsv")
    if os.path.exists(status_path):
        for r in read_tsv(status_path):
            status[r["golem_linear"].lower()] = r
    status_dir = os.path.join(out_dir, "status.d")        # 每個函式一個檔，並行工作時不會互相衝突
    if os.path.isdir(status_dir):
        for name in sorted(os.listdir(status_dir)):
            if name.endswith(".tsv"):
                for r in read_tsv(os.path.join(status_dir, name)):
                    status[r["golem_linear"].lower()] = r
    known = {f["golem_linear"] for f in fns}
    bad = [k for k in status if k not in known]
    if bad:
        print("status.tsv 有不在 functions.tsv 的位址：", ", ".join(bad))
        sys.exit(1)
    for f in fns:
        f["level"] = status.get(f["golem_linear"], {"level": "none"})["level"]
    return fns, status


def summary(out_dir):
    fns, _ = load(out_dir)
    total_n, total_b = len(fns), sum(int(f["size"]) for f in fns)
    print(f"函式 {total_n} 個，{total_b} bytes（映像 {IMAGE_SIZE} bytes）")
    classes = ["game-pure", "game-logic", "unknown", "hw-dependent", "video-dos", "runtime"]
    print(f"{'分類':14s} {'函式':>5s} {'bytes':>7s} {'模組':>4s}  " + "  ".join(f"{l:>13s}" for l in LEVELS + ("none",)))
    for c in classes:
        rows = [f for f in fns if f["class"] == c]
        b = sum(int(f["size"]) for f in rows)
        mods = len({f["seg_off"].split(":")[0] for f in rows})
        cells = []
        for l in LEVELS + ("none",):
            k = [f for f in rows if f["level"] == l]
            cells.append(f"{len(k):4d}/{sum(int(f['size']) for f in k):>7d}")
        print(f"{c:14s} {len(rows):5d} {b:7d} {mods:4d}  " + "  ".join(f"{x:>13s}" for x in cells))
    done = [f for f in fns if f["level"] != "none"]
    print(f"已有等級：{len(done)} 個函式，{sum(int(f['size']) for f in done)} bytes")
    nonrt = [f for f in fns if f["class"] != "runtime"]
    print(f"非 runtime 的函式 {len(nonrt)} 個、{sum(int(f['size']) for f in nonrt)} bytes 才是反編譯的分母")


def show(out_dir, seg_off):
    fns, _ = load(out_dir)
    key = seg_off.upper()
    f = next((x for x in fns if x["seg_off"] == key), None)
    if f is None:
        sys.exit(f"找不到 {seg_off}")
    print(json.dumps(f, ensure_ascii=False, indent=1))
    by_lin = {x["golem_linear"]: x for x in fns}
    callees = [r["callee_golem_linear"] for r in read_tsv(os.path.join(out_dir, "callgraph.tsv"))
               if r["caller_golem_linear"] == f["golem_linear"]]
    callers = [r["caller_golem_linear"] for r in read_tsv(os.path.join(out_dir, "callgraph.tsv"))
               if r["callee_golem_linear"] == f["golem_linear"]]
    print("呼叫:", ", ".join(f"{by_lin[c]['seg_off']}({by_lin[c]['class']})" for c in callees) or "(無)")
    print("被呼叫:", ", ".join(by_lin[c]["seg_off"] for c in callers) or "(無)")


def order(out_dir, classes):
    fns, _ = load(out_dir)
    by_lin = {f["golem_linear"]: f for f in fns}
    edges = {}
    for r in read_tsv(os.path.join(out_dir, "callgraph.tsv")):
        edges.setdefault(r["caller_golem_linear"], []).append(r["callee_golem_linear"])
    want = {f["golem_linear"] for f in fns if f["class"] in classes}
    depth = {}

    def d(k, stack=()):
        if k in depth:
            return depth[k]
        if k in stack:
            return 0
        v = 0
        for c in edges.get(k, []):
            if c in want:
                v = max(v, d(c, stack + (k,)) + 1)
        depth[k] = v
        return v

    sys.setrecursionlimit(10000)
    rows = sorted(want, key=lambda k: (d(k), by_lin[k]["seg_off"]))
    print("depth\tseg_off\tsize\tclass\tlevel\tname")
    for k in rows:
        f = by_lin[k]
        print(f"{depth[k]}\t{f['seg_off']}\t{f['size']}\t{f['class']}\t{f['level']}\t{f['ida_name']}")


def mark(out_dir, seg_off, level, source, evidence):
    if level not in LEVELS:
        sys.exit("level 只能是 " + ", ".join(LEVELS))
    fns = read_tsv(os.path.join(out_dir, "functions.tsv"))
    f = next((x for x in fns if x["seg_off"] == seg_off.upper()), None)
    if f is None:
        sys.exit("找不到 " + seg_off)
    d = os.path.join(out_dir, "status.d")
    os.makedirs(d, exist_ok=True)
    fid = f["seg_off"].replace(":", "_")
    with open(os.path.join(d, fid + ".tsv"), "w") as fh:
        fh.write("golem_linear\tlevel\tsource\tevidence\n%s\t%s\t%s\t%s\n" % (f["golem_linear"], level, source, evidence))
    print("已記錄", f["seg_off"], level)


def main():
    cmd = sys.argv[1]
    if cmd == "build":
        build(*sys.argv[2:7])
    elif cmd == "summary":
        summary(sys.argv[2])
    elif cmd == "show":
        show(sys.argv[2], sys.argv[3])
    elif cmd == "mark":
        mark(*sys.argv[2:7])
    elif cmd == "order":
        order(sys.argv[2], set(sys.argv[3:]) or {"game-pure", "game-logic"})
    else:
        sys.exit("用法: ledger.py build|summary|show|order ...")


main()
