"""第37輪：原版比較器與43邊界分隊物件鍵；原版輸入只讀。"""
import hashlib
import json
from pathlib import Path
import struct


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


frames_path = Path("docs/internals/data/battle-second-frames.json")
data = json.loads(frames_path.read_text())
root = Path("workplace/oracle/battle79-second-r36-a")
mem_path = root / "entry-mem.bin"
assert sha(mem_path) == "9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec"
mem = mem_path.read_bytes()
p = Path("workplace/ida/re79-render-r37-35D1.json")
export = json.loads(p.read_text())
assert export["tool"] == "IDA Pro 9.4"
assert export["input_sha256"] == sha(p.with_suffix(".com"))
semantics = {
    "35D1:02B4": "第一物件比較鍵讀unsigned byte +3",
    "35D1:02BE": "第二物件比較鍵讀unsigned byte +3",
    "35D1:02C4": "第一鍵unsigned較小時回傳-1，較小先畫",
    "35D1:02E1": "第一鍵unsigned較大時回傳1，較大後畫",
    "35D1:02EE": "相同第一鍵時讀第一物件word +4",
    "35D1:02F8": "相同第一鍵時讀第二物件word +4",
    "35D1:02FE": "第二鍵unsigned較大時回傳-1，較大先畫",
    "35D1:031B": "第二鍵unsigned較小時回傳1，較小後畫",
    "35D1:0322": "兩鍵相同時回傳0；通用排序器的相等鍵置換未驗",
}
entries = []
for f in export["functions"]:
    for r in f["instructions"]:
        loc = r["original_location"]
        if loc not in semantics:
            continue
        addr = int(r["runtime_linear_hex"], 16)
        raw = bytes.fromhex(r["original_bytes_hex"])
        assert mem[addr:addr+len(raw)] == raw
        entries.append({**r, "semantic": semantics[loc], "grade": "proven", "warning": "",
                        "source": ["docs/re/377-battle-render-order.md", sha(p)]})
assert len(entries) == len(semantics)
rows = []
for frame in data["frames"]:
    mp = root / (frame["name"] + "-mem.bin")
    assert sha(mp) == frame["hashes"]["-mem.bin"]
    m = mp.read_bytes()
    n = struct.unpack_from("<H", m, 0x54b30+0xeade)[0]
    objs = []
    for i in range(n):
        off, seg = struct.unpack_from("<HH", m, 0x54b30+0xeae0+4*i)
        a = seg*16+off
        objs.append((m[a+3], struct.unpack_from("<H", m, a+4)[0]))
    assert n == 153 and objs == sorted(objs, key=lambda r: (r[0], -r[1]))
    keys = [key for priority, key in objs if priority == 128]
    assert len(keys) == len(set(keys)) == 135
    wanted = sorted([m[0x668b2+r["index"]*78+1] | m[0x668b2+r["index"]*78+3]<<8
                     for r in frame["units"]], reverse=True)
    assert keys == wanted
    rows.append({"name": frame["name"], "objects": n, "unit_priority": 128,
                 "keys": keys, "ram_sha256": sha(mp)})
report = {"schema": "logh3.semantic-index.v1", "tool": "IDA Pro 9.4",
          "program_sha256": data["program_sha256"], "dosgolem_head": data["dosgolem_head"],
          "input_sha256": sha(mem_path), "cut_sha256": export["input_sha256"],
          "export_sha256": sha(p), "address_space": export["address_space"],
          "frame_source_sha256": sha(frames_path), "entries": entries, "frames": rows,
          "unknown": ["y203寫入來源與S1/S3差異", "任意方向物件相等鍵的置換效果"]}
out = Path("docs/internals/data/battle-draw-order.json")
out.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n")
print("9筆原始指令已分級，43邊界各153物件比較鍵及135唯一分隊鍵相符", sha(out))
