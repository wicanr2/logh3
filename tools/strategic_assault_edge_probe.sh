#!/usr/bin/env bash
# 只能在gui.sh xrun容器內執行；DRAFT副本，不改/src/internal。
set -euo pipefail
task_dir="$(mktemp -d /tmp/assault56-r99-model.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /src/go.mod /src/go.sum "$task_dir/"
cp -a /src/internal "$task_dir/"
ln -s /src/docs "$task_dir/docs"
cp /src/tools/scripts/strategic_assault_edge_model.go "$task_dir/model.go"
python3 - "$task_dir" <<'PY'
from pathlib import Path
import sys
root=Path(sys.argv[1])
p=root/'internal/rules/strategicassault.go'
s=p.read_text()
anchor='type StrategicPlacement struct {'
assert s.count(anchor)==1
s=s.replace(anchor,anchor+'\n Guard func([]byte) int',1)
anchor='strategicCounter(p, facings[i], ang, d, sf, attacker, r, res)'
assert s.count(anchor)==1
s=s.replace(anchor,'guard := 0\n if place != nil && place.Guard != nil { guard = place.Guard(p) }\n strategicCounter(p, facings[i], ang, d, sf, attacker, guard, r, res)',1)
anchor='d Fixed, sf StrategicFleet, attacker int, r *Rand,'
assert s.count(anchor)==1
s=s.replace(anchor,'d Fixed, sf StrategicFleet, attacker, guard int, r *Rand,',1)
anchor='base := int(binary.LittleEndian.Uint16(p[planetDefence:])) / div // 守軍 0'
assert s.count(anchor)==1
s=s.replace(anchor,'base := ((int(binary.LittleEndian.Uint16(p[planetDefence:])) + guard) & 65535) / div',1)
p.write_text(s)
p=root/'model.go';s=p.read_text()
anchor='place := &rules.StrategicPlacement{Slots: slots}'
assert s.count(anchor)==1
s=s.replace(anchor,anchor+r'''
            place.Guard = func(p []byte) int {
                count := int(binary.LittleEndian.Uint16(b["lists"][2:]))
                if count > (len(b["lists"])-0x26)/2 { panic("原版池清單越界") }
                total := 0
                radius := rules.FixedFromInt(rules.PlanetQueryRadius(int(binary.LittleEndian.Uint16(p[0x14:]))))
                for i:=0;i<count;i++ {
                    pool := int(binary.LittleEndian.Uint16(b["lists"][0x26+2*i:]))
                    sh := slot.ShipAt(pool)
                    if sh[0]!=p[0x11] || sh[0x1e]&128!=0 || sh[4]>1 { continue }
                    if rules.FixedDist(int(binary.LittleEndian.Uint16(p[0x46:])),int(binary.LittleEndian.Uint16(p[0x48:])),int(binary.LittleEndian.Uint16(sh[0x34:])),int(binary.LittleEndian.Uint16(sh[0x36:]))) > radius { continue }
                    total = min(65535,total+int(binary.LittleEndian.Uint16(sh[8:])))
                }
                return total/100
            }
''',1)
s=s.replace('"正式StrategicAssault的typed欄位比較；完整原池仍未驗證"','"DRAFT副本增加原版清單守軍；完整原池及正式Game仍未驗證"',1)
p.write_text(s)
PY
cd "$task_dir"
# 原版JSON及輸出均用呼叫端傳入的絕對/src/workplace路徑。
go run model.go -input "${1:?原版JSON絕對路徑}" -output "${2:?新比較JSON絕對路徑}"
