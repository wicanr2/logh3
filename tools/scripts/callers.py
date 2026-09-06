"""在執行期記憶體 dump 上做三件事：倒 hex、找函式起點、找誰呼叫它。

程式碼段的段值每次執行都可能不同，但**同一份 dump 內**線性位址是穩定的，
所以這裡一律用線性位址講話，段值只在組 `call far` 簽章時用一次。

用法：
  tools/py.sh tools/scripts/callers.py hex   <dump> <線性hex> <長度>
  tools/py.sh tools/scripts/callers.py entry <dump> <線性hex>          # 往前找函式起點
  tools/py.sh tools/scripts/callers.py xref  <dump> <線性hex> <段hex>  # 找 call far 到它的地方
"""
import pathlib, sys

def load(p):
    return pathlib.Path(p).read_bytes()

def hexdump(mem, addr, n):
    for base in range(addr, addr + n, 16):
        row = mem[base:base+16]
        print('%05X  %s' % (base, ' '.join('%02X' % b for b in row)))

def entry(mem, addr):
    """往前找最近的 retf(CB)／ret(C3)，其後第一個 push bp 家族就是函式起點。"""
    for i in range(addr, max(0, addr - 4096), -1):
        if mem[i] in (0xCB, 0xC3):
            j = i + 1
            while j < addr and mem[j] in (0x55, 0x56, 0x57, 0x53):   # push bp/si/di/bx
                if mem[j] == 0x55:
                    return j
                j += 1
            return i + 1
    return None

def xref(mem, addr, seg):
    off = addr - seg * 16
    if not 0 <= off <= 0xFFFF:
        print('位址 %05X 不在段 %04X 內' % (addr, seg)); return
    sig = bytes([0x9A, off & 0xFF, off >> 8, seg & 0xFF, seg >> 8])
    print('找 call far %04X:%04X ── 簽章 %s' % (seg, off, sig.hex().upper()))
    pos, n = 0, 0
    while True:
        pos = mem.find(sig, pos)
        if pos < 0:
            break
        print('  %05X' % pos); n += 1; pos += 1
    print('  共 %d 處' % n)

def nearxref(mem, addr, seg):
    """同段 near call（E8 rel16）與「把位移當常數用」的地方。

    函式被 near call 或被存成回呼指標時，`call far` 的簽章找不到它，
    但位移這兩個位元組還是會出現在程式碼裡。"""
    off = addr - seg * 16
    print('near call → %04X:%04X' % (seg, off))
    n = 0
    for i in range(len(mem) - 3):
        if mem[i] == 0xE8:
            rel = int.from_bytes(mem[i+1:i+3], 'little', signed=True)
            if (i + 3 + rel) & 0xFFFF == off & 0xFFFF and abs(i - addr) < 0x10000:
                print('  %05X' % i); n += 1
    print('  共 %d 處' % n)
    lo, hi = off & 0xFF, off >> 8
    print('位移常數 %02X %02X 出現在：' % (lo, hi))
    pat, pos, m = bytes([lo, hi]), 0, 0
    while True:
        pos = mem.find(pat, pos)
        if pos < 0:
            break
        if 0x1000 <= pos < 0x40000:     # 只看程式碼區
            print('  %05X  前後: %s' % (pos, mem[pos-6:pos+6].hex().upper()))
            m += 1
        pos += 1
    print('  共 %d 處（只列程式碼區）' % m)

def farptr(mem, addr):
    """掃全部記憶體，找任何解讀成 far 指標後正好指向這個線性位址的四位元組。

    同一個線性位址可以用很多組 `段:位移` 表示，所以不能只試一個段——
    要反過來算：每四個位元組當 (位移, 段) 解一次，看線性值對不對。"""
    n = 0
    for i in range(len(mem) - 4):
        off = mem[i] | mem[i+1] << 8
        seg = mem[i+2] | mem[i+3] << 8
        if seg * 16 + off == addr:
            kind = 'call far' if i >= 1 and mem[i-1] == 0x9A else ''
            print('  %05X  %04X:%04X %s  前後: %s' % (i, seg, off, kind, mem[max(0,i-6):i+6].hex().upper()))
            n += 1
    print('  共 %d 處' % n)

def climb(mem, addr, depth=8):
    """從一個位址一路往上追呼叫鏈：函式起點 → 誰指向它 → 那個人的函式起點…

    只在「呼叫者唯一」時往上走；分岔就停下來把候選印出來，
    讓人決定走哪一條，不要自己挑一條當成答案。"""
    seen = set()
    for level in range(depth):
        e = entry(mem, addr)
        if e is None or e in seen:
            print('（追不上去了）'); return
        seen.add(e)
        cands = []
        for i in range(len(mem) - 4):
            off = mem[i] | mem[i+1] << 8
            seg = mem[i+2] | mem[i+3] << 8
            if seg * 16 + off == e:
                cands.append(i)
        print('%d. %05X 的函式起點 %05X ← %d 個指標 %s'
              % (level, addr, e, len(cands), ' '.join('%05X' % c for c in cands)))
        if len(cands) != 1:
            return
        addr = cands[0]

cmd, path = sys.argv[1], sys.argv[2]
mem = load(path)
if cmd == 'hex':
    hexdump(mem, int(sys.argv[3], 16), int(sys.argv[4]))
elif cmd == 'entry':
    a = int(sys.argv[3], 16)
    e = entry(mem, a)
    print('函式起點 %05X（距離 %d bytes）' % (e, a - e) if e else '找不到')
elif cmd == 'xref':
    xref(mem, int(sys.argv[3], 16), int(sys.argv[4], 16))
elif cmd == 'climb':
    climb(mem, int(sys.argv[3], 16))
elif cmd == 'farptr':
    farptr(mem, int(sys.argv[3], 16))
elif cmd == 'nearxref':
    nearxref(mem, int(sys.argv[3], 16), int(sys.argv[4], 16))
