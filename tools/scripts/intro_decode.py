"""照 `docs/re/36` 的語法解 `IBM16.350` 的影格。

    tools/py.sh tools/scripts/intro_decode.py <第一格> <格數> [輸出檔]

一格 ＝ 四個 plane，每個 plane 後面接一個分隔位元組。plane 的資料是
一串 token，寫進 80 × 227 的位元圖（一格一個 byte，**一欄一欄往下走**：
位移每次 `+80`，走到 `>= 18160` 就 `-18159` 換下一欄）。

token 的樣子（`docs/re/36`）：

| 樣子 | 意思 |
|---|---|
| `XX YY`（XX ≠ YY）| 把 `XX` XOR 進目前這一格，只吃掉 `XX` |
| `XX XX n`（`n` 有 bit7）| 把 `XX` XOR 進 `n & 0x7F` 格 |
| `XX XX n1 n2`（`n1` 沒有 bit7）| 同上，格數 ＝ `(n1 << 8) \\| n2` |
| `00 00 …` | 一樣的兩種計數，但只前進不寫 |

每個 plane 從 `di = 0` 開始，走完 `LIMIT` 格就停：**第一次是 0x100、
之後改成 0x50**（`mov word cs:[0248],0100` 之後被改寫）——那是換欄時的
終止界限，不是格數上限。

輸出是四個 plane 各 18160 bytes 的 XOR 累加（不含底圖），依序接在一起。
"""
import sys

PLANE = 18160          # 80 × 227
STRIDE = 80
WRAP = 18160
DATA = '/orig/IBM16.350'
LENS = '/orig/IBMFLS16.350'


def frame_bounds():
    fl = open(LENS, 'rb').read()
    lens = [fl[2 * i] | fl[2 * i + 1] << 8 for i in range(len(fl) // 2)]
    acc, out = 0, [0]
    for n in lens:
        acc += n
        out.append(acc)
    return out, lens


def decode_plane(data, pos, buf):
    """解一個 plane，回吃掉的位元組數。

    停的條件是 `di == limit`（原版是 `cmp di,cs:[0248] / jz`），**不是 >=**。
    `di` 每格 `+80`，一超過 18160 就 `-18159` ——換句話說換欄之後 `di` 就是
    欄號 1、2、3…，而 limit 在第一次換欄時被改成 0x50 ＝ 80 欄。
    一開始的 0x100 只是個「這一輪不會等於」的值。
    """
    di = 0
    limit = 0x100
    start = pos
    while True:
        if di == limit:
            return pos - start
        a = data[pos]
        b = data[pos + 1]
        pos += 1
        if b != a:
            buf[di] ^= a
            di += STRIDE
            if di >= WRAP:
                di -= WRAP - 1
                limit = 0x50
            continue

        pos += 1                      # 跳脫：吃掉第二個
        n = data[pos]
        pos += 1
        if n & 0x80:
            count = n & 0x7F
        else:
            count = (n << 8) | data[pos]
            pos += 1
        for _ in range(count):
            if a:
                buf[di] ^= a
            di += STRIDE
            if di >= WRAP:
                di -= WRAP - 1
                limit = 0x50


def main():
    first = int(sys.argv[1])
    count = int(sys.argv[2])
    out = sys.argv[3] if len(sys.argv) > 3 else None
    data = open(DATA, 'rb').read()
    bounds, lens = frame_bounds()
    planes = [bytearray(PLANE) for _ in range(4)]
    pos = bounds[first]
    for f in range(first, first + count):
        used = 0
        for p in range(4):
            used += decode_plane(data, pos + used, planes[p])
            used += 1                 # 分隔位元組
        if used != lens[f]:
            print(f'第 {f} 格：吃掉 {used}，長度表寫 {lens[f]}')
        pos += lens[f]
    if out:
        with open(out, 'wb') as fh:
            for p in planes:
                fh.write(p)
    nz = [sum(1 for v in p if v) for p in planes]
    print(f'第 {first}–{first + count - 1} 格解完，各 plane 非零位元組 {nz}')


if __name__ == '__main__':
    main()
