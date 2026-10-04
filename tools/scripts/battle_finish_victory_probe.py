"""第94輪S3正常玩家勝負探針，沿固定seed，禁止覆寫舊收據。"""
import hashlib
import ast
import json
from pathlib import Path
import subprocess
import sys

revision = sys.argv[1]
assert revision.isalnum()
source = Path("tools/scripts/battle_finish_player_probe.py")
expected = json.loads(Path("docs/internals/data/fleet-order-action.json").read_text())["sources"][str(source)]["sha256"]
assert hashlib.sha256(source.read_bytes()).hexdigest() == expected
s = source.read_text()
old = 'replace("click(300, 322)", "click(300, 282) // 公開S1列：第8列，122+8*20。")\n'
assert s.count(old) == 1
s = s.replace(old, "")
old = 'replace("u.CruiseSpeed > g.battleUnits[lead].CruiseSpeed", "u.Ships > g.battleUnits[lead].Ships")\n'
assert s.count(old) == 1
s = s.replace(old, "")
# S3先沿已通過正常玩家測試的鏡頭與直線通路策略。
# 第93輪這三項替換是S1探索，不能當成S3改善的既定事實。
lines = s.splitlines(keepends=True)
remove = []
for node in ast.parse(s).body:
    if not isinstance(node, ast.Expr) or not isinstance(node.value, ast.Call):
        continue
    call = node.value
    if not isinstance(call.func, ast.Name) or call.func.id != "replace":
        continue
    first = call.args[0].value if isinstance(call.args[0], ast.Constant) else ""
    if "for _, v := range g.battleUnits" in first or first == "v.Squadron != 0" or "Terrain.LineOpen" in first:
        remove.append((node.lineno-1, node.end_lineno))
assert len(remove) == 3
for start, end in reversed(remove):
    del lines[start:end]
s = "".join(lines)
old = 'for row<2&&candidate.Mask&(1<<row)==0 {row++}'
assert s.count(old) == 1
s = s.replace(old, 'row=2;for row>0&&candidate.Mask&(1<<row)==0 {row--}')
s = s.replace('"scenario":"S1"', '"scenario":"S3"')
s = s.replace("finish79-r93", "finish79-r94")
path = Path(f"workplace/finish79-r94-generator-{revision}.py")
assert not path.exists()
path.write_text(s)
subprocess.run(["python3", str(path), revision], check=True)
# 每支旗艦都經公開右鍵捲動後下令，避免只跟最快艦隊的探針偏差。
go = Path(f"workplace/finish79-r94-{revision}-probe_test.go")
text = go.read_text()
start = text.index("\t\t\t\tfor i, u := range g.battleUnits {\n\t\t\t\t\tif i != lead")
end = text.index("\t\t\t\t// 只下目前鏡頭可點的旗艦。", start)
camera = text[start:end]
begin = camera.index("\t\t\t\t\twant :=")
finish = camera.rindex("\t\t\t\t\tbreak\n")
camera = camera[begin:finish]
text = text[:start] + text[end:]
anchor = "\t\t\t\t\tx, y := unitScreenPos(g, i)"
assert text.count(anchor) == 1
text = text.replace(anchor, camera + anchor)
# fld 在同一層只宣告一次。
anchor = "\t\t\t\t\tx, y := unitScreenPos(g, i)\n\t\t\t\t\tfld := g.res.Battle.Field"
assert text.count(anchor) == 1
text = text.replace(anchor, "\t\t\t\t\tx, y := unitScreenPos(g, i)")
go.write_text(text)
