"""唯讀觀測 X11 的首戰「輸入」標籤；不注入遊戲狀態。"""
import ctypes as c
import sys
import time
from pathlib import Path
from pngout import game_palette

window, scale, template = int(sys.argv[1]), int(sys.argv[2]), Path(sys.argv[3])
pixels = (template / 'screen.bin').read_bytes()
palette = game_palette(template / 'screen.png')
assert len(pixels) == 640 * 480
lib = c.CDLL('libX11.so.6')
lib.XOpenDisplay.argtypes, lib.XOpenDisplay.restype = [c.c_char_p], c.c_void_p
lib.XGetImage.argtypes = [c.c_void_p, c.c_ulong, c.c_int, c.c_int,
                         c.c_uint, c.c_uint, c.c_ulong, c.c_int]
lib.XGetImage.restype = c.c_void_p
lib.XGetPixel.argtypes, lib.XGetPixel.restype = [c.c_void_p, c.c_int, c.c_int], c.c_ulong
lib.XDestroyImage.argtypes = [c.c_void_p]
lib.XCloseDisplay.argtypes = [c.c_void_p]
display = lib.XOpenDisplay(None)
assert display, '沒有 X11 display'
# 同盟軍「輸入」字樣。避開游標、動畫、日期及艦艇。
x0, y0, w, h = 574, 240, 36, 17
expected = []
for y in range(y0, y0 + h):
    for x in range(x0, x0 + w):
        i = pixels[y * 640 + x] * 3
        r, g, b = palette[i:i + 3]
        expected.append((r << 16) | (g << 8) | b)
assert len(set(expected)) > 1, '模板不含輸入標籤'
deadline = time.monotonic() + 60
try:
    while time.monotonic() < deadline:
        image = lib.XGetImage(display, window, x0 * scale, y0 * scale,
                             w * scale, h * scale, c.c_ulong(-1).value, 2)
        assert image, '無法讀取視窗'
        try:
            matches = all(lib.XGetPixel(image, (i % w) * scale, (i // w) * scale) == rgb
                          for i, rgb in enumerate(expected))
        finally:
            lib.XDestroyImage(image)
        if matches:
            print(f'正常首戰輸入標籤就緒：{scale}x，{len(expected)}像素相同', flush=True)
            break
        time.sleep(0.2)
    else:
        raise SystemExit('60秒內未到正常首戰輸入畫面')
finally:
    lib.XCloseDisplay(display)
