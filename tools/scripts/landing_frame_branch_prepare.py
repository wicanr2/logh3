"""沿凍結第108輪觀察器生成第109輪三個明示分支；不改舊來源。"""
import hashlib
import json
import os
from pathlib import Path

assert Path.cwd() == Path('/src') and os.getuid() == 1000
old = Path('tools/scripts/landing_scene_native.go')
source = old.read_bytes()
index = json.loads(Path('docs/internals/data/landing-scene.json').read_text())
assert hashlib.sha256(source).hexdigest() == index['sources'][str(old)]['sha256']
t = source.decode()
t = t.replace('第108輪', '第109輪')
t = t.replace('baseline|failure|capture|off|click', 'counterfirst|empire|multiple')
t = t.replace('kind != "baseline" && kind != "failure" && kind != "capture" && kind != "off" && kind != "click"',
              'kind != "counterfirst" && kind != "empire" && kind != "multiple"')
t = t.replace('revision := "r5"', 'revision := "r1"')
t = t.replace('landing48-r108-', 'landing48-r109-branch-')
t = t.replace('bash tools/landing_scene_oracle.sh ', 'bash tools/landing_frame_branch_oracle.sh ')
needle = '\t\t\tif kind == "off" {\n'
assert t.count(needle) == 1
t = t.replace(needle, '''
			set(planet+0x14, []byte{0x70,0x17,0x70,0x17}, "三分支統一防衛力及上限6000；保留固定seed，不挑結果")
			if kind == "counterfirst" {
				set(fleet+0x3c, []byte{50}, "0503以艦隊完整度50小於行星防衛比例100，強制走先反擊分支")
			}
			if kind == "empire" {
				set(fleet+1, []byte{0}, "明示將本次攻方陣營改帝國；不是正常玩家局面")
				set(fleet+4, []byte{0}, "明示帝國提督0，供已存在的帝國旗艦COMANI圖段")
				set(planet+0x11, []byte{1}, "明示敵方行星改同盟")
				ship := far(ds+0xcfb4)
				for j:=0;j<int(m.Read8(fleet+0x0f));j++ {
					idx:=m.Read16(fleet+0x10+uint32(j*4))
					set(ship+uint32(idx)*78, []byte{0}, "完整攻方名冊SHIP陣營同步帝國")
				}
			}
			if kind == "multiple" {
				other := ds+0xc074+30*122
				if m.Read8(other+1)!=1 || m.Read8(other+0x0f)==0 || m.Read8(other+4)==255 {panic("第二攻方30身份不符")}
				set(other+0x48, []byte{1}, "第二攻方30命令登陸")
				set(other+0x70, []byte{m.Read8(other+0x70)|1}, "第二攻方30明示補登陸eligibility")
				off:=uint16(other-ds)
				set(list+4, []byte{byte(off),byte(off>>8),0xb3,0x54}, "父名冊末尾加入現存艦隊30遠指標")
				set(entry+10, []byte{2,0}, "0486名冊長度改2；完整堆疊注入記錄")
			}
''' + needle)
target = Path('tools/scripts/landing_frame_native.go')
payload = t.encode()
if target.exists():
    assert target.read_bytes() == payload, '拒絕改寫已生成的觀察器'
else:
    target.write_bytes(payload)
print(target, hashlib.sha256(payload).hexdigest())

# r1的雙名冊只消費第一支。0486另以FLEET+5A篩選行星，r2明示補第二支目標。
t2 = t.replace('revision := "r1"', 'revision := "r2"')
needle2 = '\t\t\t\tset(other+0x48, []byte{1}, "第二攻方30命令登陸")'
assert t2.count(needle2) == 1
t2 = t2.replace(needle2, needle2 + '\n\t\t\t\tset(other+0x5a, []byte{0}, "第二攻方30目標行星0；0486:04CF與名冊行星核對")')
t2 = t2.replace('"bash tools/landing_frame_branch_oracle.sh " + kind + " " + side,',
                '"bash tools/landing_frame_branch_oracle.sh " + kind + " " + side + " r2",')
target2 = Path('tools/scripts/landing_frame_native_r2.go')
if target2.exists():
    assert target2.read_bytes() == t2.encode(), '拒絕改寫r2觀察器'
else:
    target2.write_text(t2)
print(target2, hashlib.sha256(t2.encode()).hexdigest())
