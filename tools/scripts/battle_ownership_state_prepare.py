"""從固定dosgolem來源逐字抽出全部保存型別，產生一次性全欄位比較程式。"""
import hashlib
from pathlib import Path
import re
import sys

golem = Path('/golem')
registry, provenance = {}, {}
for folder in ('internal/machine', 'internal/dos'):
    for p in (golem / folder).glob('*.go'):
        if p.name.endswith('_test.go'):
            continue
        raw = p.read_bytes()
        text = raw.decode()
        for m in re.finditer(r'^type (\w+) struct\s*\{', text, re.M):
            start, depth, end = m.start(), 1, m.end()
            # 先去行註解，只用型別宣告內的實際大括號數定位。
            tail = re.sub(r'//[^\n]*', '', text[m.end():])
            for i, char in enumerate(tail):
                depth += (char == '{') - (char == '}')
                if depth == 0:
                    registry[m.group(1)] = text[m.start():m.end()] + tail[:i+1]
                    provenance[m.group(1)] = (str(p.relative_to(golem)), hashlib.sha256(raw).hexdigest())
                    break

todo, selected = ['machineState', 'dosState'], {}
while todo:
    name = todo.pop()
    if name in selected:
        continue
    definition = registry[name]
    selected[name] = definition
    todo.extend(token for token in re.findall(r'\b\w+\b', definition) if token in registry and token not in selected)

body = r'''
func must(err error) { if err != nil { panic(err) } }
func sum(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func load(path string) (machineState, dosState, string) {
 raw, err := os.ReadFile(path); must(err)
 zr, err := gzip.NewReader(bytes.NewReader(raw)); must(err)
 br := bufio.NewReader(zr)
 var m machineState; var d dosState
 section:=func()[]byte {
  var n uint64; must(binary.Read(br,binary.LittleEndian,&n))
  if n>16<<20 {panic("保存段長超界")}
  data:=make([]byte,int(n)); _,err:=io.ReadFull(br,data);must(err);return data
 }
 must(gob.NewDecoder(bytes.NewReader(section())).Decode(&m))
 must(gob.NewDecoder(bytes.NewReader(section())).Decode(&d))
 rest, err := io.ReadAll(br); must(err); if len(rest)!=0 { panic("保存檔有未讀取資料") }
 must(zr.Close())
 if m.Magic!="DOSGOLEM-M" || d.Magic!="DOSGOLEM-D" { panic("格式身分不符") }
 // 這兩張slice直接由原版SaveState的map range生成，以H識別，不是玩家次序。
 sort.Slice(d.Handles,func(i,j int)bool{return d.Handles[i].H<d.Handles[j].H})
 sort.Slice(d.EMSHandles,func(i,j int)bool{return d.EMSHandles[i].H<d.EMSHandles[j].H})
 return m,d,sum(raw)
}
func main() {
 names:=[]string{"entry","control","capture0","capture1","partial0","partial1","capture0-second","capture1-second","capture0-finish","capture1-finish","capture0-complete","capture1-complete","partial0-complete","partial1-complete"}
 rows:=[]map[string]any{}
 for _,name:=range names {
  file:="end.state"; if name=="entry" {file="entry.state"}
  a:=fmt.Sprintf("workplace/oracle/ownership79-r82-%s-a/%s",name,file)
  b:=fmt.Sprintf("workplace/oracle/ownership79-r82-%s-b/%s",name,file)
  ma,da,ha:=load(a); mb,db,hb:=load(b)
  if !reflect.DeepEqual(ma,mb) || !reflect.DeepEqual(da,db) {panic("全部保存欄位不同："+name)}
  canonical,err:=json.Marshal([]any{ma,da}); must(err)
  control:=ma; control.Mem=append([]byte(nil),ma.Mem...); control.Mem[0]^=1
  if reflect.DeepEqual(control,mb) {panic("RAM一byte負對照未攔截")}
  rows=append(rows,map[string]any{"name":name,"raw_sha256_a":ha,"raw_sha256_b":hb,"decoded_all_fields_equal":true,"canonical_sha256":sum(canonical),"negative_control_passed":true,"machine_fields":reflect.TypeOf(ma).NumField(),"dos_fields":reflect.TypeOf(da).NumField()})
 }
 out:=map[string]any{"schema":"logh3.battle-ownership-state-compare.v1","rows":rows,"method":"讀取固定f7523f9完整machineState及dosState型別；maps按鍵比較，僅將map生成的Handles／EMSHandles按H排序，其他欄位全部DeepEqual；一byteRAM負對照必須失敗"}
 data,err:=json.MarshalIndent(out,"","  "); must(err)
 f,err:=os.OpenFile("workplace/ownership79-r82-state-compare.json",os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644); must(err)
 _,err=f.Write(append(data,'\n'));must(err);must(f.Close())
 fmt.Println("完整保存欄位相同",len(rows),"機器",rows[0]["machine_fields"],"DOS",rows[0]["dos_fields"])
}
'''
imports = '"bufio"; "bytes"; "compress/gzip"; "crypto/sha256"; "encoding/binary"; "encoding/gob"; "encoding/hex"; "encoding/json"; "fmt"; "io"; "os"; "reflect"; "sort"'
revision = sys.argv[1]
assert re.fullmatch(r'v[0-9]+', revision)
target = Path(f'/task/workplace/ownership79-r82-state-compare-{revision}.go')
body = body.replace('ownership79-r82-state-compare.json', f'ownership79-r82-state-compare-{revision}.json')
target.open('x').write('package main\nimport ('+imports+')\n'+'\n'.join(selected.values())+'\n'+body)
meta = Path(f'/task/workplace/ownership79-r82-state-schema-{revision}.txt')
meta.open('x').write('\n'.join(f'{name} {provenance[name][0]} {provenance[name][1]}' for name in sorted(selected))+'\n')
print('逐字保存型別', sorted(selected), '輸出', target)
