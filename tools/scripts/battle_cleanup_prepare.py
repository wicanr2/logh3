"""由固定第86輪控制器及新8-1父保存生成窄幅原版觀測器。Docker內執行。"""
import hashlib
from pathlib import Path


def digest(b):
    return hashlib.sha256(b).hexdigest()


source = Path("workplace/promotion79-r86-global.go").read_bytes()
assert digest(source) == "895f92e7ac5f3431569f2b84cf5bf342114e249f87051f6d07345fafd6573b8b"
parent = Path("workplace/oracle/cleanup79-r88-entry-a")
ram = (parent / "mem-in.bin").read_bytes()
assert len(ram) == 1048576
state = (parent / "entry.state").read_bytes()
control = source.decode().split("type addr struct", 1)[1]
control = "type addr struct" + control
control = control.replace("ownership79-r82-entry-a", "cleanup79-r88-entry-a")
control = control.replace("598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe", digest(state))
control = control.replace("b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900", digest(ram))
control = control.replace("0x060D", "0x001D")
control = control.replace("0x09BF", "0x02DA")
imports = source.decode().split("func digest", 1)[0]
main = r'''
func digest(b []byte) string { h:=sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func must(e error) { if e!=nil { panic(e) } }
func main() {
 output:=flag.String("output","","新原版JSON"); flag.Parse()
 if *output=="" {panic("缺輸出")}
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
  "lists":hex.EncodeToString(o.Bytes(far(0x54B3,0x9894),2+2+32+128))}}
 rows:=[]map[string]any{}
 var calls []map[string]any
 for _,a:=range []addr{far(0x185C,0x0177),far(0x185C,0x01ED),far(0x185C,0x020C),far(0x2AD2,0x01EF),far(0x1727,0x0298),far(0x1261,0x0104),far(0x12AA,0x03BC)} {
  a:=a;o.OnCall(a,func(o *control){r:=o.Regs();args:=[]uint16{};for i:=0;i<4;i++{args=append(args,o.Word(far(r.SS,r.SP+4+uint16(i*2))))}
   calls=append(calls,map[string]any{"location":a.String(),"step":o.Steps(),"args":args,"entry_bytes_hex":hex.EncodeToString(o.Bytes(a,24))})})
 }
 for _,fleet:=range []int{12,16,17,28,29,30,31} {
  o.reload();count:=int(o.Bytes(far(0x54B3,0xC074+uint16(fleet*122)),122)[15]);if count<2{panic("固定艦隊不足兩組")}
  alternate:=0;for i:=1;i<count;i+=2{alternate|=1<<i}
  for _,mask:=range []int{0,2,1<<(count-1),(1<<count)-2,alternate} {
   o.reload();calls=nil
   if far(o.m.CPU.Seg[cpu.CS],o.m.CPU.IP)!=far(0x185C,0x001D){panic("父保存不是8-1入口")}
   inject:=[]map[string]any{}
   set:=func(a addr,b []byte){o.WriteBytes(a,b);if !bytes.Equal(o.Bytes(a,len(b)),b){panic("注入回讀不符")};inject=append(inject,map[string]any{"linear":fmt.Sprintf("lin:%X",a.Linear()),"hex":hex.EncodeToString(b)})}
   set(far(0x54B3,0x94DC),[]byte{0x78,0x56,0x34,0x12})
   set(far(0x54B3,0x8F21),[]byte{1,1})
   set(far(0x54B3,0x8F38),[]byte{0,0})
   f:=o.Bytes(far(0x54B3,0xC074+uint16(fleet*122)),122)
   for i:=0;i<count;i++ {if mask&(1<<i)!=0 {pool:=binary.LittleEndian.Uint16(f[0x10+i*4:]);set(far(o.Word(far(0x54B3,0xCFB6)),o.Word(far(0x54B3,0xCFB4))+pool*78+8),[]byte{0,0})}}
   before:=tables();start:=o.Steps();_,e:=o.CallBudget(40000000,far(0x185C,0x001D));must(e)
   rows=append(rows,map[string]any{"name":fmt.Sprintf("fleet%d-count%d-mask%X",fleet,count,mask),"fleet":fleet,"count":count,"mask":mask,"before":before,"after":tables(),"calls":calls,"injections":inject,"inputs":o.inputs,"frames":o.frames,"steps":o.Steps()-start,"seed_after_hex":hex.EncodeToString(o.Bytes(far(0x54B3,0x94DC),4))})
  }
 }
 result:=map[string]any{"schema":"logh3.battle-cleanup-native.v1","program_sha256":digest(exe),"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","dosgolem_dirty":false,"go_version":runtime.Version(),"uid":os.Getuid(),"tick":20000,"seed":"0x12345678","command":"bash tools/battle_cleanup_oracle.sh <a|b>","scope":"固定8-1入口明示非旗艦艘數0，停8-2；完整原池及呼叫序列，非正常玩家全流程。","rows":rows}
 raw,e:=json.MarshalIndent(result,"","  ");must(e);file,e:=os.OpenFile(*output,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);_,e=file.Write(append(raw,'\n'));must(e);must(file.Close());fmt.Println("清空分隊原版案例",len(rows),*output)
}
'''
dest = Path("workplace/cleanup79-r88-native.go")
with dest.open("x") as f:
    f.write(imports + main + control)
print(dest, digest(dest.read_bytes()), "父state", digest(state), "父RAM", digest(ram))
com = Path("workplace/ida/cleanup79-r88-185C.com")
with com.open("xb") as f:
    f.write(bytes(256) + ram[0x185c0:0x185c0+0x1000])
print(com, digest(com.read_bytes()))

ida = Path("tools/scripts/battle_promotion_ida.py").read_bytes()
assert digest(ida) == "b3ca5d2829918d12aae07f492011613b8043e0d4180bd50d4b36291fdc51dbb7"
dest = Path("workplace/ida/cleanup79-r88-ida.py")
with dest.open("x") as f:
    f.write(ida.decode().replace("promotion79-r84-", "cleanup79-r88-").replace("logh3.battle-promotion-ida.v1", "logh3.battle-cleanup-ida.v1"))
print(dest, digest(dest.read_bytes()))
