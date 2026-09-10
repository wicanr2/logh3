"""把收據的舊 probe 旗標翻成 dosgolem main 的新語法。

    tools/py.sh tools/scripts/translate_receipt_flags.py <批次名> [更多批次...]
    tools/py.sh tools/scripts/translate_receipt_flags.py --all

舊收據綁的是 `logh3-com-support`（`ef5be47`）。main 那邊改了兩處語法：

    -clicks   x,y,at[,hold] 用 `;` 分隔   →   步數:X:Y[:鍵] 用 `,` 分隔
    -cmdline  <字串>                      →   -args <字串>

`-rclicks` 還在、格式與 `-clicks` 相同；`-steps` 仍然是「跑到第幾道指令」，
不用動。**每個 click 各自的 hold 在新版沒有對應**——新版只有全域的
`-click-hold`，所以一份收據裡的 hold 不一致時這支會拒絕翻譯並說出來，
不會自己挑一個。

輸出寫到 `workplace/flags-new/<批次>.txt`（一行一個旗標，與 flags.txt 同格式）。
"""
import pathlib
import sys

ROOT = pathlib.Path("/work")
ORACLE = ROOT / "workplace" / "oracle"
OUT = ROOT / "workplace" / "flags-new"


def read_flags(batch):
    """回 [旗標...]。優先讀 flags.txt——meta.txt 那一行的旗標含空格會拆錯。"""
    d = ORACLE / batch
    f = d / "flags.txt"
    if f.exists():
        return [l for l in f.read_text(encoding="utf-8").split("\n") if l != ""], "flags.txt"
    m = d / "meta.txt"
    if not m.exists():
        return None, None
    for line in m.read_text(encoding="utf-8").split("\n"):
        if line.startswith("probe 旗標: "):
            # 只能按空白拆，所以值含空白的（-cmdline " 9 2"）會壞——標出來。
            return line[len("probe 旗標: "):].split(), "meta.txt（按空白拆，值含空白者不可靠）"
    return None, None


def translate_clicks(v):
    """x,y,at[,hold];… → (步數:X:Y,…, holds)"""
    out, holds = [], set()
    for part in v.split(";"):
        part = part.strip()
        if not part:
            continue
        f = part.split(",")
        if len(f) < 3:
            raise ValueError("clicks 少了欄位：%r" % part)
        x, y, at = f[0], f[1], f[2]
        if len(f) > 3:
            holds.add(f[3])
        out.append("%s:%s:%s" % (at, x, y))
    return ",".join(out), holds


def translate(flags):
    out, holds, notes = [], set(), []
    i = 0
    while i < len(flags):
        a = flags[i]
        nxt = flags[i + 1] if i + 1 < len(flags) else None
        if a in ("-clicks", "-rclicks") and nxt is not None:
            v, h = translate_clicks(nxt)
            holds |= h
            out += [a, v]
            i += 2
            continue
        if a == "-cmdline" and nxt is not None:
            notes.append("-cmdline 改名成 -args")
            out += ["-args", nxt]
            i += 2
            continue
        out.append(a)
        i += 1
    if len(holds) > 1:
        raise ValueError("同一份收據裡 click 的 hold 不一致（%s）——"
                         "新版只有全域的 -click-hold，要人決定" % "、".join(sorted(holds)))
    if holds:
        h = holds.pop()
        if "-click-hold" not in out:
            out += ["-click-hold", h]
            notes.append("每個 click 的 hold %s 提成全域 -click-hold" % h)
    return out, notes


def main():
    args = sys.argv[1:]
    if args == ["--all"]:
        args = sorted(p.name for p in ORACLE.iterdir() if p.is_dir())
    OUT.mkdir(parents=True, exist_ok=True)
    ok = bad = skip = 0
    for batch in args:
        flags, src = read_flags(batch)
        if flags is None:
            print("%-28s 沒有旗標可讀" % batch)
            skip += 1
            continue
        try:
            new, notes = translate(flags)
        except ValueError as e:
            print("%-28s ⚠ %s" % (batch, e))
            bad += 1
            continue
        (OUT / (batch + ".txt")).write_text("\n".join(new) + "\n", encoding="utf-8")
        ok += 1
        if len(args) <= 5:
            print("%-28s ← %s" % (batch, src))
            for n in notes:
                print("    ·", n)
            print("    " + " ".join(new))
    print("\n翻好 %d 份、%d 份要人處理、%d 份沒有旗標" % (ok, bad, skip))


main()
