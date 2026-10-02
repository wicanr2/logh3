"""第39輪第三階段兩分支的固定邊界；只產生新批次。"""
import hashlib
from pathlib import Path
import sys

HEAD = "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c"
POINTS = {
    "s1": [("entry", 1265700575), ("label", 1265720899), ("seat1", 1267737608), ("end", 1270680842)],
    "s3": [("entry", 1226638239), ("label", 1226658563), ("discovered", 1226956594),
           ("header", 1227120799), ("unit-0", 1245284656), ("unit-1", 1246588790),
           ("unit-2", 1248008676), ("unit-3", 1249412199), ("unit-4", 1250816750),
           ("unit-5", 1252204864), ("unit-6", 1253599326), ("unit-7", 1255007656),
           ("seat1", 1256220182), ("end", 1256470038)],
}
PARENTS = {
    "s1": ("battle79-round36-s1-boundary-explore/end.state", "a0c693ec71ab63190c43d36842c9fdc6540b7e4705602b860ca0b2f1a334d52b"),
    "s3": ("battle79-move-entry-a/machine.state", "39ed892ae729c2b2bd1a0b208439d106e4913867e1b50134396fabd704b39b55"),
}
def main():
    name = sys.argv[1]
    assert name in ("battle79-spot-r39-s1-a", "battle79-spot-r39-s1-b", "battle79-spot-r39-s3-a", "battle79-spot-r39-s3-b")
    scenario = name.split("-")[-2]
    parent, digest = PARENTS[scenario]
    assert hashlib.sha256((Path("workplace/oracle") / parent).read_bytes()).hexdigest() == digest
    prefix = "/src/workplace/logh3-" + name
    flags = ["-dump-at", ";".join(f"{step}:{prefix}/{label}.png" for label, step in POINTS[scenario]),
             "-dump-mem-at", ";".join(f"{step}:lin:0:1048576:{prefix}/{label}-mem.bin" for label, step in POINTS[scenario])]
    p = Path("workplace") / (name + "-flags.txt")
    assert not p.exists()
    p.write_text("\n".join(flags) + "\n")


if __name__ == "__main__":
    main()
