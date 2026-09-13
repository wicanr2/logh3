"""把「新舊畫面對不上」的收據分類，判斷哪些是同一個遊戲狀態。

    tools/py.sh tools/scripts/triage_receipts.py

讀 `workplace/receipt-check.json`（`tools/check_receipts.sh` 產的）與兩邊的
色號陣列，對每一份對不上的收據做**形狀**分析，輸出
`workplace/receipt-triage.json`。

⚠ **為什麼要先分類再重錄。** 收據沒有版控，重錄就是覆蓋，蓋掉之後
「原本停在哪個畫面」再也查不回來。而 149 份逐一用肉眼看既慢又不可重複
——換一個人看會得到另一份結論。所以先用形狀把「只有游標位置不同」
這種**可證明同狀態**的挑出來，剩下需要人判斷的才真的去看。

分類（由強到弱）：

| 類別 | 判準 | 意思 |
|---|---|---|
| `cursor-only` | 所有差異群集都是 13×16 的方塊，且數量 ≤ 2 | 游標畫在別的地方，其餘逐位元組相同 |
| `tiny` | 差異像素 ≤ 64 且集中在一個 ≤ 16×16 的框裡 | 一小塊，多半是閃爍相位 |
| `palette-only` | 色號陣列相同、只有調色盤不同 | 淡入淡出的相位 |
| `layout-same` | 差異像素 < 5%，且**非零像素的分布**幾乎相同 | 同一個版面，內容有出入 |
| `needs-eyes` | 其餘 | 要人看 |
"""
import collections
import json
import pathlib

ROOT = pathlib.Path("/work")
CHECK = ROOT / "workplace" / "receipt-check.json"
NEW = ROOT / "workplace" / "check"          # tools/check_receipts.sh 倒的新畫面
OLD = ROOT / "workplace" / "oracle"
OUT = ROOT / "workplace" / "receipt-triage.json"
W, H = 640, 480
CURSOR_W, CURSOR_H = 13, 16


def covered_by_cursor_boxes(diff_idx, limit=2):
    """差異像素能否由至多 limit 個 13×16 游標矩形覆蓋。

    舊判準先按相鄰列分群；兩個游標若在同一列附近，會被併成一個寬群而
    假陰性。這裡直接驗幾何覆蓋，游標彼此重疊或同列也不受影響。
    """
    points = {(i % W, i // W) for i in diff_idx}
    if not points:
        return True

    def fits_one(group):
        if not group:
            return True
        xs = [p[0] for p in group]
        ys = [p[1] for p in group]
        return max(xs) - min(xs) < CURSOR_W and max(ys) - min(ys) < CURSOR_H

    if fits_one(points):
        return True
    if limit < 2:
        return False

    # 任一可行的第一個矩形都必須蓋住最左上的一個差異點；枚舉該點在
    # 13×16 矩形內的所有位置，再檢查剩餘像素是否落在第二個矩形內。
    px, py = min(points, key=lambda p: (p[1], p[0]))
    for x0 in range(px - CURSOR_W + 1, px + 1):
        for y0 in range(py - CURSOR_H + 1, py + 1):
            rest = {
                (x, y)
                for x, y in points
                if not (x0 <= x < x0 + CURSOR_W and y0 <= y < y0 + CURSOR_H)
            }
            if fits_one(rest):
                return True
    return False


def clusters(diff_idx):
    """把差異像素按「相鄰的列」分群，回每一群的 (x0, x1, y0, y1, 像素數)。"""
    rows = collections.defaultdict(list)
    for i in diff_idx:
        rows[i // W].append(i % W)
    ys = sorted(rows)
    out, cur = [], [ys[0]]
    for y in ys[1:]:
        # 差一列以內算同一群：游標那種實心方塊每一列都有，中間不會斷。
        if y - cur[-1] <= 2:
            cur.append(y)
        else:
            out.append(cur)
            cur = [y]
    out.append(cur)
    res = []
    for g in out:
        xs = [x for y in g for x in rows[y]]
        n = sum(len(rows[y]) for y in g)
        res.append((min(xs), max(xs), g[0], g[-1], n))
    return res


def classify(old, new, pal_diff):
    d = [i for i in range(min(len(old), len(new))) if old[i] != new[i]]
    if not d:
        return ("palette-only" if pal_diff else "same"), {}
    cl = clusters(d)
    info = {
        "diff_px": len(d),
        "clusters": [{"x": [c[0], c[1]], "y": [c[2], c[3]], "px": c[4]} for c in cl],
    }
    # 游標是 13×16 的實心方塊；新舊各畫一個就是兩群。
    def is_cursor(c):
        return (c[1] - c[0] + 1) <= CURSOR_W and (c[3] - c[2] + 1) <= CURSOR_H
    if len(d) <= 2 * CURSOR_W * CURSOR_H and covered_by_cursor_boxes(d):
        return "cursor-only", info
    if len(d) <= 64 and len(cl) == 1 and is_cursor(cl[0]):
        return "tiny", info
    # 版面比對：非零像素的分布（哪些列有東西）幾乎相同 ＝ 同一個版面。
    if len(d) < W * H * 0.05:
        ro = {i // W for i in range(len(old)) if old[i]}
        rn = {i // W for i in range(len(new)) if new[i]}
        if ro and len(ro ^ rn) <= 4:
            return "layout-same", info
    return "needs-eyes", info


def main():
    rc = json.loads(CHECK.read_text(encoding="utf-8"))
    bad = [r for r in rc["results"] if r["screen_diff"] != 0 or r["pal_diff"] != 0]
    out, tally = [], collections.Counter()
    for r in bad:
        n, o = NEW / f"{r['name']}.bin", OLD / r["name"] / "screen.bin"
        if not n.exists() or not o.exists():
            tally["missing"] += 1
            out.append({"name": r["name"], "class": "missing"})
            continue
        kind, info = classify(o.read_bytes(), n.read_bytes(), r["pal_diff"])
        tally[kind] += 1
        out.append({"name": r["name"], "class": kind, **info})
    out.sort(key=lambda x: (x["class"], -x.get("diff_px", 0)))
    OUT.write_text(json.dumps(
        {"note": "新舊畫面對不上的收據分類；判準見 tools/scripts/triage_receipts.py",
         "source": "workplace/receipt-check.json", "count": len(out), "items": out},
        ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"對不上的 {len(bad)} 份：")
    for k in ["cursor-only", "tiny", "palette-only", "layout-same", "needs-eyes", "missing"]:
        if tally[k]:
            print(f"  {k:14} {tally[k]:4}")
    print(f"\n寫出 {OUT}")


if __name__ == "__main__":
    main()
