"""第36輪方向圖示原始bytes、位移與分級索引。"""
import hashlib
import json
from pathlib import Path
import struct
import sys

def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

mem_path = Path("workplace/oracle/battle79-second-r36-a/entry-mem.bin")
mem = mem_path.read_bytes()
assert sha(mem_path) == "9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec"
assert mem[0x401be:0x405be] == Path("/orig/SPRITE.SPR").read_bytes()[:1024]
if sys.argv[1:] == ["prepare"]:
    for segment in ("25DB", "3982", "35D1"):
        base = Path("workplace/ida") / ("re79-second-r36-"+segment)
        p = base.with_suffix(".com")
        assert not p.exists() and not base.with_suffix(".com.i64").exists()
        p.write_bytes(bytes(256)+mem[int(segment,16)*16:int(segment,16)*16+4096])
        base.with_suffix(".index.json").write_text(json.dumps({"inputs": {segment: sha(p)}, "entries": []}, indent=2)+"\n")
    print("三份新一次性切段與unknown索引已建立")
    sys.exit(0)
semantics = {
    "25DB:0408": "完整顯示旗標非0時進完整圖示分支",
    "25DB:040C": "否則檢查SHIP+17是否已達3",
    "25DB:0413": "完整顯示分支檢查SHIP+3是否為旗艦序號0",
    "25DB:0438": "旗艦呼叫04E7且旗標1，啟用方向",
    "25DB:0487": "非完整顯示旗艦呼叫04E7且旗標0，隱藏方向",
    "25DB:04F9": "將Fleet+4A送入八向索引",
    "25DB:0503": "方向索引存Fleet+6A",
    "25DB:050C": "隱藏時Fleet+6A為9",
    "25DB:06D4": "世界X減鏡頭X後加DS:4A3C方向X位移",
    "25DB:06FF": "世界Y減鏡頭Y後加DS:4A3C方向Y位移",
    "25DB:08D3": "以Fleet+6A取得方向圖塊指標",
    "3982:000F": "方向圖塊索引左移7，每塊128 bytes",
    "3982:0011": "圖塊池段值來自DS:F12E",
    "3982:0015": "圖塊池位移來自DS:F12C",
    "3982:002F": "Facing右移4",
    "3982:0031": "方向中點取整先加1",
    "3982:0032": "再右移1",
    "3982:0034": "八方向索引以7遮罩",
    "35D1:0053": "加入物件後呼叫028D，排序語意仍unknown",
}
rows, exports = [], []
for segment in ("25DB", "3982", "35D1"):
    p = Path("workplace/ida") / ("re79-second-r36-" + segment + ".json")
    e = json.loads(p.read_text())
    assert e["tool"] == "IDA Pro 9.4"
    exports.append({"path": str(p), "sha256": sha(p), "input_sha256": e["input_sha256"]})
    for f in e["functions"]:
        for r in f["instructions"]:
            loc = r["original_location"]
            if loc not in semantics: continue
            raw = bytes.fromhex(r["original_bytes_hex"])
            linear = int(segment, 16)*16 + int(loc.split(":")[1], 16)
            assert mem[linear:linear+len(raw)] == raw
            rows.append({**r, "semantic": semantics[loc], "grade": "proven", "warning": "",
                         "source": ["docs/re/376-second-battle-render.md", sha(p)]})
assert len(rows) == len(semantics)
data = {"schema": "logh3.semantic-index.v1", "tool": "IDA Pro 9.4",
        "program_sha256": "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918",
        "ram_sha256": sha(mem_path), "ida_base_hex": "0x10200",
        "address_space": "cut offset0=IDA 0x10200; original CS:IP and dosgolem linear separate",
        "exports": exports, "entries": rows,
        "direction_offsets": [list(struct.unpack_from("<hh", mem, 0x54b30+0x4a3c+4*i)) for i in range(8)],
        "direction_tiles": {"first": 0, "count": 8, "bytes_per_tile": 128,
                            "raw_sha256": hashlib.sha256(mem[0x401be:0x405be]).hexdigest(),
                            "source_file": "SPRITE.SPR", "source_sha256": sha(Path("/orig/SPRITE.SPR")), "grade": "proven"},
        "unknown": ["同順位重疊物件的實際繪製次序", "S1與S3的y203格表殘留差異"]}
out = Path("docs/internals/data/battle-directions.json")
out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
print("19筆原始指令分級、八向位移及1024 bytes素材已核，JSON", sha(out))
