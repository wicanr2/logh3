# .AMG 影像解碼——照 GIN3PS.EXE 348A:02BA 那支常式寫的。
#
# 直行掃描：`lodsw` 取兩個 byte ＝ 同一 x 的相鄰兩列（`add di,80`）；
# 兩個 byte 相同時是重複組（後面跟著「第二個值」與「次數」）。
# 每個直行走 dl（＝ header 的 50）個 word ＝ 100 列，走完 x+1、dl 重設，
# 直到 dh（＝ header 的 10）個直行用完。四個 plane 的資料連續排在一起。
import struct, sys, json

def decode(img):
    """回傳 (寬, 高, 色號陣列, 用掉的位元組數)。

    四個 plane 是**同一個狀態機**：重複組的次數可以跨過 plane 邊界，
    程式在 dh 用完時只是換 map mask 再回到同一個迴圈（348A:0343）。
    把每個 plane 當成獨立一段解，跨界的那幾個 word 就會落在錯的地方。
    """
    w, h = struct.unpack_from('<HH', img, 8)
    rows = h * 2
    planes = [bytearray(w * rows) for _ in range(4)]
    pl, col, di, left = 0, 0, 0, h
    i = 12

    def put(a, b):
        nonlocal pl, col, di, left
        if pl > 3:
            return
        planes[pl][di] = a; di += w
        planes[pl][di] = b; di += w
        left -= 1
        if left == 0:
            col += 1; di = col; left = h
            if col == w:          # dh 用完 → 換 plane，di 回到起點
                pl += 1; col = 0; di = 0

    while pl <= 3 and i + 1 < len(img):
        al, ah = img[i], img[i+1]; i += 2
        if al != ah:
            put(al, ah)
        else:
            if i + 1 >= len(img):
                break
            v2, cnt = img[i], img[i+1]; i += 2
            for _ in range(cnt if cnt else 65536):
                if pl > 3:
                    break
                put(al, v2)

    px = bytearray(w * 8 * rows)
    for y in range(rows):
        for bx in range(w):
            for bit in range(8):
                v = 0
                for k in range(4):
                    if planes[k][y*w+bx] & (0x80 >> bit):
                        v |= 1 << k
                px[y*w*8 + bx*8 + bit] = v
    return w * 8, rows, bytes(px), i

if __name__ == '__main__':
    f = open('/orig/FACE.AMG', 'rb').read()
    idx = [struct.unpack_from('<HI', f, k*6) for k in range(83)]
    s, o = idx[0]
    W, H, px, used = decode(f[o:o+s])
    print(f'第 0 張：{W}x{H}，用掉 {used}/{s} bytes')
    print('色號分佈:', sorted(set(px)))
    for y in range(0, H, 6):
        print(''.join('.' if v == 0 else format(v, 'X') for v in px[y*W:y*W+W]))
