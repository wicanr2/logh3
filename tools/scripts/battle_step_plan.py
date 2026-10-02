"""研究372：從已驗入口建立固定移動單步注入計畫（不選種子）。"""
import hashlib
import argparse
import json
import struct
from pathlib import Path

W = Path("/work/workplace")
raw = (W / "oracle/battle79-move-entry-a/mem-in.bin").read_bytes()
assert hashlib.sha256(raw).hexdigest() == "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d"
ds = 0x54B30
word = lambda off: struct.unpack_from("<H", raw, ds + off)[0]
pool_seg, pool_off = word(0xCFB6), word(0xCFB4)
pool = pool_seg * 16 + pool_off
map_at = word(0x3C1C) * 16 + word(0x3C1A)
fleet_at = ds + 0xC814
positions = {"fleet": (fleet_at, 122), "flagship": (pool + 102 * 78, 78),
             "unit": (pool + 104 * 78, 78), "scout": (pool + 103 * 78, 78),
             "enemy": (pool + 142 * 78, 78)}
records = {k: bytearray(raw[at:at + size]) for k, (at, size) in positions.items()}


def p16(b, off, v):
    struct.pack_into("<H", b, off, v & 65535)


def p32(b, off, v):
    struct.pack_into("<I", b, off, v & 0xFFFFFFFF)


f = records["fleet"]
f[15] = 2
p16(f, 0x10, 102); p16(f, 0x14, 104); p16(f, 0x38, 103)
f[0x3A], f[0x3B], f[0x48], f[0x49], f[0x4A], f[0x4C], f[0x4D] = 255, 0, 0, 0, 0, 0, 4
p32(f, 0x52, 32 << 16); p32(f, 0x56, 32 << 16)
for off, value in [(0x5C, 1000), (0x5E, 1000), (0x64, 1100), (0x66, 1000)]:
    p16(f, off, value)
for label in ["flagship", "unit", "scout"]:
    b = records[label]
    b[5], b[0x1E] = 20, 127
    b[0x16] &= 253
    p32(b, 0x24, 32 << 16)
    for off in [0x28, 0x2A, 0x34, 0x36]:
        p16(b, off, 1000)

common = [{"linear": ds + 0x9894, "hex": "0100"},
          {"linear": ds + 0x9898, "hex": "10"},
          {"linear": ds + 0x9896, "hex": "0000"},
          {"linear": ds + 0x94DC, "hex": "78563412"},
          {"linear": map_at, "hex": bytes(128 * 102).hex()}]
common += [{"linear": positions[k][0], "hex": v.hex()} for k, v in records.items()]
far_f = [0xC814, 0x54B3]
far_u = lambda index: [pool_off + 78 * index, pool_seg]
cases = []


def add(name, seg, off, args, edits=None, extra=None):
    injections = []
    for key, edit in (edits or {}).items():
        b = records[key].copy()
        edit(b)
        injections.append({"linear": positions[key][0], "hex": b.hex()})
    injections += extra or []
    cases.append({"name": name, "injections": injections,
                  "calls": [{"segment": seg, "offset": off, "args": args}]})


def change(b, fields):
    for off, value, width in fields:
        if width == 1: b[off] = value & 255
        elif width == 2: p16(b, off, value)
        else: p32(b, off, value)


# 旗艦同步：相同現在座標不應抄目的地；不同才同步兩者。
for label, index in [("flagship", 102), ("scout", 103)]:
    for same in [True, False]:
        add(f"sync-{label}-{same}", 0x1C61, 0x0229, far_f + far_u(index),
            {label: lambda b, same=same: change(b, [(0x28, 1100, 2), (0x34, 1000 if same else 980, 2)])})

# 直接分隊步進：步長與扣額分離、到點、斜向及地圖夾界。
for step in [8, 12, 16, 24]:
    for dx, dy in [(100, 0), (100, 100), (3, 4), (-100, 0)]:
        add(f"unit-step-{step}-{dx}-{dy}", 0x1C61, 0x0444, far_u(104) + [step],
            {"unit": lambda b, dx=dx, dy=dy: change(b, [(0x28, 1000 + dx, 2), (0x2A, 1000 + dy, 2), (0x24, 8 << 16, 4)])})
