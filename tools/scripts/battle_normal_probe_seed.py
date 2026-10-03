"""#79正常入口探測的測試期Go覆蓋，不編輯正式產品檔案。

tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round44-normal-fixed-a
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re

parser = argparse.ArgumentParser()
parser.add_argument('--prefix', default='round44-normal')
args = parser.parse_args()
assert re.fullmatch(r'round[0-9]+-normal', args.prefix)
root = Path.cwd()
source = root / "internal/game/rebellion.go"
target = root / ("workplace/"+args.prefix+"-seed.go")
overlay = root / ("workplace/"+args.prefix+"-overlay.json")
assert source.is_file() and target.parent.stat().st_uid == os.getuid()
assert not target.exists() and not overlay.exists()
text = source.read_text()
assert "func (g *Game) rand94DC()" in text
text += """
// 僅供正常玩家探測；測試期覆蓋，不進正式產品。
func (g *Game) NormalProbeStrategySeed() uint32 {
    if g.strategyRand == nil { return strategyRandSeed }
    return g.strategyRand.Seed()
}
func (g *Game) NormalProbeSetStrategySeed(seed uint32) {
    g.rand94DC().SetSeed(seed)
}
"""
target.write_text(text)
overlay.write_text(json.dumps({"Replace": {str(source): str(target)}}, indent=2) + "\n")
print("原始產品未修改；來源SHA-256", hashlib.sha256(source.read_bytes()).hexdigest())
