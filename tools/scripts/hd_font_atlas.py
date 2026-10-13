#!/usr/bin/env python3
"""在固定 HD 工具映像內排版思源宋體／黑體，不讀取或放大原字模。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont, features


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, default=Path("/work/hd125/font"))
    parser.add_argument("--family", choices=["serif", "sans"], default="serif")
    args = parser.parse_args()
    out = args.output
    out.mkdir(parents=True, exist_ok=True)
    source = Path("/usr/share/fonts/opentype/noto") / ("NotoSansCJK-Regular.ttc" if args.family == "sans" else "NotoSerifCJK-Regular.ttc")
    license_source = Path("/usr/share/doc/fonts-noto-cjk/copyright")
    face = ImageFont.truetype(str(source), 42, index=3)
    assert "TC" in face.getname()[0], face.getname()
    unavailable = bytes(face.getmask("\uffff"))
    controls = "\0☺☻♥♦♣♠•◘○◙♂♀♪♫☼►◄↕‼¶§▬↨↑↓→←∟↔▲▼"
    chars = [(code, True, controls[code] if code < 32 else bytes([code]).decode("cp437"))
             for code in range(256)]
    for lead in range(0xA1, 0xFA):
        for trail in list(range(0x40, 0x7F)) + list(range(0xA1, 0xFF)):
            try:
                text = bytes([lead, trail]).decode("big5")
            except UnicodeDecodeError:
                continue
            if len(text) == 1:
                chars.append((lead << 8 | trail, False, text))
    glyphs, omitted, pages = [], [], []
    cw, ch, columns, rows = 48, 45, 64, 64
    page = None
    number = 0
    for code, half, text in chars:
        if text == "\0" or (not text.isspace() and bytes(face.getmask(text)) == unavailable):
            omitted.append({"code": code, "half": half, "unicode": text})
            continue
        if number % (columns * rows) == 0:
            if page is not None:
                name = f"atlas-{len(pages)}.png"
                page.save(out / name, optimize=False, compress_level=9)
                pages.append({"file": name, "sha256": digest(out / name)})
            page = Image.new("L", (cw * columns, ch * rows))
        slot = number % (columns * rows)
        x, y = slot % columns * cw, slot // columns * ch
        width = 24 if half else 48
        if text and not text.isspace():
            # 固定字格；少數寬符號縮小字級，不變更字格或字距。
            font = face
            box = font.getbbox(text)
            size = 42
            while box[2] - box[0] > width - 2 or box[3] - box[1] > ch - 2:
                size -= 1
                font = ImageFont.truetype(str(source), size, index=3)
                box = font.getbbox(text)
            px = x + (width - (box[2] - box[0])) // 2 - box[0]
            # 保留共同字型基線；高度異常的符號只在自己的字格內置中。
            py = y + 39 - font.getmetrics()[0]
            if py + box[1] < y or py + box[3] > y + ch:
                py = y + (ch - (box[3] - box[1])) // 2 - box[1]
            ImageDraw.Draw(page).text((px, py), text, font=font, fill=255)
            ink = page.crop((x, y, x + width, y + ch)).getbbox()
            assert ink and 0 <= ink[0] < ink[2] <= width and 0 <= ink[1] < ink[3] <= ch
        glyphs.append({"code": code, "half": half, "unicode": text,
                       "page": len(pages), "x": x, "y": y, "w": width, "h": ch})
        number += 1
    name = f"atlas-{len(pages)}.png"
    page.save(out / name, optimize=False, compress_level=9)
    pages.append({"file": name, "sha256": digest(out / name)})
    (out / "LICENSE.txt").write_bytes(license_source.read_bytes())
    manifest = {"schema": 1, "scale": 3, "font": face.getname(), "face_index": 3,
                "source": str(source), "source_sha256": digest(source),
                "license": "SIL Open Font License 1.1", "license_file": "LICENSE.txt",
                "license_sha256": digest(out / "LICENSE.txt"),
                "generator_sha256": digest(Path(__file__)),
                "freetype": features.version("freetype2"), "cell": [cw, ch],
                "pages": pages, "glyphs": glyphs, "fallback": omitted}
    (out / "MANIFEST.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"output": str(out), "glyphs": len(glyphs), "fallback": len(omitted),
                      "pages": len(pages), "source_sha256": digest(source)}, ensure_ascii=False))


if __name__ == "__main__":
    main()
