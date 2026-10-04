"""第93輪原版艦隊命令局部驗證；選點替身明示，不宣稱正常玩家對拍。"""
from pathlib import Path
import hashlib

source = Path("workplace/finish79-r89-native.go").read_bytes()
assert hashlib.sha256(source).hexdigest() == "cbe370d25969cd96e78effc7bc50940c133f9fbab180b53eab4cf20d58ad5580"
prefix = source.decode().split("func digest")[0]
control = "type addr" + source.decode().split("type addr", 1)[1]
control = control.replace('if entry != far(0x185C, 0x001D)', 'if entry != far(0x1CC5, 0x0813) && entry != far(0x1CC5, 0x08BD)')
control = control.replace('a == far(0x185C, 0x001B)', 'a == far(0xFFF0, 0)')
main = r'''
func digest(b []byte) string { h:=sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func must(e error) { if e!=nil { panic(e) } }
func main() {
 output:=flag.String("output","","新原版JSON");flag.Parse()
 exe,e:=os.ReadFile("/orig/GIN3.COM");must(e)
 if digest(exe)!="770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {panic("原版不符")}
 o:=&control{callbacks:make(map[addr]func(*control))};defer o.Close()
 rows:=[]map[string]any{}
 for _,entry:=range []addr{far(0x1CC5,0x0813),far(0x1CC5,0x08BD)} {for _,action:=range []byte{0,3,127,255} {
  o.reload();o.callbacks=make(map[addr]func(*control))
  fleet:=far(0x54B3,0xC074+16*122);fr:=o.Bytes(fleet,122)
  if fr[1]!=1 || fr[15]==0 {panic("父艦隊不符")}
  pool:=far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4)))
  flagPool:=binary.LittleEndian.Uint16(fr[16:]);flag:=far(pool.Seg,pool.Off+flagPool*78)
  if o.Bytes(flag,4)[3]!=0 {panic("非旗艦")}
  injections:=[]map[string]any{}
  set:=func(a addr,b []byte,reason string){o.WriteBytes(a,b);if !bytes.Equal(o.Bytes(a,len(b)),b){panic("注入回讀")};injections=append(injections,map[string]any{"linear":fmt.Sprintf("lin:%X",a.Linear()),"hex":hex.EncodeToString(b),"reason":reason})}
  set(far(flag.Seg,flag.Off+0x1e),[]byte{action},"獨立測試原旗艦行動碼；不重擲seed")
  beforeShip:=o.Bytes(pool,352*78);beforeFleet:=o.Bytes(far(0x54B3,0xC074),32*122)
  seedBefore:=o.Bytes(far(0x54B3,0x94DC),4)
  selected:=false
  o.OnCall(far(0x2583,0x013D),func(o *control){
   if selected {panic("重複選點")};selected=true
   r:=o.Regs();arg:=func(i uint16)uint16{return o.Word(far(r.SS,r.SP+4+i*2))}
   // 只替代互動返回，目的地取原艦隊當前位置，朝向取原傳入值。
   b:=make([]byte,2);binary.LittleEndian.PutUint16(b,o.Word(far(fleet.Seg,fleet.Off+0x5c)));set(far(arg(1),arg(0)),b,"選點替身x")
   binary.LittleEndian.PutUint16(b,o.Word(far(fleet.Seg,fleet.Off+0x5e)));set(far(arg(3),arg(2)),b,"選點替身y")
   o.m.CPU.R[cpu.AX]=arg(4)
   o.m.CPU.IP=o.Word(far(r.SS,r.SP));o.m.CPU.Seg[cpu.CS]=o.Word(far(r.SS,r.SP+2));o.m.CPU.R[cpu.SP]+=4
  })
  r:=o.Regs();sp:=r.SP
  push:=func(v uint16){sp-=2;b:=make([]byte,2);binary.LittleEndian.PutUint16(b,v);set(far(r.SS,sp),b,"明示far診斷呼叫stack")}
  push(fleet.Seg);push(fleet.Off);push(0xFFF0);push(0)
  o.m.CPU.R[cpu.SP]=sp;o.m.CPU.Seg[cpu.CS]=entry.Seg;o.m.CPU.IP=entry.Off
  start:=o.Steps();ax,e:=o.CallBudget(4000000,entry);must(e);if !selected {panic("未經選點")}
  afterShip:=o.Bytes(pool,352*78)
  rows=append(rows,map[string]any{"entry":entry.String(),"flagship_pool":flagPool,"action":action,"before_ship_hex":hex.EncodeToString(beforeShip),"after_ship_hex":hex.EncodeToString(afterShip),"ship_equal":bytes.Equal(beforeShip,afterShip),"before_fleet_hex":hex.EncodeToString(beforeFleet),"after_fleet_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0xC074),32*122)),"seed_before_hex":hex.EncodeToString(seedBefore),"seed_after_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4)),"injections":injections,"cpu_entry":entry.String(),"cpu_return":"FFF0:0000","selection_substitute":"2583:013D返回原位置與原朝向；不改程式bytes","ax":ax,"steps":o.Steps()-start})
 }}
 raw,e:=json.MarshalIndent(map[string]any{"schema":"logh3.fleet-order-action-native.v1","scope":"局部原版函式診斷；選點互動替身；非正常玩家收據","program_sha256":digest(exe),"parent_state_sha256":"a62af47fb7d15a136e2e2ed51760ef0e9dd82538f68d5f2b0011b8","dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","go_version":runtime.Version(),"uid":os.Getuid(),"rows":rows},"","  ");must(e)
 f,e:=os.OpenFile(*output,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);_,e=f.Write(append(raw,'\n'));must(e);must(f.Close());fmt.Println("原版艦隊命令行動碼",len(rows),*output)
}
'''
with Path("workplace/order79-r93-native.go").open("x") as f:
    f.write(prefix + main + control)

ram = Path("workplace/oracle/cleanup79-r88-entry-a/mem-in.bin").read_bytes()
assert hashlib.sha256(ram).hexdigest() == "1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf"
for segment in (0x1cc5, 0x2583, 0x25db):
    with Path(f"workplace/ida/order79-r93-{segment:04X}.com").open("xb") as f:
        f.write(bytes(256) + ram[segment*16:segment*16+0x2000])
s = Path("tools/scripts/battle_promotion_ida.py").read_text()
with Path("workplace/ida/order79-r93-ida.py").open("x") as f:
    f.write(s.replace("promotion79-r84-", "order79-r93-").replace("battle-promotion-ida", "fleet-order-action-ida"))
print("原版局部探針與一次性IDA輸入已生成")
