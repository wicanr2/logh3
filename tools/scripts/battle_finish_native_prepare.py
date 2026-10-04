"""由固定父收據工具產生第89輪原版觀測器；只供診斷，不改正式程式。"""
from pathlib import Path
import hashlib

source = Path("workplace/cleanup79-r88-native.go").read_bytes()
assert hashlib.sha256(source).hexdigest() == "673cb53134aa548b85a19ac1942bf54add25fcdc782d802f73663e0e49d5ad39"
prefix = source.decode().split("func digest")[0]
control = "type addr" + source.decode().split("type addr", 1)[1]
control = control.replace("far(0x185C, 0x02DA)", "far(0x185C, 0x001B)")
control = control.replace("已核對的8-4入口", "已核對的8-1入口").replace("未經8-4就到8-4", "未經8-1就到8-5返回").replace("未到8-4", "未到8-5返回")
main = r'''
func digest(b []byte) string { h:=sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func must(e error) { if e!=nil { panic(e) } }
func main() {
 output:=flag.String("output","","新原版JSON"); flag.Parse(); if *output=="" {panic("缺輸出")}
 exe,e:=os.ReadFile("/orig/GIN3.COM");must(e)
 if digest(exe)!="770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {panic("原版不符")}
 o:=&control{callbacks:make(map[addr]func(*control))};defer o.Close()
 tables:=func()map[string]string{return map[string]string{
  "admiral":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C),82*28)),
  "fleet":hex.EncodeToString(o.Bytes(far(0x54B3,0xC074),32*122)),
  "ship":hex.EncodeToString(o.Bytes(far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))),352*78)),
  "work":hex.EncodeToString(o.Bytes(far(0x54B3,0x8F14),230)),
  "planet":hex.EncodeToString(append(o.Bytes(far(0x54B3,0xCFB8),64),o.Bytes(far(o.Word(far(0x54B3,0x5338)),o.Word(far(0x54B3,0x5336))),81*82)...)),
  "stellar":hex.EncodeToString(o.Bytes(far(0x54B3,0xA79C),32*126)),
  "lists":hex.EncodeToString(o.Bytes(far(0x54B3,0x9894),164))}}
 type scenario struct{mask,human,round,recovery int}
 scenarios:=[]scenario{}
 for _,round:=range []int{5,8,9} {for _,mask:=range []int{0,1,126,127} {for human:=0;human<4;human++ {scenarios=append(scenarios,scenario{mask,human,round,0})}}}
 for _,round:=range []int{0,1,6} {for _,mask:=range []int{1,126} {scenarios=append(scenarios,scenario{mask,0,round,0})}}
 for recovery:=1;recovery<=3;recovery++ {for _,mask:=range []int{1,126} {scenarios=append(scenarios,scenario{mask,0,8,recovery})}}
 rows:=[]map[string]any{}
 for _,s:=range scenarios {
  o.reload();active:=false;calls:=[]map[string]any{};inject:=[]map[string]any{};var before map[string]string
  set:=func(a addr,b []byte){o.WriteBytes(a,b);if !bytes.Equal(o.Bytes(a,len(b)),b){panic("注入回讀不符")};inject=append(inject,map[string]any{"linear":fmt.Sprintf("lin:%X",a.Linear()),"hex":hex.EncodeToString(b)})}
  o.OnCall(far(0x185C,0x09BF),func(o *control){
   if active {panic("重複8-5")};active=true
   set(far(0x54B3,0x94DC),[]byte{0x78,0x56,0x34,0x12})
   set(far(0x54B3,0x8F21),[]byte{byte(1-(s.human&1)),byte(1-((s.human>>1)&1))})
   round:=make([]byte,2);binary.LittleEndian.PutUint16(round,uint16(s.round));set(far(0x54B3,0x8F1A),round)
   fleets:=[]int{12,16,17,28,29,30,31}
   if o.Word(far(0x54B3,0x9894))!=7 || !bytes.Equal(o.Bytes(far(0x54B3,0x9898),7),[]byte{12,16,17,28,29,30,31}) {panic("父艦隊清單不符")}
   for i,f:=range fleets {a:=far(0x54B3,0xC074+uint16(f*122));faction:=o.Bytes(a,2)[1];expected:=byte(1);if i==0{expected=0};if faction!=expected{panic("父艦隊陣營不符")};v:=o.Bytes(a,1)[0]&0xFE;if s.mask&(1<<i)!=0 {v|=1};set(a,[]byte{v})}
   if s.recovery!=0 {
    p:=far(o.Word(far(0x54B3,0xCFFA)),o.Word(far(0x54B3,0xCFF8)));max:=o.Word(far(p.Seg,p.Off+0x16));d,rate:=uint16(0),uint16(100)
    if s.recovery==2 {d=65535};if s.recovery==3 {d=max-1;rate=0}
    value:=make([]byte,2);binary.LittleEndian.PutUint16(value,d);set(far(p.Seg,p.Off+0x14),value)
    binary.LittleEndian.PutUint16(value,rate);set(far(p.Seg,p.Off+0x18),value)
   }
   before=tables()
  })
  locations:=[]addr{far(0x185C,0x0A80),far(0x185C,0x0AF9),far(0x185C,0x0B18),far(0x20B4,0x1512),far(0x20B4,0x15E1),far(0x20B4,0x1662),far(0x20B4,0x0D96),far(0x20B4,0x0E7E),far(0x3826,0x0405),far(0x17A0,0x00AC)}
  for _,a:=range locations {a:=a;o.OnCall(a,func(o *control){if !active{return};r:=o.Regs();args:=[]uint16{};for i:=0;i<8;i++ {args=append(args,o.Word(far(r.SS,r.SP+4+uint16(i*2))))}
   call:=map[string]any{"location":a.String(),"step":o.Steps(),"args":args,"entry_bytes_hex":hex.EncodeToString(o.Bytes(a,24))};calls=append(calls,call)
   if a==far(0x20B4,0x1512) || a==far(0x20B4,0x1662) {call["tables"]=tables()}
   if a==far(0x3826,0x0405) {text:=[]byte{};p:=far(args[1],args[0]);for i:=0;i<2048;i++ {v:=o.Bytes(far(p.Seg,p.Off+uint16(i)),1)[0];if v==0{break};text=append(text,v)};call["text_hex"]=hex.EncodeToString(text)}
   if a==far(0x17A0,0x00AC) {ret:=far(o.Word(far(r.SS,r.SP+2)),o.Word(far(r.SS,r.SP)));call["return_location"]=ret.String();o.OnCall(ret,func(o *control){call["return_ax"]=o.m.CPU.R[cpu.AX]})}
  })}
  start:=o.Steps();ax,e:=o.CallBudget(40000000,far(0x185C,0x001D));must(e);if !active{panic("未到8-5")}
  rows=append(rows,map[string]any{"name":fmt.Sprintf("mask%02X-human%d-round%d-recovery%d",s.mask,s.human,s.round,s.recovery),"mask":s.mask,"human":s.human,"round":s.round,"recovery":s.recovery,"before":before,"after":tables(),"ax":ax,"calls":calls,"injections":inject,"inputs":o.inputs,"frames":o.frames,"steps":o.Steps()-start,"seed_after_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4))})
 }
 result:=map[string]any{"schema":"logh3.battle-finish-native.v1","program_sha256":digest(exe),"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","dosgolem_dirty":false,"go_version":runtime.Version(),"uid":os.Getuid(),"tick":20000,"seed":"0x12345678","command":"bash tools/battle_finish_oracle.sh <a|b>","scope":"固定8-1完整保存自然經8-1..8-4；8-5入口明示士氣bit0、人類座位、回合及seed；停185C:001B。診斷，不是正常玩家勝負收據。","rows":rows}
 raw,e:=json.MarshalIndent(result,"","  ");must(e);f,e:=os.OpenFile(*output,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);_,e=f.Write(append(raw,'\n'));must(e);must(f.Close());fmt.Println("結束判定原版案例",len(rows),*output)
}
'''
dest = Path("workplace/finish79-r89-native.go")
with dest.open("x") as f:
    f.write(prefix + main + control)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