for x, y, tx, ty in [(0, 0, 1, 1), (2031, 2023, 2047, 2047), (2047, 2047, 0, 0)]:
    add(f"unit-edge-{x}-{y}", 0x1C61, 0x0444, far_u(104) + [8],
        {"unit": lambda b, x=x, y=y, tx=tx, ty=ty: change(b, [(0x34,x,2),(0x36,y,2),(0x28,tx,2),(0x2A,ty,2)])})

# 跟隊入口：正常／回轉／追隊及明確歸隊，步長8／12／16／24。
for turned in [0, 1]:
    for lag in [0, 2]:
        for action in [127, 3]:
            add(f"follow-{turned}-{lag}-{action}", 0x1C61, 0x029C, far_f + far_u(104),
                {"fleet": lambda b, turned=turned: change(b, [(0x49,turned,1)]),
                 "unit": lambda b, lag=lag, action=action: change(b, [(0x34,900,2),(0x16, (b[0x16] & 253) | lag,1),(0x1E,action,1),(0x24,8<<16,4)])})

# 自主移動走完後仍需下一次呼叫才改歸隊。
for kind, label, index in [(0,"unit",104),(7,"scout",103)]:
    for distance in [5, 8, 100]:
        add(f"own-step-{kind}-{distance}", 0x1C61, 0x03AA, far_u(index) + far_f,
            {label: lambda b, distance=distance: change(b, [(0x1E,0,1),(0x28,1000+distance,2),(0x24,8<<16,4)])})

# 艦隊分派必須先滿足8.0；到點轉向也受外層門檻限制。
for mode in [0, 16]:
    for facing, target, budget, dx, dy in [
        (0,0,32<<16,100,0), (0,0,32<<16,100,100),
        (0,0,8<<16,3,4), (0,0,(8<<16)-1,100,0),
        (0,4,8<<16,0,0), (0,12,8<<16,0,0),
        (0,4,0,0,0), (64,0,16<<16,100,0),
    ]:
        add(f"fleet-{mode}-{facing}-{target}-{budget}-{dx}-{dy}", 0x1C03, 0x006A, far_f,
            {"fleet": lambda b, mode=mode, facing=facing, target=target, budget=budget, dx=dx, dy=dy:
             change(b, [(0x48,mode,1),(0x4A,facing,1),(0x4C,target,1),(0x56,budget,4),(0x64,1000+dx,2),(0x66,1000+dy,2)])})

# 地形與敵方碰撞：普通分隊避障，艦隊遇敵不進避障；索敵隊有特殊狀態。
wall = [{"linear": map_at + row*128 + 63, "hex": "08"} for row in range(49,53)]
all_wall = [{"linear": map_at, "hex": bytes([8])*(128*102)}]
all_wall[0]["hex"] = all_wall[0]["hex"].hex()
enemy = records["enemy"].copy();enemy[0],enemy[4],enemy[0x1E]=0,0,127
p16(enemy,0x34,1008);p16(enemy,0x36,1000)
enemy_in = [{"linear": positions["enemy"][0], "hex": enemy.hex()},
            {"linear":ds+0x9896,"hex":"0100"}, {"linear":ds+0x98BA,"hex":"8e00"}]
for obstacle, extra in [("wall",wall),("all-wall",all_wall),("enemy",enemy_in)]:
    add("fleet-block-"+obstacle,0x1C03,0x006A,far_f,extra=extra)
    for label,index,action in [("unit",104,0),("scout",103,0),("scout",103,3)]:
        add(f"unit-block-{label}-{action}-{obstacle}",0x1C61,0x0444,far_u(index)+[8],
            {"fleet":lambda b:change(b,[(0x3A,0,1)]),
             label:lambda b,action=action:change(b,[(0x1E,action,1),(0x28,1100,2)])},extra)

