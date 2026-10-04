#!/usr/bin/env bash
# 只在gui.sh xrun內執行；110 DRAFT，不改正式rules。
set -euo pipefail
task_dir="$(mktemp -d /tmp/finish79-r91-model.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
mkdir "$task_dir/internal"
mkdir "$task_dir/internal/asset"
cp /src/internal/asset/cmp.go "$task_dir/internal/asset/"
cp /src/go.mod /src/go.sum "$task_dir/"
cp -a /src/internal/rules "$task_dir/internal/"
python3 -c '
from pathlib import Path
import hashlib,sys
p=Path(sys.argv[1],"internal/rules/strategicassault.go")
s=p.read_bytes()
assert hashlib.sha256(s).hexdigest()=="1d264d86e5b285c7e2187f9ba0fdaaf98f2e291189726a8e8e87444b7d074b4e"
t=s.decode()
replacements=[
("fleets []StrategicFleet, r *Rand, place *StrategicPlacement) StrategicAssaultResult {", "fleets []StrategicFleet, r *Rand, place *StrategicPlacement) StrategicAssaultResult {\n return probeWinnerAssault(planets,sites,attacker,fleets,r,place,0)\n}\nfunc probeWinnerAssault(planets []byte, sites []PlanetSite, attacker int, fleets []StrategicFleet, r *Rand, place *StrategicPlacement, startRound int) StrategicAssaultResult {"),
("round, passes, planet := 0, 3, 0", "round, passes, planet := startRound & 255, 3, 0"),
("import \"encoding/binary\"", "import \"encoding/binary\"\nvar probeGuard = func(p []byte) int { return 0 }"),
("base := int(binary.LittleEndian.Uint16(p[planetDefence:])) / div // 守軍 0", "base := (int(binary.LittleEndian.Uint16(p[planetDefence:])) + probeGuard(p)) & 65535\n base /= div"),
("\t}\n\tfor i, sf := range fleets {\n\t\tf := sf.Fleet", "\t\tprobeAfterPlace(sf, pi, facings[i])\n\t}\n\tfor i, sf := range fleets {\n\t\tf := sf.Fleet"),
("g.Ships -= lost\n\tstrategicMorale(sf)", "g.Ships -= lost\n\tstrategicMorale(sf)\n\tprobeAfterLoss(sf,gi,lost)")]
for old,new in replacements:
 assert t.count(old)==1,old
 t=t.replace(old,new,1)
p.write_text(t)
' "$task_dir"
cp /src/workplace/finish79-r91-model-test.go "$task_dir/internal/rules/finish_raw_probe_test.go"
cp /src/tools/scripts/battle_finish_raw_model.go "$task_dir/internal/rules/finish_raw_model.go"
cd "$task_dir"
go test ./internal/rules -run '^TestDraftWinnerAssault$' -count=1 -v
