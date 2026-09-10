#!/usr/bin/env python3
"""KEIREKI.DAT（人物傳）解碼，並與資料段裡的提督姓名表對起來。

KEIREKI.DAT 版面（proven）：
  offset 0 起是 uint16 LE 位移表；第一筆位移本身就是表的長度，
  所以筆數 = offsets[0] // 2 = 83。最後一筆等於檔案長度，當終止哨兵。
  兩個位移之間就是一段 Big5（倚天）文字，段內不含 \\0，換行由遊戲自己排版。
  第 82 筆長度為 0（offsets[82] == offsets[83]），所以實際有內文的是 82 筆，
  正好對上 ADMIRAL.DT* 的 82 名提督。

姓名不在任何一個外部檔案裡——GIN3PS.EXE 是 TINYPROG 打包過的，姓名表在
解包後的 64 KB 資料段內，`\\0` 分隔，第 0 筆「萊因哈特」在 0x61A8。
本腳本吃 workplace/strtab.bin（該資料段的傾印，取得方法見 docs/internals/14）。
82 個提督名之後接著一張指標表，再往下是 34 個旗艦名（0x6570 起）。
"""
import json, os, struct

SRC = "/orig/KEIREKI.DAT"
ADMIRAL = "/orig/ADMIRAL.DTX"  # +1 ＝ 陣營旗標（0 帝國／1 同盟），confirmed
STRTAB = os.environ.get("LOGH3_STRTAB", "/work/workplace/strtab.bin")
NAME_OFF, NAME_N = 0x61A8, 82
SHIP_OFF, SHIP_N = 0x6570, 34
OUT_JSON = "/work/docs/internals/data/keireki.json"
OUT_MD = "/work/docs/lore/character-biographies.md"
# 頭像：FACE.AMG 第 n 張 ＝ 提督索引 n（docs/internals/09、docs/re/271）。
# 這份 JSON 由 amg_export.py 產生，跑本腳本之前要先跑它。
FACE_JSON = "/work/docs/internals/data/face-amg.json"


def read_cstrings(d, off, n, skip_lead=True):
    """從 off 起讀 n 個 \\0 結尾的 Big5 字串，回 (字串, 原始位元組)。"""
    out, p = [], off
    while len(out) < n:
        e = d.index(b"\x00", p)
        raw = d[p:e]
        out.append((raw.decode("big5hkscs", errors="replace"), raw))
        p = e + 1
    return out


def big5_codes(raw):
    """把一段 Big5 位元組兩兩併成 uint16——畫面用的是字碼，不是字串。"""
    return [raw[k] << 8 | raw[k + 1] for k in range(0, len(raw) - 1, 2)]


data = open(SRC, "rb").read()
n = struct.unpack_from("<H", data, 0)[0] // 2
offs = list(struct.unpack_from("<%dH" % n, data, 0))

names, ships = [], []
if os.path.exists(STRTAB):
    st = open(STRTAB, "rb").read()
    names = read_cstrings(st, NAME_OFF, NAME_N)
    ships = [t for t, _ in read_cstrings(st, SHIP_OFF, SHIP_N)]

