"""沿第89輪固定完整控制器，產生#56戰略入口的可丟棄原版探針。"""
from pathlib import Path
import hashlib

source = Path("workplace/finish79-r89-native.go").read_bytes()
assert hashlib.sha256(source).hexdigest() == "cbe370d25969cd96e78effc7bc50940c133f9fbab180b53eab4cf20d58ad5580"
s = source.decode()
a, b = s.index("func main() {"), s.index("type addr struct")
main = r'''func main() {
 output:=flag.String("output","","新原版JSON"); flag.Parse(); if *output=="" {panic("缺輸出")}
 exe,e:=os.ReadFile("/orig/GIN3.COM");must(e)
 if digest(exe)!="770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {panic("原版不符")}
 o:=&control{};defer o.Close()
 tables:=func()map[string]string{return map[string]string{
  "admiral":hex.EncodeToString(o.Bytes(far(0x54B3,0xB77C),82*28)),
  "fleet":hex.EncodeToString(o.Bytes(far(0x54B3,0xC074),32*122)),
  "ship":hex.EncodeToString(o.Bytes(far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))),352*78)),
  "work":hex.EncodeToString(o.Bytes(far(0x54B3,0x8F14),230)),
  "planet":hex.EncodeToString(append(o.Bytes(far(0x54B3,0xCFB8),64),o.Bytes(far(o.Word(far(0x54B3,0x5338)),o.Word(far(0x54B3,0x5336))),81*82)...)),
  "stellar":hex.EncodeToString(o.Bytes(far(0x54B3,0xA79C),32*126)),
  "lists":hex.EncodeToString(o.Bytes(far(0x54B3,0x9894),164))}}
 rows:=[]map[string]any{}
 for _,name:=range []string{"empty","near","out","far","terrain"} {
  o.reload();o.callbacks=make(map[addr]func(*control));calls:=[]map[string]any{};inject:=[]map[string]any{};var before map[string]string
  set:=func(a addr,b []byte){o.WriteBytes(a,b);if !bytes.Equal(o.Bytes(a,len(b)),b){panic("注入回讀不符")};inject=append(inject,map[string]any{"linear":fmt.Sprintf("lin:%X",a.Linear()),"hex":hex.EncodeToString(b)})}
  word:=func(a addr,v uint16){b:=make([]byte,2);binary.LittleEndian.PutUint16(b,v);set(a,b)}
  for _,id:=range []int{12,16,17,28,29,30,31} {
   p:=far(0x54B3,0xC074+uint16(id*122));v:=o.Bytes(p,1)[0]|1;if id==12{v&=0xfe};set(p,[]byte{v})
  }
  o.OnCall(far(0x20B4,0x1512),func(o *control){
   set(far(0x54B3,0x94DC),[]byte{0x78,0x56,0x34,0x12})
   base:=far(o.Word(far(0x54B3,0xCFFA)),o.Word(far(0x54B3,0xCFF8)));p:=base;found:=false
   for i:=0;i<int(o.Bytes(far(0x54B3,0xCFFC),1)[0]);i++ {candidate:=far(base.Seg,base.Off+uint16(i*82));if o.Bytes(far(candidate.Seg,candidate.Off+0x11),1)[0]==1{p=candidate;found=true;break}}
   if !found {panic("沒有同盟行星")}
   f:=o.Bytes(far(0x54B3,0xC074+16*122),122);pool:=uint16(0xffff)
   for i:=0;i<int(f[15]);i++ {v:=binary.LittleEndian.Uint16(f[16+i*4:]);sh:=far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))+v*78);if o.Bytes(far(sh.Seg,sh.Off+4),1)[0]<=1 {pool=v;break}}
   if pool==0xffff {panic("沒有守軍艦種")}
   sh:=far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))+pool*78)
   x,y:=o.Word(far(p.Seg,p.Off+0x46)),o.Word(far(p.Seg,p.Off+0x48));if name=="far" {x+=512}
   word(far(sh.Seg,sh.Off+0x34),x);word(far(sh.Seg,sh.Off+0x36),y)
   v:=o.Bytes(far(sh.Seg,sh.Off+0x1e),1)[0]&0x7f;if name=="out" {v|=0x80};set(far(sh.Seg,sh.Off+0x1e),[]byte{v})
   word(far(0x54B3,0x9896),1);word(far(0x54B3,0x98BA),pool)
   if name=="empty" || name=="terrain" {word(far(0x54B3,0x9896),0)}
   if name=="terrain" {
    // 34個實際敵方編制依序占據中心與前33個環狀候選；下一點由原始地形決定。
    dirs:=[8][2]int{{0,-32},{32,-32},{32,0},{32,32},{0,32},{-32,32},{-32,0},{-32,-32}};at:=0
    for _,id:=range []int{16,17,28,29,30,31} {f:=o.Bytes(far(0x54B3,0xC074+uint16(id*122)),122)
     for i:=0;i<int(f[15]);i++ {pool:=binary.LittleEndian.Uint16(f[16+i*4:]);sh:=far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))+pool*78);xx,yy:=1800,1800
      if at==0 {xx,yy=int(x),int(y)} else if at<=33 {k:=at-1;d:=dirs[k&7];xx,yy=int(x)+d[0]*(k/8+1),int(y)+d[1]*(k/8+1)}
      word(far(sh.Seg,sh.Off+0x34),uint16(xx));word(far(sh.Seg,sh.Off+0x36),uint16(yy));at++
     }
    }
    if at<34 {panic("地形診斷缺34組")}
   }
   before=tables()
  })
  for _,a:=range []addr{far(0x203A,0x04D1),far(0x17A0,0x00AC),far(0x1588,0x04EA)} {
   a:=a;o.OnCall(a,func(o *control){r:=o.Regs();args:=[]uint16{};for i:=0;i<6;i++{args=append(args,o.Word(far(r.SS,r.SP+4+uint16(2*i))))};ret:=far(o.Word(far(r.SS,r.SP+2)),o.Word(far(r.SS,r.SP)));c:=map[string]any{"location":a.String(),"step":o.Steps(),"args":args,"return_location":ret.String(),"entry_bytes_hex":hex.EncodeToString(o.Bytes(a,24))};calls=append(calls,c);o.OnCall(ret,func(o *control){c["return_ax"]=o.m.CPU.R[cpu.AX]})})
  }
  // 明示診斷入口及人工far-return，不模擬正常玩家操作。
  sp:=o.m.CPU.R[cpu.SP]-6;o.m.CPU.R[cpu.SP]=sp
  set(far(o.m.CPU.Seg[cpu.SS],sp),[]byte{0,0,0xff,0xff,26,0})
  o.m.CPU.Seg[cpu.CS]=0x20b4;o.m.CPU.IP=0x1497
  start:=o.Steps();ax,e:=o.CallBudget(40000000,far(0x20B4,0x1497));must(e);if before==nil{panic("未到1512")}
  rows=append(rows,map[string]any{"name":name,"region":26,"before":before,"after":tables(),"ax":ax,"calls":calls,"injections":inject,"entry_cs":0x20b4,"entry_ip":0x1497,"entry_sp":sp,"stop":"FFFF:0000","inputs":o.inputs,"frames":o.frames,"steps":o.Steps()-start,"seed_after_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4))})
 }
 result:=map[string]any{"schema":"logh3.strategic-assault-edge-native.v1","program_sha256":digest(exe),"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","dosgolem_dirty":false,"go_version":runtime.Version(),"uid":os.Getuid(),"tick":20000,"seed":"0x12345678","command":"bash tools/strategic_assault_edge_oracle.sh <a|b>","scope":"完整父保存，明示1497診斷入口；在1512固定seed及殘留池清單／位置，不能替代正常玩家路徑。","rows":rows}
 raw,e:=json.MarshalIndent(result,"","  ");must(e);f,e:=os.OpenFile(*output,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);_,e=f.Write(append(raw,'\n'));must(e);must(f.Close());fmt.Println("戰略入口原版案例",len(rows),*output)
}
'''
s = s[:a] + main + s[b:]
s = s.replace('if entry != far(0x185C, 0x001D)', 'if entry != far(0x20B4, 0x1497)')
s = s.replace('if a == far(0x185C, 0x001B)', 'if a == far(0xFFFF, 0x0000)')
dest = Path("workplace/assault56-r99-native-v3.go")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
