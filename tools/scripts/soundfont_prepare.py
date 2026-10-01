"""#69：固定來源下載／SHA-256 核對／音色庫暫存；不是玩家轉檔器。

入口與權利邊界：docs/re/356-player-music-library.md。
預設只用快取；下載須同時明示 --download 與包裝器 LOGH3_NET=1。
不解壓任意路徑，只讀已鎖定的三個 regular file。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import tarfile
import urllib.request


def sha256(path):
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def owned(path):
    st = path.lstat()
    if path.is_symlink() or st.st_uid != os.getuid() or st.st_gid != os.getgid():
        raise ValueError(f"來源快取或輸出擁有權／形態不符：{path}")


def main():
    ap = argparse.ArgumentParser(description="準備已授權音色庫；不含原版遊戲樂曲")
    ap.add_argument("--cache", type=Path, default=Path("/work/workplace/music69-source"))
    ap.add_argument("--download", action="store_true")
    args = ap.parse_args()
    doc = json.loads(Path("/work/docs/internals/data/player-music-library.json").read_text())
    sf = doc["soundfont"]
    archive = sf["archive"]
    cache = args.cache
    cache.mkdir(parents=True, exist_ok=True)
    owned(cache)
    source = cache / archive["file"]
    if source.exists():
        owned(source)
    else:
        if not args.download:
            raise ValueError("缺固定版本來源快取；下載時需明示 --download 與 LOGH3_NET=1")
        tmp = source.with_suffix(source.suffix + ".partial")
        if tmp.exists():
            owned(tmp)
        try:
            with urllib.request.urlopen(archive["url"], timeout=60) as src, tmp.open("wb") as out:
                total = 0
                while block := src.read(1024 * 1024):
                    total += len(block)
                    if total > archive["bytes"]:
                        raise ValueError("下載超出固定 archive 長度")
                    out.write(block)
            if tmp.stat().st_size != archive["bytes"] or sha256(tmp) != archive["sha256"]:
                raise ValueError("下載來源長度／SHA-256 不符，拒絕使用")
            tmp.replace(source)
        finally:
            if tmp.exists():
                tmp.unlink()
    if source.stat().st_size != archive["bytes"] or sha256(source) != archive["sha256"]:
        raise ValueError("快取來源長度／SHA-256 不符，拒絕使用")
    output = cache / "prepared"
    output.mkdir(exist_ok=True)
    owned(output)
    with tarfile.open(source, "r:gz") as tar:
        for member, expected in archive["member_sha256"].items():
            entry = tar.getmember(member)
            if not entry.isfile():
                raise ValueError(f"鎖定的來源成員不是 regular file：{member}")
            dest = output / Path(member).name
            if dest.exists():
                owned(dest)
            with tar.extractfile(entry) as src, dest.open("wb") as out:
                shutil.copyfileobj(src, out)
            owned(dest)
            if sha256(dest) != expected:
                raise ValueError(f"音色庫／授權成員 SHA-256 不符：{member}")
    font = output / sf["file"]
    if font.stat().st_size != sf["bytes"] or sha256(font) != sf["sha256"]:
        raise ValueError("SF2 與既有合成器音色庫不一致")
    print(f"已驗固定 archive、SF2、完整授權與貢獻者文件：{output}")


if __name__ == "__main__":
    main()
