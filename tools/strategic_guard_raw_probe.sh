#!/usr/bin/env bash
# gui.sh xrun容器中的可丟棄原池探針，不改正式internal。
set -euo pipefail
task_dir="$(mktemp -d /tmp/strategic-guard-raw.XXXXXX)"
trap 'rm -rf "$task_dir"' EXIT
cp /src/go.mod /src/go.sum "$task_dir/"
cp -a /src/internal "$task_dir/"
cp /src/tools/scripts/strategic_guard_raw_probe.go "$task_dir/probe.go"
ln -s /src/docs "$task_dir/docs"
python3 - "$task_dir" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1])/'internal/rules/battlefinish.go'
s=p.read_text()
a=s.index('\tvar seen [ShipCount]bool')
b=s.index('\t// 每次戰損重算全陣營費用',a)
s=s[:a]+'''\tfor _, pool := range poolIDs {
        if pool<0||pool>=ShipCount{return nil,fmt.Errorf("保留池越界")}
    }
'''+s[b:]
p.write_text(s)
PY
cd "$task_dir"
go run probe.go -input "${1:?原版JSON}" -output "${2:?新輸出JSON}"
