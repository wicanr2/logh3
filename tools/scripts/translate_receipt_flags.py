"""把收據的舊 probe 旗標翻成 dosgolem main 的新語法。

    tools/py.sh tools/scripts/translate_receipt_flags.py <批次名> [更多批次...]
    tools/py.sh tools/scripts/translate_receipt_flags.py --all

舊收據綁的是 `logh3-com-support`（`ef5be47`）。main 那邊改了三處語法：

    -clicks   x,y,at[,hold] 用 `;` 分隔   →   步數:X:Y[:鍵] 用 `,` 分隔
    -cmdline  <字串>                      →   -args <字串>
    -dump-mem lin:<位址hex>:<長度dec>:<路徑>  →  <lo>-<hi>:<路徑>（都是 hex）

**滑鼠座標不要動。** 兩版的水平倍率都是「依視訊模式自動判定」，mode 12h
一律 1，收據上的 X 就是遊戲看到的 X（`docs/re/270`）。main 有一支
`-xscale` 可以強制倍率，設成 2 會讓遊戲收到兩倍的 X 而走上另一條路——
畫面照樣畫得出來，所以它看起來很像「修好了」。

`-rclicks` 還在、格式與 `-clicks` 相同；`-steps` 仍然是「跑到第幾道指令」，
不用動。**每個 click 各自的 hold 在新版沒有對應**——新版只有全域的
`-click-hold`，所以一份收據裡的 hold 不一致時這支會拒絕翻譯並說出來，
不會自己挑一個。

⚠ **第四處差異這支沒有翻：每次點擊的 20 萬步前移。** 舊 probe
（`ef5be47:cmd/probe/main.go` 的 `const moveLead = 200_000`）在每一次腳本
點擊的 `at − 200000` 送一次 `EvMove`，而且座標沒變也照送；新版是同一道
指令裡 `MoveMouse` ＋ `PressMouse`，而 `MoveMouse` 座標沒變就直接 return。
**遊戲分得出來**：`s3cmdfl3` 少了那次前移之後，艦隊選單照樣開，右側卻
不換成分隊面板（`docs/re/282`）。要補回來有兩條路——`-click-premove`
（依**輪詢次數**，會延到按下，而且套用到每一次點擊），或
`-mouse-x/-mouse-y/-mouse-at`（只補得了最後那一次，但步數是精確的）。

自動補等於改掉全部 455 份的旗標、要連帶重錄一遍，所以這支**維持原樣**；
改成自動補是父層決策，先問使用者。為了不讓手補過的旗標被下一次
`--all` 無聲洗掉，已經存在而且內容不同的 `flags-new/<批次>.txt`
**不覆蓋**，要蓋掉加 `--force`。

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
            return regroup(line[len("probe 旗標: "):].split()), "meta.txt（重組含空白的值）"
    return None, None


def is_flag(tok):
    """`-x` 開頭且第二個字元是字母才算旗標名。

    `-1`（`-mouse-x -1` 的預設值）第二個字元是數字，是**值**不是旗標。
    """
    return len(tok) > 1 and tok[0] == "-" and tok[1].isalpha()


def regroup(toks):
    """把按空白拆開的旗標重新組回去。

    `meta.txt` 那一行是給人看的，值含空白的（`-poke lin:...=03 02 01 04`、
    `-cmdline " 9 2"`）拆完就散成好幾個 token。**散掉不會報錯**：probe 收到
    多餘的位置參數會沉默地忽略，於是那個 `-poke` 等於沒下，收據跑出另一個
    局面而旗標看起來完全正常。
    """
    out, cur = [], None
    for t in toks:
        if is_flag(t):
            if cur is not None:
                out.append(cur)
                cur = None
            out.append(t)
        else:
            cur = t if cur is None else cur + " " + t
    if cur is not None:
        out.append(cur)
    return out


def translate_dump_mem(v):
    """lin:<位址hex>:<長度dec>:<路徑> → <lo>-<hi>:<路徑>（逗號分隔多段）"""
    out = []
    for part in v.split(","):
        part = part.strip()
        if not part:
            continue
        if not part.startswith("lin:"):
            out.append(part)
            continue
        _, addr, ln, path = part.split(":", 3)
        lo = int(addr, 16)
        # ⚠ 舊語法的位址是十六進位、長度是**十進位**（AGENTS.md）。
        # 兩個都當成十六進位讀會安靜地換掉範圍大小。
        hi = lo + int(ln, 10) - 1
        out.append("%X-%X:%s" % (lo, hi, path))
    return ",".join(out)


def translate_clicks(v):
    """x,y,at[,hold];… → (步數:X:Y[::按住],…, holds)

    hold 全部一樣時提成全域的 `-click-hold`（旗標比較短，也與已翻好的
    那幾百份一致）；**不一樣時逐次寫進第五欄**——同一段序列裡不同的鈕
    要不同的按住時間是常態，硬挑一個值會讓收據跑出另一個畫面。
    """
    rows, holds = [], set()
    for part in v.split(";"):
        part = part.strip()
        if not part:
            continue
        f = part.split(",")
        if len(f) < 3:
            raise ValueError("clicks 少了欄位：%r" % part)
        x, y, at = f[0], f[1], f[2]
        hold = f[3] if len(f) > 3 else ""
        if hold:
            holds.add(hold)
        rows.append((at, x, y, hold))
    if len(holds) > 1:
        # 第四欄（鍵）留空 ＝ 用預設的左鍵；-rclicks 那邊照樣把它改成右鍵。
        return ",".join("%s:%s:%s::%s" % r for r in rows), set()
    return ",".join("%s:%s:%s" % r[:3] for r in rows), holds


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
        if a == "-dump-mem" and nxt is not None and "lin:" in nxt:
            out += [a, translate_dump_mem(nxt)]
            notes.append("-dump-mem 改成 <lo>-<hi>:<路徑>")
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
    force = "--force" in args
    args = [a for a in args if a != "--force"]
    if args == ["--all"]:
        args = sorted(p.name for p in ORACLE.iterdir() if p.is_dir())
    OUT.mkdir(parents=True, exist_ok=True)
    ok = bad = skip = kept = 0
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
        # ⚠ **手改過的旗標不覆蓋。** 翻譯翻不出「每次點擊的 20 萬步前移」
        # （見開頭），有幾份收據靠手加的 `-mouse-at`／`-click-premove` 才
        # 停得到該停的畫面。無聲覆蓋掉的話，下一次重錄會錄成另一個狀態，
        # 而旗標檔看起來完全正常。
        out_path = OUT / (batch + ".txt")
        text = "\n".join(new) + "\n"
        if out_path.exists() and out_path.read_text(encoding="utf-8") != text and not force:
            print("%-28s 旗標與翻譯結果不同（手改過？）→ 保留不覆蓋，要蓋掉加 --force" % batch)
            kept += 1
            continue
        out_path.write_text(text, encoding="utf-8")
        ok += 1
        if len(args) <= 5:
            print("%-28s ← %s" % (batch, src))
            for n in notes:
                print("    ·", n)
            print("    " + " ".join(new))
    print("\n翻好 %d 份、保留手改的 %d 份、%d 份要人處理、%d 份沒有旗標"
          % (ok, kept, bad, skip))


main()
