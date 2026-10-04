"""由第91輪固定觀察器記錄占領callback前後的資料，不改輸入。"""
from pathlib import Path
import hashlib

b = Path("workplace/finish79-r91-native.go").read_bytes()
assert hashlib.sha256(b).hexdigest() == "7f41a1a106485b14cb23eec105cd9df0ecd2b0b1abdefe37e45b522703243519"
s = b.decode()
anchor = '   if a==far(0x3826,0x0405) {'
assert s.count(anchor) == 1
s = s.replace(anchor, '''   if a==far(0x185C,0x0B18) {
    call["tables"]=tables();call["seed_hex"]=hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4))
    ret:=far(o.Word(far(r.SS,r.SP+2)),o.Word(far(r.SS,r.SP)))
    call["return_location"]=ret.String()
    o.OnCall(ret,func(o *control){call["return_tables"]=tables();call["return_seed_hex"]=hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4))})
   }
''' + anchor, 1)
s = s.replace("logh3.battle-finish-raw-native.v1", "logh3.battle-finish-cursor-native.v1")
s = s.replace("bash tools/battle_finish_raw_oracle.sh <a|b>", "bash tools/battle_finish_cursor_oracle.sh <a|b>")
dest = Path("workplace/finish79-r91-cursor-native.go")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
