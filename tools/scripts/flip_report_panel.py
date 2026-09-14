"""從訊息表產生戰略報告的共用版面規格（`docs/spec/034`、`046`、`047`）。

    tools/py.sh tools/scripts/flip_report_panel.py <訊息表json> <輸出json>

版面數字量自收據 `chain3`（`docs/re/246`、`248`）；文字取自
`strategy-report-strings.json`，所以字碼與原版逐位元組相同。
"""
import json
import sys

src = json.load(open(sys.argv[1], encoding="utf-8"))
by_ds = {s["ds"]: s for s in src["strings"]}


def fullwidth(raw):
    """把一段 Big5 位元組轉成全形字碼陣列，跳過控制碼與半形。"""
    out, i = [], 0
    while i < len(raw):
        if i + 1 < len(raw) and 0xA1 <= raw[i] <= 0xF9:
            out.append(raw[i] << 8 | raw[i + 1])
            i += 2
        else:
            i += 1
    return out


def lines(ds):
    """一條字串拆成逐行的字碼。原版那幾條開頭的全形空白照留——
    它們是第 1 行的尾巴，`\\n` 之後才是報告的第 2 行。"""
    raw = bytes.fromhex(by_ds[ds]["raw"])
    parts = raw.split(b"\n")
    # 開頭那一段（全形空白）不印，砍掉。
    if len(parts) > 1 and set(parts[0]) <= {0xA1, 0x40, 0x20}:
        parts = parts[1:]
    return [fullwidth(p) for p in parts if fullwidth(p)]


def raw_from(ds):
    """從涵蓋 ds 的來源項目切到 NUL；ds:0732 落在 0722 那筆原始塊內。"""
    target = int(ds, 16)
    for s in src["strings"]:
        start = int(s["ds"], 16)
        raw = bytes.fromhex(s["raw"])
        if start <= target < start + len(raw):
            out = raw[target - start:]
            return out.split(b"\0", 1)[0]
    raise KeyError(ds)


def split_lines(raw):
    parts = raw.split(b"\n")
    if len(parts) > 1 and set(parts[0]) <= {0xA1, 0x40, 0x20}:
        parts = parts[1:]
    return [fullwidth(p) for p in parts if fullwidth(p)]


def entry(ds):
    ls = split_lines(raw_from(ds))
    return {"ds": ds, "lines": ls,
            "text": [bytes(b for c in l for b in (c >> 8, c & 0xFF)).decode("big5")
                     for l in ls]}

panel = {
    "note": ("星域易主報告的版面（`docs/spec/034`）。座標量自收據 `chain3`，"
             "文字取自 `strategy-report-strings.json`，字碼與原版逐位元組相同；"
             "叛亂、占據與提督 fate 報告共用同一版面（`docs/spec/046`、`047`、`048`）。"),
    "line_step": 20,
    # 星系名的字色跟著**新的擁有者**走，三格都是量到的：
    #   chain7 打下敵方星域 → 0（帝國）色 6
    #   chain3 被敵方打下來 → 1（同盟）色 5
    #   chain6 變成無主     → 2 色 7
    # 三份收據的玩家座位都是帝國，所以「依擁有者」與「依敵我」分不開。
    "name": {"x": 336, "y": 182, "ink_by_owner": [6, 5, 7],
             "ink_note": "chain7／chain3／chain6 各一格；座位都是帝國"},
    # `ds:085D` ＝ `\x1B[%d,%s\x1B[%d,星系`，尾巴那兩個字就是接在星系名後面的。
    "suffix": {"ink": 7, "ds": "085D",
               "codes": fullwidth(bytes.fromhex(by_ds["085D"]["raw"]))[-2:],
               "text": "星系"},
    "message": {"x": 336, "y": 202, "ink": 7},
    "face": {"x": 544, "y": 180, "w": 72, "h": 93,
             "source": "FACE.AMG",
             "index_from": "ds:8F4D[座位陣營]",
             "note": "第 n 張 ＝ 提督索引 n（docs/spec/012）；收據 chain3 逐像素差 0"},
    "messages": {
        "taken_by_us": entry("08DA"),
        "lost_to_enemy": entry("08EE"),
        "our_grip_lost": entry("088B"),
        "we_partly_landed": entry("08BC"),
        "rebellion": entry("086F"),
        "occupation_new_us": entry("0732"),
        "occupation_new_enemy": entry("0756"),
        "occupation_continuing_us": entry("0778"),
        "occupation_continuing_enemy": entry("079A"),
        "admiral_missing_title": entry("080A"),
        "admiral_name_suffix": entry("081D"),
        "admiral_dead": entry("0825"),
        "admiral_survived": entry("083A"),
    },
    "evidence": {
        "taken_by_us": "量到（訊息表；收據 chain7 打下敵方星域畫的就是這句）",
        "lost_to_enemy": "量到（收據 chain3）",
        "our_grip_lost": "量到（收據 chain6：自己的星域變混雜就成無主）",
        "we_partly_landed": "量到（收據 chain8：敵方的星域變混雜就成無主）",
        "rebellion": "量到（收據 rbx292）",
        "occupation_new_us": "量到（收據 occupation-report3；docs/re/306）",
        "occupation_new_enemy": "原始分支與訊息表（056A:059E；docs/re/306）",
        "occupation_continuing_us": "原始分支與訊息表（056A:059E；docs/re/306）",
        "occupation_continuing_enemy": "原始分支與訊息表（056A:059E；docs/re/306）",
        "admiral_missing_title": "原始組字與訊息表（056A:0311；docs/re/307）",
        "admiral_name_suffix": "原始組字與訊息表（056A:0311；docs/re/307）",
        "admiral_dead": "原始死亡分支與訊息表（056A:0311；docs/re/307）",
        "admiral_survived": "原始生還分支與訊息表（056A:0311；docs/re/307）",
    },
}
json.dump(panel, open(sys.argv[2], "w", encoding="utf-8"),
          ensure_ascii=False, indent=1)
open(sys.argv[2], "a", encoding="utf-8").write("\n")
print("寫出 %s" % sys.argv[2])
