"""由固定第89輪控制器導出逐指令寫入觀察；不更動執行器或原版。"""
from pathlib import Path
import hashlib

p = Path("workplace/finish79-r89-native.go")
b = p.read_bytes()
assert hashlib.sha256(b).hexdigest() == "cbe370d25969cd96e78effc7bc50940c133f9fbab180b53eab4cf20d58ad5580"
s = b.decode()
start = s.index(" type scenario struct")
end = s.index(" rows:=[]map[string]any{}", start)
s = s[:start] + " type scenario struct{mask,human,round,recovery int}\n scenarios:=[]scenario{{1,0,5,0},{126,0,5,0}}\n" + s[end:]
s = s.replace("   before=tables()", "   before=tables();o.watching=true", 1)
s = s.replace('"before":before,"after":tables(),', '"before":before,"after":tables(),"writes":o.writes,', 1)
s = s.replace("logh3.battle-finish-native.v1", "logh3.battle-finish-raw-native.v1")
s = s.replace("bash tools/battle_finish_oracle.sh <a|b>", "bash tools/battle_finish_raw_oracle.sh <a|b>")
s = s.replace("    frames []map[string]any\n", "    frames []map[string]any\n    watching bool\n    writes []map[string]any\n", 1)
s = s.replace("    o.frames = nil\n", "    o.frames = nil\n    o.watching=false; o.writes=nil\n", 1)
old = "        if err := o.m.Step(); err != nil { return 0, err }"
new = r'''
        ids:=[]int{12,16,17,28,29,30,31}
        prev:=make([]uint16,len(ids));instruction:=[]byte(nil)
        if o.watching {instruction=o.Bytes(a,12);for i,f:=range ids{prev[i]=o.Word(far(0x54B3,0xC074+uint16(f*122)+0x5A))}}
        if err := o.m.Step(); err != nil { return 0, err }
        if o.watching {for i,f:=range ids {p:=far(0x54B3,0xC074+uint16(f*122)+0x5A);v:=o.Word(p);if v!=prev[i]{o.writes=append(o.writes,map[string]any{"instruction":a.String(),"instruction_bytes_hex":hex.EncodeToString(instruction),"step":o.Steps()-1,"fleet":f,"linear":fmt.Sprintf("lin:%X",p.Linear()),"before_word":prev[i],"after_word":v})}}}
'''
assert s.count(old) == 1
s = s.replace(old, new, 1)
dest = Path("workplace/finish79-r91-native.go")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
