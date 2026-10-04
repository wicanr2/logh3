"""#79 第85輪：擴充凍結的第84輪觀測器，不覆寫舊收據。"""
from pathlib import Path
import hashlib

parent = Path('workplace/promotion79-r84-native-v2.go').read_bytes()
assert hashlib.sha256(parent).hexdigest() == 'a61e1b4d5d9b5e4a55ba63bcd08713b0823ad1d168ea5c2796c233f78dcebb34'
s = parent.decode()
s = s.replace('var calls []map[string]any', '''var calls []map[string]any
 tables:=func()map[string]string { return map[string]string{
  "admiral":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C),82*28)),
  "fleet":hex.EncodeToString(o.Bytes(far(0x54B3,0xC074),32*122)),
  "ship":hex.EncodeToString(o.Bytes(far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))),352*78)),
  "work":hex.EncodeToString(o.Bytes(far(0x54B3,0x8F14),230)),
  "stellar":hex.EncodeToString(o.Bytes(far(0x54B3,0xA7BC),32*126))} }
 var recalcs []map[string]any''')
s = s.replace('far(0x28E0,0x05F9),', 'far(0x185C,0x066E),far(0x185C,0x06BB),far(0x090A,0x0E6D),far(0x297B,0x116C),far(0x28E0,0x05F9),', 1)
s = s.replace('calls=append(calls,row)', '''if a==far(0x185C,0x066E) || a==far(0x185C,0x06BB) {
    recalcs=append(recalcs,map[string]any{"location":a.String(),"step":o.Steps(),"tables":tables()})
   }
   calls=append(calls,row)''', 1)
s = s.replace('for _,tc:=range cases {', '''for _,f:=range []int{12,16} { for _,pattern:=range []int{5,6,7} { cases=append(cases,testCase{f,26,0,0,-1,pattern}) } }
 for _,tc:=range cases {''', 1)
s = s.replace('o.reload(); calls=nil', 'o.reload(); calls=nil; recalcs=nil', 1)
s = s.replace('if tc.pattern!=0 {', 'if tc.pattern>0 && tc.pattern<5 {', 1)
s = s.replace('if tc.adj>=0 {', '''if tc.pattern>=5 {
   f:=o.Bytes(fr,122)
   set(fr,[]byte{f[0]|1}); set(far(fr.Seg,fr.Off+0x3C),[]byte{42})
   for i:=0;i<int(f[0x0F]);i++ {
    pool:=binary.LittleEndian.Uint16(f[0x10+i*4:])
    sh:=far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))+pool*78)
    full:=o.Word(far(sh.Seg,sh.Off+0xA)); alive:=full/2
    if tc.pattern==6 { alive=full/10 }; if tc.pattern==7 { alive=0 }
    b:=make([]byte,2); binary.LittleEndian.PutUint16(b,alive); set(far(sh.Seg,sh.Off+8),b)
   }
  }
  if tc.adj>=0 {''', 1)
start=s.index('  tables:=func()map[string]string')
end=s.index('\n  before:=tables()',start)
s=s[:start]+s[end:]
s=s.replace('"before":before,"after":tables()', '"before":before,"after":tables(),"recalculations":recalcs', 1)
s=s.replace('logh3.battle-promotion-native.v1','logh3.battle-promotion-recalc-native.v1')
s=s.replace('bash tools/battle_promotion_oracle.sh <a|b>','bash tools/battle_promotion_recalc_oracle.sh <a|b>')
s=s.replace('"rows":rows}', '"observer_parent_sha256":digest(parentObserver),"rows":rows}')
s=s.replace('exe,e:=os.ReadFile', 'parentObserver,e:=os.ReadFile("/task/workplace/promotion79-r84-native-v2.go"); must(e)\n exe,e:=os.ReadFile',1)
dest=Path('workplace/promotion79-r85-recalc.go')
with dest.open('x') as f:
    f.write(s)
print(dest,hashlib.sha256(dest.read_bytes()).hexdigest())