side = []
if os.path.exists(ADMIRAL):
    ad = open(ADMIRAL, "rb").read()
    side = [ad[i * 28 + 1] for i in range(len(ad) // 28)]

recs = []
for i in range(n - 1):
    raw = data[offs[i]:offs[i + 1]]
    if not raw:
        continue
    # 原版把每段補到固定長度，尾巴的全形空白是排版用的，不是內文。
    body = raw.rstrip(b"\x00")
    txt = body.decode("big5hkscs", errors="replace").rstrip("\u3000 ")
    # 畫面用的是**字碼**，不是 UTF-8 字串——把去掉排版填充之後那一段
    # 原始 Big5 位元組兩兩併成 uint16 存起來（`docs/spec/012`）。
    trimmed = body[:len(body)]
    while trimmed.endswith(b"\xa1@") or trimmed.endswith(b" "):
        trimmed = trimmed[:-2] if trimmed.endswith(b"\xa1@") else trimmed[:-1]
    codes = [trimmed[k] << 8 | trimmed[k + 1] for k in range(0, len(trimmed) - 1, 2)]
    recs.append({
        "index": i,
        "name": names[i][0] if i < len(names) else "",
        "name_codes": big5_codes(names[i][1]) if i < len(names) else [],
        "side": int(side[i]) if i < len(side) else None,
        "offset": offs[i],
        "bytes": len(raw),
        "text": txt,
        "text_codes": codes,
    })

os.makedirs(os.path.dirname(OUT_JSON), exist_ok=True)
json.dump({
    "source": "KEIREKI.DAT",
    "encoding": "Big5（倚天）",
    "count": len(recs),
    "names_from": {"file": "GIN3PS.EXE 資料段", "offset": NAME_OFF, "count": NAME_N},
    "ships": ships,
    "records": recs,
}, open(OUT_JSON, "w", encoding="utf-8"), ensure_ascii=False, indent=1)


def wrap(t, width=38):
    return "\n".join(t[i:i + width] for i in range(0, len(t), width))


# 頭像引用：原版素材不進 Git，所以只寫「哪一張 ＋ 解碼後像素的 SHA-256」。
# 拿的是 sha256_pixels 不是 sha256_png——像素雜湊只看解碼結果，換一個
# zlib 版本不會變；PNG 雜湊會，掛在 82 篇底下會安靜地全部過期。
faces = {}
if os.path.exists(FACE_JSON):
    for m in json.load(open(FACE_JSON))["images"]:
        if m.get("admiral") is not None:
            faces[m["admiral"]] = m["sha256_pixels"]
if len(faces) != NAME_N:
    # 少了就大聲講。安靜地少掉 82 行頭像引用，看起來會像「本來就沒有」。
    print("⚠ 只讀到 %d 筆頭像雜湊（要 %d）：先跑 amg_export.py，"
          "否則人物傳這一版不會有頭像引用" % (len(faces), NAME_N))


def face_line(i):
    h = faces.get(i)
    if not h:
        return ""
    return "頭像 `face-%02d`（像素 SHA-256 `%s`）\n\n" % (i, h)


os.makedirs(os.path.dirname(OUT_MD), exist_ok=True)
with open(OUT_MD, "w", encoding="utf-8") as f:
    f.write("""# 人物傳

《銀河英雄傳說III SP》PC DOS 中文版的 82 篇人物傳，逐字取自原版
`KEIREKI.DAT`，中譯出自微波軟體（Micronet 授權的台灣代理商）當年的
中文化工程。文字未經改寫，只補上原版由排版程式決定的換行。

姓名取自 `GIN3PS.EXE` 解包後資料段 0x61A8 起的姓名表，順序與
`ADMIRAL.DT*` 的提督索引一致；陣營取自 `ADMIRAL.DTX` 每筆的 `+1`
（0 ＝ 帝國、1 ＝ 同盟），剛好把 82 人切成 0–39 與 40–81 兩半。

| 來源 | 值 |
|---|---|
| 檔案 | `KEIREKI.DAT`，18,046 bytes |
| 編碼 | Big5（倚天） |
| 筆數 | 83 筆位移，其中 82 筆有內文 |
| 資料 | `docs/internals/data/keireki.json` |

## 頭像

每一篇底下的 `face-NN` 是這個人在 `FACE.AMG` 裡的那一張——**第 n 張就是
提督索引 n**，證據與逐項等級見 [`docs/internals/09`](../internals/09-face-amg.md)
與 [`docs/re/271`](../re/271-face-amg-index-is-the-admiral-index.md)。

原版素材不進版控，所以這裡只寫編號與雜湊。要拿到圖：自備原版之後跑

```
tools/py.sh tools/scripts/amg_export.py
```

它會把 83 張寫到 `workplace/assets/face/face-NN.png`（80 × 100、4 位元色號），
並更新 `docs/internals/data/face-amg.json`。**像素 SHA-256** 算的是解碼後
`80 × 100` 個色號位元組（`sha256sum` 不能直接對 PNG 用，PNG 的雜湊隨 zlib
版本變動；`face-amg.json` 兩種都有）。

""")
    cur = None
    for r in recs:
        if r["side"] != cur:
            cur = r["side"]
            f.write("## %s\n\n" % ("自由行星同盟軍" if cur else "銀河帝國軍"))
        f.write("### %02d %s\n\n%s%s\n\n" % (r["index"], r["name"] or "（未知）",
                                             face_line(r["index"]), r["text"]))
    if ships:
        f.write("## 旗艦名\n\n")
        f.write("資料段 0x6570 起，34 筆。\n\n")
        for i, s in enumerate(ships):
            f.write("%d. %s\n" % (i, s))
        f.write("\n")

print("人物傳 %d 篇，旗艦 %d 艘 → %s" % (len(recs), len(ships), OUT_MD))
