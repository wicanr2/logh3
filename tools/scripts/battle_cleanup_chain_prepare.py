"""由已凍結8-1觀測器生成四案8-1至8-4鏈；不改原有收據。"""
import hashlib
from pathlib import Path

src = Path("workplace/cleanup79-r88-native.go").read_bytes()
assert hashlib.sha256(src).hexdigest() == "673cb53134aa548b85a19ac1942bf54add25fcdc782d802f73663e0e49d5ad39"
s = src.decode().replace("0x02DA", "0x09BF")
s = s.replace("var calls []map[string]any", "var calls []map[string]any\n var promotionBefore map[string]string\n o.OnCall(far(0x185C,0x060D),func(o *control){promotionBefore=tables()})")
s = s.replace("[]int{12,16,17,28,29,30,31}", "[]int{12,29}")
s = s.replace("alternate:=0;for i:=1;i<count;i+=2{alternate|=1<<i}", "")
s = s.replace("[]int{0,2,1<<(count-1),(1<<count)-2,alternate}", "[]int{2,(1<<count)-2}")
s = s.replace("o.reload();calls=nil", "o.reload();calls=nil;promotionBefore=nil")
s = s.replace("before:=tables();start:=o.Steps()", "admiral:=int(f[4]);set(far(0x54B3,0xB77C+uint16(admiral*28+2)),[]byte{0x20,0x4e,0,0,28});before:=tables();start:=o.Steps()")
s = s.replace('"before":before,"after":tables()', '"before":before,"promotion_before":promotionBefore,"after":tables()')
s = s.replace('"logh3.battle-cleanup-native.v1"', '"logh3.battle-cleanup-chain.v1"')
s = s.replace('bash tools/battle_cleanup_oracle.sh <a|b>', 'bash tools/battle_cleanup_chain_oracle.sh <a|b>')
s = s.replace('固定8-1入口明示非旗艦艘數0，停8-2', '固定8-1入口明示非旗艦艘數0與一次足額功績，停8-5')
dest = Path("workplace/cleanup-chain79-r88-native.go")
with dest.open("x") as f:
    f.write(s)
runner = Path("tools/battle_cleanup_native_run.sh").read_text().replace("cleanup79-r88", "cleanup-chain79-r88")
with Path("workplace/cleanup-chain79-r88-run.sh").open("x") as f:
    f.write(runner)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
