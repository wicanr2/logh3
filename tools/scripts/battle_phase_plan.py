"""#79固定第6回合的33圈邊界；來源見第三十四輪與研究339。

由既有dosgolem probe的-dump-at產生planar PNG及色號，不用-shots的mode13路徑。
執行前確認snapshot身份；輸出只准寫專用dosgolem的明確暫存子目錄。
"""
import hashlib
from pathlib import Path
import sys

name = sys.argv[1]
assert name in ("battle79-phase-r34-a", "battle79-phase-r34-b")
snapshot = Path("workplace/oracle/battle79-move-entry-a/machine.state")
assert hashlib.sha256(snapshot.read_bytes()).hexdigest() == "39ed892ae729c2b2bd1a0b208439d106e4913867e1b50134396fabd704b39b55"
COUNTS = [1213433576, 1213930450, 1214422221, 1214896450, 1215413352,
          1215922266, 1216433242, 1216954453, 1217467733, 1217982775,
          1218500238, 1219010770, 1219522850, 1220045146, 1220542460,
          1221023729, 1221534182, 1221981078, 1222429100, 1222794336,
          1223185615, 1223587666, 1223937618, 1224340181, 1224690440,
          1225085998, 1225493030, 1225799091, 1226023478, 1226209864,
          1226341266, 1226482918, 1226625851]
points = [("entry", 1213409021)] + [(f"loop-{i:02d}", n) for i, n in enumerate(COUNTS)] + [("end", 1226638239)]
prefix = f"/src/workplace/logh3-{name}"
flags = ["-dump-at", ";".join(f"{n}:{prefix}/{k}.png" for k, n in points),
         "-dump-mem-at", ";".join(f"{n}:lin:0:1048576:{prefix}/{k}-mem.bin" for k, n in points)]
out = Path("workplace") / (name + "-flags.txt")
if out.exists():
    raise FileExistsError(out)
out.write_text("\n".join(flags) + "\n")
