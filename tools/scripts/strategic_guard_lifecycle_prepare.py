"""由固定第89輪控制器續跑原版退場與下一星域載圖；不注入CS/IP。"""
from pathlib import Path
import hashlib

source=Path('workplace/finish79-r89-native.go').read_bytes()
assert hashlib.sha256(source).hexdigest()=='cbe370d25969cd96e78effc7bc50940c133f9fbab180b53eab4cf20d58ad5580'
s=source.decode()
a,b=s.index('func main() {'),s.index('type addr struct')
main=r'''func main() {
 output:=flag.String("output","","新JSON");flag.Parse();if *output==""{panic("缺輸出")}
 exe,e:=os.ReadFile("/orig/GIN3.COM");must(e);if digest(exe)!="770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918"{panic("原版不符")}
 o:=&control{};defer o.Close();rows:=[]map[string]any{}
 for _,nextAssault:=range []bool{false,true} {
  o.reload();o.callbacks=make(map[addr]func(*control));o.writes=nil;events:=[]map[string]any{};injections:=[]map[string]any{}
  list:=func()map[string]any{return map[string]any{"count":o.Word(far(0x54B3,0x9896)),"bytes":hex.EncodeToString(o.Bytes(far(0x54B3,0x9894),164))}}
  record:=func(name string){r:=o.Regs();events=append(events,map[string]any{"name":name,"step":o.Steps(),"list":list(),"cs":o.m.CPU.Seg[cpu.CS],"ip":o.m.CPU.IP,"ss":r.SS,"sp":r.SP,"stack_hex":hex.EncodeToString(o.Bytes(far(r.SS,r.SP),32)),"ship_hex":hex.EncodeToString(o.Bytes(far(0x668B,0x0002),27456))})}
  set:=func(a addr,b []byte){o.WriteBytes(a,b);if !bytes.Equal(o.Bytes(a,len(b)),b){panic("注入回讀不符")};injections=append(injections,map[string]any{"linear":fmt.Sprintf("lin:%X",a.Linear()),"hex":hex.EncodeToString(b)})}
  record("parent-8-1")
  if nextAssault {
   f:=far(0x54B3,0xC074);raw:=o.Bytes(f,122);if raw[4]==255||raw[1]!=0{panic("艦隊0不是帝國有效艦隊")}
   set(far(f.Seg,f.Off+6),[]byte{27});set(f,[]byte{raw[0]&0xfe})
   star:=far(0x54B3,0xA79C+27*126+0x13);set(star,[]byte{o.Bytes(star,1)[0]|16})
  }
  o.OnCall(far(0x185C,0x09BF),func(o *control){record("8-5-before-poke");set(far(0x54B3,0x94DC),[]byte{0x78,0x56,0x34,0x12});for _,id:=range []int{12,16,17,28,29,30,31}{p:=far(0x54B3,0xC074+uint16(id*122));v:=o.Bytes(p,1)[0]&0xfe;if id==12{v|=1};set(p,[]byte{v})};record("8-5-after-poke")})
  for _,a:=range []addr{far(0x12AA,0x03BC),far(0x1588,0x00F6),far(0x1588,0x015B),far(0x20B4,0x1497),far(0x20B4,0x1512),far(0x185C,0x001B),far(0x0374,0x03E8)} {
   a:=a;o.OnCall(a,func(o *control){record(a.String());if a==far(0x12AA,0x03BC)||a==far(0x1588,0x00F6)||a==far(0x1588,0x015B)||a==far(0x20B4,0x1497){r:=o.Regs();ret:=far(o.Word(far(r.SS,r.SP+2)),o.Word(far(r.SS,r.SP)));o.OnCall(ret,func(o *control){record(a.String()+"-return")});if a==far(0x20B4,0x1497)&&nextAssault{if o.Bytes(far(r.SS,r.SP+4),1)[0]!=27{panic("不是下一星域27")};o.stop=ret}}})
  }
  o.stop=far(0x3705,0x023A)
  start:=o.Steps();ax,e:=o.CallBudget(60000000,far(0x185C,0x001D));record("stop");failure:="";if e!=nil{failure=e.Error()}
  rows=append(rows,map[string]any{"next_assault":nextAssault,"steps":o.Steps()-start,"ax":ax,"error":failure,"events":events,"writes":o.writes,"injections":injections,"inputs":o.inputs,"frames":o.frames})
 }
 result:=map[string]any{"schema":"logh3.strategic-guard-lifecycle.v1","program_sha256":digest(exe),"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","dosgolem_dirty":false,"go_version":runtime.Version(),"uid":os.Getuid(),"tick":20000,"command":"bash tools/strategic_guard_lifecycle_oracle.sh <a|b>","scope":"固定正常8-1保存續行，明示8-5敗方與seed、下一星域占領；沒有CS/IP注入或人工return。","rows":rows}
 raw,e:=json.MarshalIndent(result,"","  ");must(e);f,e:=os.OpenFile(*output,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);_,e=f.Write(append(raw,'\n'));must(e);must(f.Close());fmt.Println("戰後清單案例",len(rows),*output)
}
'''
s=s[:a]+main+s[b:]
s=s.replace(' "encoding/binary"\n','',1)
s=s.replace('    m *machine.Machine','    stop addr\n    writes []map[string]any\n    m *machine.Machine',1)
s=s.replace('if a == far(0x185C, 0x001B) {','if a == o.stop {',1)
anchor='        if err := o.m.Step(); err != nil { return 0, err }'
assert s.count(anchor)==1
s=s.replace(anchor,r'''
        previous:=o.Word(far(0x54B3,0x9896));instruction:=o.Bytes(a,12)
        if err := o.m.Step(); err != nil { return 0, err }
        current:=o.Word(far(0x54B3,0x9896));if previous!=current{o.writes=append(o.writes,map[string]any{"location":a.String(),"step":o.Steps()-1,"bytes_hex":hex.EncodeToString(instruction),"before":previous,"after":current,"linear":"lin:5E3C6"})}
''',1)
p=Path('workplace/assault56-r100-lifecycle-v6.go');assert p.parent.stat().st_uid==1000
with p.open('x') as f:f.write(s)
print(p,hashlib.sha256(p.read_bytes()).hexdigest())
