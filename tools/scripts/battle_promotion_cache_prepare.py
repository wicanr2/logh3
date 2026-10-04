"""沿凍結第86輪觀測器，只加兩個明示歸屬變更診斷。"""
from pathlib import Path
import hashlib

p = Path('workplace/promotion79-r86-global.go')
b = p.read_bytes()
assert hashlib.sha256(b).hexdigest() == '895f92e7ac5f3431569f2b84cf5bf342114e249f87051f6d07345fafd6573b8b'
s = b.decode()
assert s.count('for _,tc:=range cases {') == 1
s = s.replace('for _,tc:=range cases {', 'cases=[]testCase{{12,26,0,0,-1,8},{16,26,0,0,-1,9}}\n for _,tc:=range cases {')
s = s.replace('if tc.pattern>=5 {', 'if tc.pattern>=5 && tc.pattern<=7 {')
s = s.replace('before:=tables();', '''if tc.pattern>=8 {
   region,owner:=13,byte(1)
   if tc.pattern==9 { region,owner=28,byte(0) }
   set(far(0x54B3,0xA79C+uint16(region*126+0x12)),[]byte{owner})
  }
  before:=tables();''')
s = s.replace('logh3.battle-promotion-global-native.v1', 'logh3.battle-promotion-cache-native.v1')
s = s.replace('bash tools/battle_promotion_global_oracle.sh <a|b>', 'bash tools/battle_promotion_cache_oracle.sh <a|b>')
dest = Path('workplace/promotion79-r87-cache.go')
with dest.open('x') as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())

model = Path('tools/scripts/battle_promotion_global_probe.go').read_bytes()
assert hashlib.sha256(model).hexdigest() == '0af869de9fa1f669d5c498e3f15450f91d00e72a4550ff3f64bd70ecdc05e129'
s = model.decode().replace('promotion79-r86-global-probe-v3.json', 'promotion79-r87-cache-model-v2.json')
s = s.replace('promotion79-r86-global-a.json', 'promotion79-r87-cache-a.json').replace('promotion79-r86-global-b.json', 'promotion79-r87-cache-b.json')
s = s.replace('owners[i] = rules.RegionOwnerOf(stellar[i*126 : (i+1)*126])', 'owners[i] = int(m[ds+0xa79c+i*126+0x12])')
s = s.replace('paths := graph.UnrestPaths(owners)', '''paths := graph.UnrestPaths(owners)
        currentOwners:=make([]int,32)
        for i:=range currentOwners { currentOwners[i]=rules.RegionOwnerOf(stellar[i*126:(i+1)*126]) }
        rebuilt:=graph.UnrestPaths(currentOwners)
        rebuiltDifferent:=0
        for f:=0;f<2;f++ { for i:=0;i<32;i++ { for j:=0;j<32;j++ {
            if paths.Dist[f][i][j]!=rebuilt.Dist[f][i][j] {rebuiltDifferent++}
            if paths.Mid[f][i][j]!=rebuilt.Mid[f][i][j] {rebuiltDifferent++}
        } } }
        if rebuiltDifferent==0 && currentOwners[13]!=owners[13] {panic("歸屬注入未形成可區分的路徑反對照")}''')
s = s.replace('"planet_sha256": digest(planetInput), "admiral_input_sha256"', '"planet_sha256": digest(planetInput), "premature_rebuilt_different_entries":rebuiltDifferent,"admiral_input_sha256"')
s = s.replace('logh3.battle-promotion-global-probe.v3', 'logh3.battle-promotion-cache-model.v2')
dest = Path('workplace/promotion79-r87-cache-model-v2.go')
with dest.open('x') as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
