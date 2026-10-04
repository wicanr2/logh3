"""固定第88輪RAM導出8-5與勝方結算IDA輸入，不改正式.i64。"""
from pathlib import Path
import hashlib

ram = Path("workplace/oracle/cleanup79-r88-entry-a/mem-in.bin").read_bytes()
assert hashlib.sha256(ram).hexdigest() == "1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf"
for segment in (0x185c, 0x20b4):
    dest = Path(f"workplace/ida/finish79-r89-{segment:04X}.com")
    with dest.open("xb") as f:
        f.write(bytes(256) + ram[segment*16:segment*16+0x2000])
    print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
source = Path("tools/scripts/battle_promotion_ida.py").read_bytes()
assert hashlib.sha256(source).hexdigest() == "b3ca5d2829918d12aae07f492011613b8043e0d4180bd50d4b36291fdc51dbb7"
dest = Path("workplace/ida/finish79-r89-ida.py")
with dest.open("x") as f:
    f.write(source.decode().replace("promotion79-r84-", "finish79-r89-").replace("logh3.battle-promotion-ida.v1", "logh3.battle-finish-ida.v1"))
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
