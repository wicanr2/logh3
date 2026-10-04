"""由第89輪固定控制器導出敗方守軍探針，保存完整注入與查詢結果。"""
from pathlib import Path
import hashlib

source = Path("workplace/finish79-r89-native.go").read_bytes()
assert hashlib.sha256(source).hexdigest() == "cbe370d25969cd96e78effc7bc50940c133f9fbab180b53eab4cf20d58ad5580"
s = source.decode()
start = s.index(" type scenario struct")
end = s.index(" rows:=[]map[string]any{}", start)
s = s[:start] + ''' type scenario struct{mask,human,round,recovery,guard int}
 scenarios:=[]scenario{{1,0,5,0,1},{1,0,5,0,2},{1,0,5,0,3},{1,0,5,0,4},{1,0,5,0,5},{1,0,0,0,1},{126,0,5,0,1}}
''' + s[end:]
injection = r'''
   if s.guard!=0 {
    winner:=byte(0);loser:=12;if s.mask==1{winner=1}else{loser=16}
    pp:=far(o.Word(far(0x54B3,0xCFFA)),o.Word(far(0x54B3,0xCFF8)))
    n:=int(o.Bytes(far(0x54B3,0xCFFC),1)[0]);found:=false
    for i:=0;i<n;i++ {p:=far(pp.Seg,pp.Off+uint16(i*82));if o.Bytes(far(p.Seg,p.Off+0x11),1)[0]!=winner {pp=p;found=true;break}}
    if !found{panic("無敵方行星")}
    x,y:=o.Word(far(pp.Seg,pp.Off+0x46)),o.Word(far(pp.Seg,pp.Off+0x48))
    f:=o.Bytes(far(0x54B3,0xC074+uint16(loser*122)),122)
    moved:=0;want:=1;if s.guard==4 || s.guard==5 {want=2}
    for i:=0;i<int(f[15]) && moved<want;i++ {
     pool:=binary.LittleEndian.Uint16(f[16+i*4:]);p:=far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))+pool*78)
     if o.Bytes(far(p.Seg,p.Off+4),1)[0]>1 {continue}
     value:=make([]byte,2);xx:=x;if s.guard==3 {xx+=512}
     binary.LittleEndian.PutUint16(value,xx);set(far(p.Seg,p.Off+0x34),value)
     binary.LittleEndian.PutUint16(value,y);set(far(p.Seg,p.Off+0x36),value)
     out:=o.Bytes(far(p.Seg,p.Off+0x1E),1)[0]&0x7F;if s.guard==2 {out|=0x80};set(far(p.Seg,p.Off+0x1E),[]byte{out})
     if s.guard==5 {alive:=uint16(1);if moved==0{alive=65535};binary.LittleEndian.PutUint16(value,alive);set(far(p.Seg,p.Off+8),value)}
     moved++
    }
    if moved!=want {panic("守軍種類不足")}
   }
'''
s = s.replace("   before=tables()", injection + "   before=tables()", 1)
s = s.replace("far(0x17A0,0x00AC)}", "far(0x17A0,0x00AC),far(0x203A,0x04D1)}", 1)
s = s.replace("if a==far(0x17A0,0x00AC) {", "if a==far(0x17A0,0x00AC) || a==far(0x203A,0x04D1) {", 1)
s = s.replace('fmt.Sprintf("mask%02X-human%d-round%d-recovery%d",s.mask,s.human,s.round,s.recovery)', 'fmt.Sprintf("mask%02X-human%d-round%d-guard%d",s.mask,s.human,s.round,s.guard)')
s = s.replace('"recovery":s.recovery,', '"recovery":s.recovery,"guard":s.guard,', 1)
s = s.replace("logh3.battle-finish-native.v1", "logh3.battle-finish-guard-native.v1")
s = s.replace("bash tools/battle_finish_oracle.sh <a|b>", "bash tools/battle_finish_guard_oracle.sh <a|b>")
dest = Path("workplace/finish79-r90-guard-native.go")
with dest.open("x") as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
