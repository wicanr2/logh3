"""第38輪：CMP、格線、2048-byte bitmap 與尾列的可重生證據。"""
import hashlib
import json
from pathlib import Path
import struct


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def decode(b):
    out = bytearray()
    i = 0
    while i < len(b):
        c = b[i]
        i += 1
        if c >= 128:
            out.append(c & 127)
        else:
            assert i < len(b)
            out.extend([c] * b[i])
            i += 1
    assert len(out) == 128 * 102
    return out


def bitmap(cells, lookup):
    # 在有界的虛擬鄰接資料內重現讀取，沒有真的越界讀取主機記憶體。
    src = bytearray(cells)
    for i, c in enumerate(src):
        x, y = i % 128, i // 128
        if c == 0:
            if x % 15 == 0:
                src[i] = 2 if y % 12 == 0 else 4
            elif y % 12 == 0:
                src[i] = 3
    grid = bytes(src)
    src += bytes.fromhex("02 08") + bytearray(2048)
    start = len(grid) + 2
    for j in range(128):
        r = (j + 1) * 4 // 5
        for x in range(128):
            src[start + j * 16 + x // 8] |= lookup[src[r * 128 + x]] & (128 >> (x % 8))
    return grid, bytes(src[start:])


frames_path = Path("docs/internals/data/battle-second-frames.json")
frames = json.loads(frames_path.read_text())
root = Path("workplace/oracle/battle79-second-r36-a")
entry = root / "entry-mem.bin"
assert sha(entry) == "9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec"
mem = entry.read_bytes()
ds = 0x54B30
lookup = mem[ds + 0x3C8C:ds + 0x3D8C]
assert hashlib.sha256(lookup).hexdigest() == "7d8ca2d554b769865a290bd0428955e46d63c1a58613ca18295c9ee10cb3b2ea"
semantics = {
    "1727:0004": "配置2048-byte小地圖bitmap",
    "1727:000F": "保存bitmap指標段值於DS:A424",
    "1727:0012": "保存bitmap指標位移於DS:A422",
    "1727:0026": "以DS:3C1A地圖遠指標呼叫bitmap建立常式",
    "1727:0044": "繪製輸入虛擬y=36，實際起列76見暫存器收據",
    "1727:0048": "繪製輸入x=480",
    "1727:004C": "呼叫2376:0B44繪製128乘128 bitmap",
    "1727:05E1": "迭代128個輸出列",
    "1727:0600": "讀八個高位優先bitmask的其中一個",
    "1727:0607": "讀來源地圖遠指標",
    "1727:060A": "來源地圖指標每像素增加1",
    "1727:0612": "以地圖byte查DS:3C8C，不限制64-entry合法圖塊範圍",
    "1727:0616": "查表byte與單像素bitmask作AND",
    "1727:0618": "累積目前八像素的輸出byte",
    "1727:062B": "每八像素寫一byte到bitmap",
    "1727:063C": "列餘數加8",
    "1727:0644": "列餘數除10",
    "1727:064B": "餘數為零時回退128-byte地圖列",
    "1727:064D": "最後一列亦回退，已讀內容不變",
    "1727:0653": "來源指標回退128 bytes，採16-bit位移",
    "2376:0B5B": "將座標helper結果BX作實際VRAM起點DI",
    "2376:0B6D": "背景迭代128列",
    "2376:0B71": "背景每列16個VRAM bytes",
    "2376:0B85": "載入bitmap來源DS:SI",
    "2376:0B88": "前景迭代128列",
    "2376:0B8F": "讀一個bitmap byte，SI加1",
    "2376:0B90": "用bitmap byte作前景AND寫入；y203最後写入者",
    "2376:0B97": "每列另加64，使總stride為80 VRAM bytes",
}
entries, exports = [], {}
for suffix in ("build", "blit-origin"):
    p = Path("workplace/ida/re79-residue-r38-" + suffix + ".json")
    d = json.loads(p.read_text())
    assert d["tool"] == "IDA Pro 9.4" and d["input_sha256"] == sha(p.with_suffix(".com"))
    exports[p.name] = {"sha256": sha(p), "input_sha256": d["input_sha256"], "index_sha256": sha(p.with_suffix(".index.json"))}
    for f in d["functions"]:
        for r in f["instructions"]:
            loc = r["original_location"]
            if loc not in semantics:
                continue
            a = int(r["runtime_linear_hex"], 16)
            b = bytes.fromhex(r["original_bytes_hex"])
            assert mem[a:a + len(b)] == b
            entries.append({**r, "semantic": semantics[loc], "grade": "proven", "warning": "",
                            "source": ["docs/re/378-minimap-bitmap-bottom.md", sha(p)]})
assert len(entries) == len(semantics)
rows = []
samples = [("S1-" + f["name"], root / (f["name"] + "-mem.bin"), "MAP08.CMP", f["hashes"]["-mem.bin"]) for f in frames["frames"]]
samples += [("S3-fight", Path("workplace/s3fight.bin"), "MAP26.CMP", "0f8a91acb618d0ec508e857853be1cfe249ec8fe522c8aacac658ec03cbaf060"),
            ("S3-ai6", Path("workplace/s3ai6.bin"), "MAP26.CMP", "427c15414ce8d265551a05ecd2d49e8157ecede82a80154b6319c7ced4a8d3e0")]
for name, p, mapname, expected in samples:
    assert sha(p) == expected
    m = p.read_bytes()
    assert m[ds + 0x3C8C:ds + 0x3D8C] == lookup
    off, seg = struct.unpack_from("<HH", m, ds + 0x3C1A)
    mapbase = seg * 16 + off
    off, seg = struct.unpack_from("<HH", m, ds + 0xA422)
    outbase = seg * 16 + off
    assert outbase == mapbase + 13058 and m[mapbase + 13056:outbase] == bytes.fromhex("02 08")
    mp = Path("/orig") / mapname
    grid, out = bitmap(decode(mp.read_bytes()), lookup)
    assert grid == m[mapbase:mapbase + 13056]
    assert out == m[outbase:outbase + 2048]
    rows.append({"name": name, "ram_sha256": sha(p), "map": mapname, "map_sha256": sha(mp),
                 "map_linear_hex": hex(mapbase), "bitmap_linear_hex": hex(outbase),
                 "grid_sha256": hashlib.sha256(grid).hexdigest(), "bitmap_sha256": hashlib.sha256(out).hexdigest(),
                 "last_row_hex": out[-16:].hex(" "), "different_grid_bytes": 0, "different_bitmap_bytes": 0})
exploration = []
for suffix in ("strip-write", "strip-source"):
    p = Path("workplace/oracle/battle79-round38-" + suffix)
    exploration.append({"path": str(p), "classification": "source exploration only",
                        "flags": (p / "flags.txt").read_text().splitlines(),
                        "hashes": {f.name: sha(f) for f in sorted(p.iterdir()) if f.is_file()}})
report = {"schema": "logh3.battle-minimap-bitmap.v1", "tool": "IDA Pro 9.4", "program_sha256": frames["program_sha256"],
          "dosgolem_head": frames["dosgolem_head"], "frame_source_sha256": sha(frames_path),
          "input_sha256": sha(entry), "address_space": "原始CS:IP、IDA .COM載入位址、dosgolem線性位址分欄；DS固定定位只適用此RAM",
          "bitmap_origin": [480, 76], "bitmap_dimensions": [128, 128], "frame_top_y": 75,
          "lookup_offset_hex": "0x3c8c", "lookup_hex": lookup.hex(" "), "lookup_sha256": hashlib.sha256(lookup).hexdigest(),
          "adjacent_header_hex": "02 08", "entries": entries, "exports": exports, "samples": rows, "exploration": exploration,
          "scope": "S1固定43邊界及兩份既有S3；原版鄰接配置以明示有界資料重現，不宣稱任意heap配置相同",
          "unknown": ["不同配置次序的heap鄰接是否相同", "完整八步其餘階段"]}
out = Path("docs/internals/data/battle-minimap-bitmap.json")
out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print(len(entries), "筆分級指令；45份13056-byte地圖與2048-byte bitmap全部相同", sha(out))
