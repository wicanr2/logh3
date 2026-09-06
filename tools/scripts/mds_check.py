# 兩個問題：(1) FC／FB 的第一個引數算不算步進？(2) 產出的 MIDI 結構對不對？
import collections, glob, os, struct
def bars(d, count_fcfb):
    segs, cur, p = [], [], 0x100
    while p + 4 <= len(d):
        b, a = d[p], list(d[p+1:p+4]); p += 4
        if b == 0xFE:
            segs.append(cur); cur = []
        else:
            cur.append((b, a))
    segs.append(cur)
    out = collections.Counter()
    for seg in segs:
        acc = 0
        for b, a in seg:
            if b < 0x80 or b in (0xE0, 0xE2, 0xE4, 0xFA):
                acc += a[0]
            elif b in (0xFB, 0xFC):
                if count_fcfb:
                    acc += a[0]
            elif b == 0xFD:
                if acc:
                    out[acc] += 1
                acc = 0
    return out

tot = {True: 0, False: 0}; hit = {True: 0, False: 0}
for f in sorted(glob.glob('/orig/*.MDS')):
    d = open(f, 'rb').read()
    for k in (True, False):
        c = bars(d, k)
        if not c:
            continue
        m = c.most_common(1)[0][0]
        tot[k] += sum(c.values()); hit[k] += c[m]
for k in (True, False):
    print('FC/FB %s步進：整齊的小節 %d / %d ＝ %.2f%%'
          % ('算' if k else '不算', hit[k], tot[k], 100*hit[k]/tot[k]))

print('=== MIDI 檢查 ===')
for f in sorted(glob.glob('/work/workplace/assets/music/*.mid'))[:24]:
    d = open(f, 'rb').read()
    assert d[:4] == b'MThd'
    _, fmt, ntrk, div = struct.unpack('>IHHH', d[4:14])
    p, longest, notes = 14, 0, 0
    for _ in range(ntrk):
        assert d[p:p+4] == b'MTrk', f
        ln = struct.unpack('>I', d[p+4:p+8])[0]
        blk = d[p+8:p+8+ln]; p += 8 + ln
        q, t = 0, 0
        while q < len(blk):
            v = 0
            while blk[q] & 0x80:
                v = (v << 7) | (blk[q] & 0x7F); q += 1
            v = (v << 7) | blk[q]; q += 1
            t += v
            st = blk[q]
            if st == 0xFF:
                q += 2; l = blk[q]; q += 1 + l
            elif st & 0xF0 in (0x90, 0x80, 0xB0):
                if st & 0xF0 == 0x90: notes += 1
                q += 3
            elif st & 0xF0 == 0xC0:
                q += 2
            else:
                raise SystemExit('未知狀態 %02X 於 %s' % (st, f))
        longest = max(longest, t)
    print('%-12s 軌 %d  最長 %6d tick ＝ %5.1f 拍  音符 %d'
          % (os.path.basename(f), ntrk, longest, longest/div, notes))
