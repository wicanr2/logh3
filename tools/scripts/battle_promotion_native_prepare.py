"""#79 子步8-4：沿固定完整machine／DOS保存續行的診斷觀測器。"""
from pathlib import Path

source = Path('tools/scripts/scout_fuel_native_prepare.py').read_text()
tail = source.split("src += r'''", 1)[1].split("'''", 1)[0]
tail = tail.replace('far(0x185C, 0x0491)', 'far(0x185C, 0x060D)')
tail = tail.replace('if a == far(0x185C, 0x060D) {', 'if a == far(0x185C, 0x09BF) {')
tail = tail.replace('8-3', '8-4').replace('8-4之前', '8-5之前')
head = r'''
package main
import (
 "bytes"
 "crypto/sha256"
 "encoding/binary"
 "encoding/hex"
 "encoding/json"
 "flag"
 "fmt"
 "image"
 "image/color"
 "image/png"
 "os"
 "runtime"
 "github.com/wicanr2/dosgolem/internal/cpu"
 "github.com/wicanr2/dosgolem/internal/dos"
 "github.com/wicanr2/dosgolem/internal/machine"
 "github.com/wicanr2/dosgolem/internal/state"
)
func digest(b []byte) string { h:=sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func must(e error) { if e!=nil { panic(e) } }
func main() {
 output:=flag.String("output","","新的輸出JSON"); flag.Parse()
 if *output=="" { panic("缺輸出") }
 exe,e:=os.ReadFile("/orig/GIN3.COM"); must(e)
 if digest(exe)!="770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" { panic("原版不符") }
 o:=&control{callbacks:make(map[addr]func(*control))}; defer o.Close()
 rows:=[]map[string]any{}
 var calls []map[string]any
 for _,a:=range []addr{far(0x28E0,0x05F9),far(0x28E0,0x06C3),far(0x28E0,0x095C),far(0x297B,0x0546),far(0x185C,0x06E8),far(0x3826,0x0405),far(0x3CD5,0x000C),far(0x185C,0x065B),far(0x185C,0x06A8),far(0x28E0,0x06B4),far(0x28E0,0x098D),far(0x28E0,0x09A0),far(0x28E0,0x070D),far(0x28E0,0x078A),far(0x28E0,0x0857),far(0x28E0,0x08E2)} {
  a:=a
  o.OnCall(a,func(o *control){
   r:=o.Regs(); args:=[]uint16{}
   for i:=0;i<6;i++ { args=append(args,o.Word(far(r.SS,r.SP+4+uint16(i*2)))) }
   row:=map[string]any{"location":a.String(),"step":o.Steps(),"ax":o.m.CPU.R[cpu.AX],"args":args,
    "admiral12_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C+12*28),28)),
    "admiral28_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C+28*28),28)),
    "admiral40_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C+40*28),28))}
   if a==far(0x3826,0x0405) {
    b:=o.Bytes(far(args[1],args[0]),512)
    for i,v:=range b { if v==0 { b=b[:i]; break } }
    row["text_hex"]=hex.EncodeToString(b)
   }
   calls=append(calls,row)
  })
 }
 type testCase struct{fleet,level,exp,human,adj,pattern int}
 cases:=[]testCase{}
 for _,level:=range []int{0,1,2,26,27,28,29,33,34,38,39,98,99,100,255} {
  for _,delta:=range []int{-1,0,1} { cases=append(cases,testCase{12,level,delta,0,-1,0}) }
 }
 for _,pattern:=range []int{1,2,3,4} { cases=append(cases,testCase{12,27,0,0,-1,pattern}) }
 for _,f:=range []int{12,16} { for _,h:=range []int{0,1,2,3} { cases=append(cases,testCase{f,26,0,h,-1,0}) } }
 cases=append(cases,testCase{12,27,0,3,28,0},testCase{12,28,0,3,28,0},testCase{12,26,70000,3,-1,0})
 for _,tc:=range cases {
  o.reload(); calls=nil
  start:=o.Steps()
  for far(o.m.CPU.Seg[cpu.CS],o.m.CPU.IP)!=far(0x185C,0x060D) {
   if o.Steps()-start>1000000 { panic("自然入口預算耗盡") }; must(o.m.Step())
  }
  inject:=[]map[string]any{}
  set:=func(a addr,b []byte){ o.WriteBytes(a,b); if !bytes.Equal(o.Bytes(a,len(b)),b) { panic("回讀不符") }; inject=append(inject,map[string]any{"linear":fmt.Sprintf("lin:%X",a.Linear()),"hex":hex.EncodeToString(b)}) }
  set(far(0x54B3,0x94DC),[]byte{0x78,0x56,0x34,0x12})
  set(far(0x54B3,0x9894),[]byte{1}); set(far(0x54B3,0x9898),[]byte{byte(tc.fleet)})
  set(far(0x54B3,0x8F25),[]byte{0}); set(far(0x54B3,0x8F38),[]byte{0,0})
  controls:=[]byte{1,1}; for f:=0;f<2;f++ { if tc.human&(1<<f)!=0 { controls[f]=0 } }; set(far(0x54B3,0x8F21),controls)
  fr:=far(0x54B3,0xC074+uint16(tc.fleet*122))
  index:=int(o.Bytes(fr,122)[4]); set(far(fr.Seg,fr.Off+5),[]byte{byte(tc.adj)})
  need:=int(o.Word(far(0x54B3,0x4DBE+uint16(min(tc.level,39)*2))))
  exp:=need+tc.exp; if tc.exp==70000 { exp=70000 }
  ar:=far(0x54B3,0xB77C+uint16(index*28)); b:=make([]byte,4); binary.LittleEndian.PutUint32(b,uint32(exp)); set(far(ar.Seg,ar.Off+2),b); set(far(ar.Seg,ar.Off+6),[]byte{byte(tc.level)})
  if tc.pattern!=0 { v:=[]byte{0,0,199,200,255}[tc.pattern]; for _,off:=range []uint16{0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13,0x17} { set(far(ar.Seg,ar.Off+off),[]byte{v}) } }
  if tc.adj>=0 { a:=far(0x54B3,0xB77C+uint16(tc.adj*28)); set(far(a.Seg,a.Off+6),[]byte{26}); binary.LittleEndian.PutUint32(b,20000); set(far(a.Seg,a.Off+2),b) }
  tables:=func()map[string]string { return map[string]string{"admiral":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C),82*28)),"fleet":hex.EncodeToString(o.Bytes(far(0x54B3,0xC074),32*122)),"ship":hex.EncodeToString(o.Bytes(far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))),352*78))} }
  before:=tables(); start=o.Steps(); _,e:=o.CallBudget(40000000,far(0x185C,0x060D)); must(e)
  name:=fmt.Sprintf("fleet%d-level%d-exp%d-human%d-adj%d-pattern%d",tc.fleet,tc.level,exp,tc.human,tc.adj,tc.pattern)
  rows=append(rows,map[string]any{"name":name,"index":index,"case":[]int{tc.fleet,tc.level,exp,tc.human,tc.adj,tc.pattern},"need":need,"injections":inject,"before":before,"after":tables(),"calls":calls,"inputs":o.inputs,"frames":o.frames,"seed_after_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4)),"steps":o.Steps()-start})
 }
 result:=map[string]any{"schema":"logh3.battle-promotion-native.v1","program_sha256":digest(exe),"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","dosgolem_dirty":false,"go_version":runtime.Version(),"uid":os.Getuid(),"tick":20000,"seed":"0x12345678","command":"bash tools/battle_promotion_oracle.sh <a|b>","parent_state_sha256":"598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe","scope":"完整保存自然到8-4，入口明示注入；停8-5，非正常玩家GUI、非完整回合同態。","rows":rows}
 raw,e:=json.MarshalIndent(result,"","  "); must(e); f,e:=os.OpenFile(*output,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644); must(e); _,e=f.Write(append(raw,'\n')); must(e); must(f.Close()); fmt.Println("升級原版案例",len(rows),*output)
}
'''
p = Path('workplace/promotion79-r84-native-v2.go')
with p.open('x') as f:
    f.write(head + tail)
print(p)