# 收回／未放出索敵隊、Action與第0分隊分派優先順序。
add("scout-rejoin-arrived",0x1C61,0x0153,far_f+[65535],
    {"fleet":lambda b:change(b,[(0x3A,0,1),(0x3B,2,1)]),
     "scout":lambda b:change(b,[(0x1E,3,1),(0x19,b[0x19]&127,1)])})
add("scout-inactive-sync",0x1C61,0x0153,far_f+[65535],
    {"scout":lambda b:change(b,[(0x34,980,2),(0x1E,128,1)])})
add("flagship-action-zero-sync",0x1C61,0x0153,far_f+[0],
    {"flagship":lambda b:change(b,[(0x34,980,2),(0x1E,0,1)])})
for action in [1,2,128,131,255]:
    add(f"unit-action-inactive-{action}",0x1C61,0x0153,far_f+[1],
        {"unit":lambda b,action=action:change(b,[(0x1E,action,1),(0x28,1100,2)])})

parser = argparse.ArgumentParser()
parser.add_argument("--extended", action="store_true")
parser.add_argument("--output", default="battle79-step-plan.json")
options = parser.parse_args()
if options.extended:
    # 起點992尚未接觸牆：驗證搜尋的正負16格順序，而非已在牆內的退路。
    add("fleet-wall-from-open", 0x1C03, 0x006A, far_f,
        {"fleet": lambda b: change(b, [(0x5C,992,2)])}, wall)
    add("unit-wall-from-open", 0x1C61, 0x0444, far_u(104)+[8],
        {"unit": lambda b: change(b, [(0x34,992,2),(0x28,1100,2)])}, wall)
    for mode in [0, 16]:
        # 首次正北落點在牆；東向通行。艦隊避障傳入的是舊座標。
        north_wall = [{"linear":map_at+50*128+62,"hex":"08"}]
        add(f"fleet-north-wall-{mode}", 0x1C03, 0x006A, far_f,
            {"fleet": lambda b, mode=mode: change(b, [(0x48,mode,1),(0x4A,64,1),(0x64,1000,2),(0x66,1100,2)])}, north_wall)
        # 完整Action只允許00/01/10/11。禁用動作不得清回轉旗標。
        for action in [0,1,2,16,17,18,128]:
            add(f"fleet-dispatch-action-{mode}-{action}",0x1C03,0x006A,far_f,
                {"fleet":lambda b,action=action:change(b,[(0x48,action,1),(0x49,0xA1,1)]),
                 "unit":lambda b:change(b,[(0x28,1100,2)])})
        add(f"fleet-arrived-clear-turned-{mode}",0x1C03,0x006A,far_f,
            {"fleet":lambda b,mode=mode:change(b,[(0x48,mode,1),(0x49,0xA1,1),(0x64,1000,2),(0x4C,4,1)])})
    for dx, dy, faction, kind, action in [
        (15,0,0,0,127),(16,0,0,0,127),(0,15,0,0,127),(0,16,0,0,127),
        (15,15,0,0,127),(16,15,0,0,127),(0,0,1,0,127),
        (0,0,0,7,127),(0,0,0,0,128),
    ]:
        e = enemy.copy()
        change(e,[(0,faction,1),(4,kind,1),(0x1E,action,1),(0x34,1000+dx,2),(0x36,1000+dy,2)])
        extra=[{"linear":positions["enemy"][0],"hex":e.hex()}]+enemy_in[1:]
        add(f"collision-{dx}-{dy}-{faction}-{kind}-{action}",0x2583,0x040C,[1000,1000,1],extra=extra)

plan = {"common":common,"observations":[{"label":k,"linear":at,"length":size} for k,(at,size) in positions.items()],"cases":cases}
output = W / options.output
assert not output.exists(), "不得覆寫先前計畫"
output.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+"\n")
print("案例",len(cases),"計畫SHA-256",hashlib.sha256(output.read_bytes()).hexdigest())
