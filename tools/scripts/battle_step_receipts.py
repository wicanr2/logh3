"""研究372：核對原版雙跑、注入與IDA定位，產生可重生的分級事實。"""
import hashlib
import json
from pathlib import Path

ROOT = Path("/work")
WORK = ROOT / "workplace"
OUT = ROOT / "docs/internals/data/battle-step.json"
sha = lambda b: hashlib.sha256(b).hexdigest()
raw = (WORK / "battle79-step-extended-a.json").read_bytes()
assert raw == (WORK / "battle79-step-extended-b.json").read_bytes()
assert sha(raw) == "4cbe15151b3abcdc0da7d7a3b115ac22c600c44e31314e3f15efba9247d3395c"
receipt = json.loads(raw)
plan_raw = (WORK / "battle79-step-plan-extended.json").read_bytes()
assert sha(plan_raw) == receipt["plan_sha256"] == "d14d311df1efd314bc671c922a73b56d704bbdb1f37da1b1c9f0ab2666bb3f1f"
plan = json.loads(plan_raw)
assert receipt["plan"] == plan and len(receipt["rows"]) == len(plan["cases"]) == 102
memory = (WORK / "oracle/battle79-move-entry-a/mem-in.bin").read_bytes()
assert sha(memory) == receipt["input_sha256"] == "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d"
assert receipt["program_sha256"] == "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918"
assert receipt["dosgolem_head"] == "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
assert len({r["name"] for r in receipt["rows"]}) == 102
for row, case in zip(receipt["rows"], plan["cases"]):
    assert row["name"] == case["name"] and row["seed_after_hex"] == "78563412"
    state = bytearray(memory)
    for injection in plan["common"] + (case["injections"] or []):
        at, data = injection["linear"], bytes.fromhex(injection["hex"])
        assert 0 <= at < at + len(data) <= 1 << 20
        state[at:at + len(data)] = data
    for observation in plan["observations"]:
        at, size = observation["linear"], observation["length"]
        assert state[at:at + size].hex() == row["before"][observation["label"]]
        assert len(bytes.fromhex(row["after"][observation["label"]])) == size
    assert len(row["returns"]) == len(case["calls"]) == 1

review = {
    "1C03:006A": "單次艦隊分派：先更新一般分隊+16 bit1，再依8.0額度及完整Action 00/01/10/11決定是否進入015C。",
    "1C03:015C": "依+48 bit4選一般0185或短程0361。",
    "1C03:0185": "清+49 bit0；未抵達先轉向並保留處理返回值，固定8像素候選按實際距離先扣額度；地形FFFF才避障。",
    "1C03:0361": "清+49 bit0；保持朝向短程平移，抵達後轉向；地形FFFF才避障，先扣實際距離。",
    "1C03:0524": "最短角差在回轉速度內免費轉到位，否則有8.0額度才轉一步並扣8.0；返回AL表示有處理。",
    "2583:040C": "通行檢查：地形阻擋回FFFF，敵分隊碰撞回池索引+1，暢通回0。",
    "2583:0451": "從候選方位起依+0/-0/+16/-16至+128/-128搜尋可走落點，全阻恢復舊座標。",
    "17A0:06D7": "掃戰場艦艇清單，略過同陣營、種類7及Action bit7；碰撞回池索引+1，無碰撞回FFFF。",
    "17A0:076B": "兩軸無號座標差均嚴格小於16才判重疊。",
}
specific = {
    "1C03:00B3": ("26 80 24 fd", "一般分隊+16 bit1先清，後續Action=7F且目前不在集結點才設起。"),
    "1C03:0105": ("7d 09", "額度不足8.0即返回0，抵達後免費轉向也受外層門檻限制。"),
    "1C03:0194": ("26 80 67 49 fe", "實際進入一般步進才清+49 bit0，其他位元保留。"),
    "1C03:020F": ("88 46 03", "保存轉向AL；之後額度不足仍以此回傳，不可誤回0。"),
    "1C03:021E": ("26 08 47 49", "未抵達的轉向返回1才設艦隊+49 bit0。"),
    "1C03:0324": ("83 7e 00 ff", "艦隊只對地形FFFF避障；一般碰撞停座標，已扣額度不退還。"),
    "1C03:0350": ("9a 51 04 83 25", "一般艦隊避障參數為舊X/Y及同兩格的可寫指標；候選差為零向量。"),
    "1C03:0370": ("26 80 67 49 fe", "短程步進也清+49 bit0。"),
    "1C03:0513": ("9a 51 04 83 25", "短程艦隊避障同樣用舊座標作候選，零向量的原版角度為64。"),
    "2583:041C": ("9a ea 04 88 15", "先問地形1588:04EA，再問敵分隊碰撞。"),
    "2583:0471": ("9a 1d 01 e1 2b", "避障方位由候選-舊座標計算。"),
    "2583:050E": ("80 46 02 10", "每組正負搜尋增加16角度。"),
    "17A0:0750": ("40", "碰撞池索引加1，避免池0碰撞被誤當暢通。"),
    "17A0:07A6": ("73 04", "X差>=16不重疊。"),
    "17A0:07AA": ("72 04", "Y差<16且X差<16才重疊。"),
}
inputs, exports, entries = {}, {}, []
for name in ["re79-movement-1C03.json", "re79-movement-r30-1C03.json", "re79-movement-r30-2583.json",
             "re79-movement-r30-17A0.json", "re79-movement-r30-overlap.json"]:
    p = WORK / "ida" / name
    j = json.loads(p.read_text())
    assert j["tool"] == "IDA Pro 9.4" and j["ida_base_hex"] == "0x10200"
    digest = sha(p.read_bytes())
    exports[name] = {"path": str(p.relative_to(ROOT)), "sha256": digest, "input_sha256": j["input_sha256"]}
    for f in j["functions"]:
        loc = f["original_location"]
        if loc not in review:
            continue
        inputs[loc.split(":")[0]] = j["input_sha256"]
        assert f["instructions"][0]["original_bytes_hex"] == "55"
        assert f["instructions"][-1]["original_operand"] == "retf"
        entries.append({"kind": "function", "original_location": loc, "semantic": review[loc],
                        "grade": "proven", "source": ["docs/re/372-battle-step-and-collision.md", digest]})
        for i in f["instructions"]:
            if i["original_location"] in specific:
                expected, semantic = specific[i["original_location"]]
                assert i["original_bytes_hex"] == expected
                entries.append({"kind":"instruction", "original_location":i["original_location"],
                                "original_bytes_hex":expected, "semantic":semantic,
                                "grade":"proven", "source":["docs/re/372-battle-step-and-collision.md",digest]})
assert {e["original_location"] for e in entries if e["kind"] == "function"} == review.keys()
assert {e["original_location"] for e in entries if e["kind"] == "instruction"} == specific.keys()
result = {"schema":"logh3.battle-step-evidence.v1", "inputs":inputs, "entries":entries, "exports":exports,
          "tool":"IDA Pro 9.4", "ida_base_hex":"0x10200", "routine_receipt_sha256":sha(raw),
          "receipt":receipt, "scope":"102案原版雙跑。艦隊單步與通行檢查的remake對拍另由規格081測試；分隊與整階段不能據此宣稱完成。"}
assert OUT.parent.stat().st_uid == 1000 and (not OUT.exists() or OUT.stat().st_uid == 1000)
OUT.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print("102案原版雙跑／注入／seed、9函式／15指令分級核對通過；摘要SHA-256",sha(OUT.read_bytes()))
