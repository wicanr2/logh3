"""核對固定原版雙跑；獨立驗計數、報告路由與無攻擊分支，不猜補傷害。"""
from pathlib import Path
import base64
import collections
import hashlib
import json


def sha(b):
    return hashlib.sha256(b).hexdigest()


def word(b, offset):
    return int.from_bytes(b[offset:offset + 2], "little")


def source(path):
    p = Path(path)
    b = p.read_bytes()
    return {"path": path, "size": len(b), "sha256": sha(b)}


a = Path("workplace/finish79-r89-a.json").read_bytes()
b = Path("workplace/finish79-r89-b.json").read_bytes()
assert a == b
assert sha(a) == "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477"
native = json.loads(a)
assert native["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
assert native["uid"] == 1000 and not native["dosgolem_dirty"]
rows = native["rows"]
assert len(rows) == 60
cases = []
all_calls = collections.Counter()
full_tables = 0
frames = []
for row in rows:
    before = {k: bytes.fromhex(v) for k, v in row["before"].items()}
    after = {k: bytes.fromhex(v) for k, v in row["after"].items()}
    counts = [0, 0]
    fleet_ids = before["lists"][4:4 + before["lists"][0]]
    assert list(fleet_ids) == [12, 16, 17, 28, 29, 30, 31]
    for index in fleet_ids:
        f = before["fleet"][index * 122:(index + 1) * 122]
        assert f[1] in (0, 1)
        if f[0] & 1 == 0:
            counts[f[1]] = (counts[f[1]] + 1) & 255
    lost = [i for i, count in enumerate(counts) if count == 0]
    assert row["ax"] == int(bool(lost)), row["name"]
    winner = None if len(lost) != 1 else 1 - lost[0]
    calls = row["calls"]
    c = collections.Counter(x["location"] for x in calls)
    all_calls.update(c)
    assert [x["args"][0] & 255 for x in calls if x["location"] == "185C:0A80"] == lost
    assert c["185C:0AF9"] == c["20B4:1512"] == int(winner is not None)
    if winner is not None:
        arg = next(x["args"] for x in calls if x["location"] == "20B4:1512")
        assert arg[:3] == [word(before["work"], 3), 0x0B18, 0x185C]
        assert arg[3] & 255 == winner and arg[4] == row["round"]
    reports = [x for x in calls if x["location"] == "3826:0405"]
    loss_reports = []
    for seat in lost:
        if row["human"] & (1 << seat):
            loss_reports.append(([0x3EA5, 0x54B3, seat, 0], "整個艦隊已經失去戰鬥意志。\n無法繼續進行戰鬥。\n"))
        elif row["human"] & (1 << (1 - seat)):
            loss_reports.append(([0x3ED4, 0x54B3, 1 - seat, 0], "敵方此艦隊已經無法戰鬥。\n"))
    for actual, (args, text) in zip(reports, loss_reports):
        assert actual["args"][:4] == args
        assert bytes.fromhex(actual["text_hex"]).decode("big5") == text
    captures = [x for x in calls if x["location"] == "185C:0B18"]
    assert len(reports) == len(loss_reports) + len(captures)
    for capture, report in zip(captures, reports[len(loss_reports):]):
        args = capture["args"]
        offset = args[0] - 0xC074
        assert args[1] == 0x54B3 and offset % 122 == 0
        f = after["fleet"][offset:offset + 122]
        seat = 65535 if row["human"] & (1 << f[1]) else 1 - f[1]
        assert report["args"][2:4] == [seat, word(f, 4)]
    for guard in (x for x in calls if x["location"] == "17A0:00AC"):
        assert guard["return_ax"] == 0
    exact = c["20B4:1662"] == 0
    expected = dict(before)
    if exact:
        if winner is not None:
            # 本批無攻擊勝方分支只會遇到回合6、8、9；保留do-while先回復再判終止。
            assert row["round"] in (6, 8, 9)
            recoveries = max(1, 9 - row["round"])
            assert c["20B4:15E1"] == recoveries
            p = bytearray(before["planet"])
            region = before["work"][3]
            n = word(p, region * 2)
            first = sum(word(p, i * 2) for i in range(region))
            for _ in range(recoveries):
                for pi in range(first, first + n):
                    off = 64 + pi * 82
                    max_def, rate = word(p, off + 22), word(p, off + 24)
                    d = (word(p, off + 20) + (max_def * rate // 100 & 65535)) & 65535
                    p[off + 20:off + 22] = min(d, max_def).to_bytes(2, "little")
            expected["planet"] = bytes(p)
        else:
            assert c["20B4:15E1"] == 0
        assert expected == after, row["name"]
        assert row["seed_after_hex"] == "78563412"
        full_tables += 1
    row_frames = row["frames"] or []
    assert len(row_frames) == len(reports)
    for i, frame in enumerate(row_frames):
        png = base64.b64decode(frame["png"])
        screen = base64.b64decode(frame["screen"])
        palette = base64.b64decode(frame["palette_rgb"])
        assert png[:8] == b"\x89PNG\r\n\x1a\n" and len(screen) == 640 * 480 and len(palette) == 48
        path = f"workplace/finish79-r89-{row['name']}-report{i}.png"
        dest = Path(path)
        if dest.exists():
            assert dest.read_bytes() == png
        else:
            with dest.open("xb") as f:
                f.write(png)
        frames.append({"case": row["name"], "report": i, "step": frame["step"],
                       "png": source(path), "screen_sha256": sha(screen), "palette_sha256": sha(palette)})
    cases.append({"name": row["name"], "counts": counts, "lost": lost, "winner": winner,
                  "round": row["round"], "recovery_variant": row["recovery"], "return_ax": row["ax"],
                  "calls": dict(c), "reports": [{"args": x["args"][:4], "text_big5_hex": x["text_hex"]} for x in reports],
                  "injections": row["injections"], "inputs": row["inputs"], "steps": row["steps"],
                  "seed_after_hex": row["seed_after_hex"],
                  "before_sha256": {k: sha(v) for k, v in before.items()},
                  "after_sha256": {k: sha(v) for k, v in after.items()},
                  "validation": "完整六表及清單独立相同" if exact else "計數／報告路由相同；占領傷害及全表尚未独立驗算"})
assert full_tables == 48 and len(frames) == 44
semantics = {
    "185C:0016": "8-4返回後呼叫8-5；觀測停止185C:001B不改入口",
    "185C:09DC": "迴圈上限讀戰場艦隊數的低byte",
    "185C:09FE": "只以FLEET+00 bit0篩選可戰艦隊",
    "185C:0A0E": "依FLEET+01陣營增加byte計數器",
    "185C:0A22": "陣營0歸零先報告",
    "185C:0A32": "陣營1歸零後報告",
    "185C:0A48": "只有陣營0可戰時執行其勝方占領",
    "185C:0A60": "只有陣營1可戰時執行其勝方占領；兩者歸零不走此分支",
    "185C:0A77": "任一陣營歸零返回AX=1",
    "185C:0A9B": "自己人類座位收到DS:3EA5，頭像引數0",
    "185C:0ACD": "敗方CPU但對方人類收到DS:3ED4，頭像引數0",
    "185C:0AFC": "勝方占領传入當下DS:8F1A，不固定新九回合",
    "185C:0C46": "占領callback以自己的操控類型選FFFF或對方座位",
    "20B4:1522": "暫改當前座位為勝方，供艦隊清單生成",
    "20B4:153E": "清單生成後恢復先前座位",
    "20B4:1569": "局部回合使用傳入回合低byte",
    "20B4:157F": "初始三次等待步先恢復行星防衛力",
    "20B4:1592": "已屬勝方的行星跳過，不加回合",
    "20B4:15C0": "攻占成功重設三次等待步",
    "20B4:15CC": "do-while在步末檢查局部回合>=9",
}
ida = []
for seg in ("185C", "20B4"):
    export = json.loads(Path(f"workplace/ida/finish79-r89-{seg}.json").read_bytes())
    assert export["kernel_version"] == "9.4" and export["uid"] == 1000
    for ins in export["instructions"]:
        if ins["original_location"] in semantics:
            ins.update(grade="proven", semantic=semantics[ins["original_location"]], warning="",
                       evidence=[f"workplace/ida/finish79-r89-{seg}.json", "workplace/finish79-r89-a.json"])
    ida.append(export)
assert sum(x["grade"] == "proven" for export in ida for x in export["instructions"]) == len(semantics)
paths = ["org_game/extracted/GIN3.COM", "workplace/oracle/cleanup79-r88-entry-a/entry.state",
         "workplace/oracle/cleanup79-r88-entry-a/mem-in.bin", "workplace/cleanup79-r88-native.go",
         "workplace/finish79-r89-native.go", "workplace/finish79-r89-a.json", "workplace/finish79-r89-b.json",
         "workplace/finish79-r89-map-explore-native.go", "workplace/finish79-r89-map-explore-a.json",
         "workplace/finish79-r89-map-explore-b.json", "workplace/finish79-r89-base-native.go",
         "workplace/finish79-r89-base-a.json", "workplace/finish79-r89-base-b.json",
         "workplace/ida/finish79-r89-185C.com", "workplace/ida/finish79-r89-20B4.com",
         "workplace/ida/finish79-r89-ida.py", "workplace/ida/finish79-r89-185C.json",
         "workplace/ida/finish79-r89-20B4.json", "tools/scripts/battle_promotion_ida.py",
         "tools/scripts/battle_finish_ida_prepare.py", "tools/scripts/battle_finish_native_prepare.py",
         "tools/battle_finish_oracle.sh", "tools/battle_finish_native_run.sh", __file__,
         "internal/game/battlefinal.go", "internal/rules/strategicassault.go"]
result = {"schema": "logh3.battle-finish-review.v1", "issue": 79, "grade": "proven limited scope",
          "scope": "60案固定8-1父保存自然至8-5返回；獨立計數、報告路由、48案完整六表及清單；非正式remake驗收。",
          "limits": ["12案占領傷害、完整原池回寫及RND未独立驗算", "守軍42次為零不推廣所有局面",
                     "正常玩家勝負、正式報告畫面及保存未驗", "110與完整062保持DRAFT"],
          "native": {k: v for k, v in native.items() if k != "rows"}, "parent_remake_head": "a97ccd7c97b3ca2597e41318c3c38d4ab352b945",
          "totals": {"cases": 60, "full_tables_independent": full_tables, "report_frames": len(frames), "calls": dict(all_calls)},
          "sources": [source(p) for p in paths], "ida": ida, "cases": cases, "frames": frames}
dest = Path("docs/internals/data/battle-finish.json")
with dest.open("x") as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
    f.write("\n")
print("60案雙跑；48案完整六表及清單、44原版報告；凍結SHA-256", sha(dest.read_bytes()))
