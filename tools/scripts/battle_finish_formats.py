"""從固定父RAM補110的六個原版報告格式，保留既有JSON欄位。"""
from pathlib import Path
import hashlib
import json

source = Path("workplace/oracle/cleanup79-r88-entry-a/mem-in.bin")
raw = source.read_bytes()
assert hashlib.sha256(raw).hexdigest() == "1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf"
target = Path("docs/internals/data/turn-window-panel.json")
assert target.stat().st_uid == target.parent.stat().st_uid == 1000
data = json.loads(target.read_text())
for key, offset in {
    "battle_lost_us": 0x3EA5,
    "battle_lost_ai": 0x3ED4,
    "battle_capture_prefix": 0x3EEE,
    "battle_capture_admiral": 0x3EF3,
    "battle_capture_ai": 0x3EFD,
    "battle_capture_tail": 0x3F09,
}.items():
    text = raw[0x54B30 + offset:].split(b"\0", 1)[0]
    data["report_formats"][key] = text.hex()
data["report_formats"]["battle_finish_note"] = (
    "185C:0A80／0B18，DS3EA5／3ED4／3EEE／3EF3／3EFD／3F09；"
    "DS線性54B30，固定cleanup79-r88-entry-a/mem-in.bin，SHA-256 "
    "1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf。"
    "原始運算元在finish79-r89-185C.json，研究423／規格110。"
)
target.write_text(json.dumps(data, ensure_ascii=False, indent=1) + "\n")
print("補六個戰場結束報告格式")
