"""172：v1.6.5回歸基準。v1.6.4的六表，只把新遊戲WORK+0C..14換成規格159的九bytes。

poster_frame_gui的路徑是「新遊戲直接存檔」。v1.6.5刻意寫了這九bytes，其餘必須逐位元組相同。
用法：tools/py.sh tools/scripts/new_game_player_baseline.py
"""
import hashlib
import json
from pathlib import Path

ROOT = Path("/work/workplace")
SEAT = {
    "alliance": bytes([1, 0, 1, 0, 1, 1, 0, 31, 0]),
    "empire": bytes([0, 0, 1, 0, 1, 0, 1, 0, 31]),
}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def build(old, save, seat, out):
    saved = json.loads((old / "RESULT.json").read_text())["saved"]
    for name, digest in saved.items():
        assert sha((save / name).read_bytes()) == digest, name
    work = bytearray((save / "WORK.DT0").read_bytes())
    assert bytes(work[0x0C:0x15]) == bytes(9), "舊基準的座位bytes應該是零"
    work[0x0C:0x15] = seat
    expected = dict(saved, **{"WORK.DT0": sha(bytes(work))})
    out.write_text(json.dumps(dict(source=str(old), seat_control=seat.hex(" "), saved=expected), ensure_ascii=False, indent=2) + "\n")
    print(out.name, expected["WORK.DT0"][:16])


rel = ROOT / "release171"
out = ROOT / "release172"
build(rel / "player-linux", rel / "player-linux/config/logh3/save", SEAT["alliance"], out / "baseline-alliance.json")
build(rel / "player-empire", rel / "player-empire/config/logh3/save", SEAT["empire"], out / "baseline-empire.json")
win = ROOT / "music69-windows/full171/player-windows"
build(win, win / "save", SEAT["alliance"], ROOT / "music69-windows/full172/baseline-alliance.json")
