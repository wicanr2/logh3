"""印存檔某個檔案的幾筆記錄（十六進位 ＋ Big5 試讀）。

    save_dump.py <檔案> <記錄長度> <起始筆> <幾筆>
"""
import sys

path, rec, start, n = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
d = open(path, "rb").read()
for i in range(start, start + n):
    r = d[i * rec:(i + 1) * rec]
    print("#%d" % i)
    for o in range(0, rec, 16):
        chunk = r[o:o + 16]
        try:
            txt = chunk.decode("big5").replace("\x00", ".")
        except Exception:
            txt = ""
        print("  +%02X  %-47s  %s" % (o, chunk.hex(" "), txt))
