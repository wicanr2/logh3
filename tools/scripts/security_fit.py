"""找「治安」是行星記錄的哪一格算出來的。

實測值（收據 sec1／sec5／sec7／sec9／hovA／main1e）：星域 → 治安 %。
"""
import pathlib, statistics
f = pathlib.Path('/orig/PLADAT.DAT').read_bytes()
HDR, REC = 64, 82
per = [f[i*2] | f[i*2+1] << 8 for i in range(32)]
start = [sum(per[:k]) for k in range(32)]
want = {0: 82, 1: 81, 2: 86, 5: 84, 7: 78, 9: 75}

def rec(i):
    o = HDR + i*REC
    return f[o:o+REC]

def planets(k):
    return [rec(i) for i in range(start[k], start[k] + per[k])]

print('每個星域的行星數：', {k: per[k] for k in want})
for off in range(REC):
    ok_avg = all(round(statistics.mean(p[off] for p in planets(k))) == v for k, v in want.items())
    ok_sum = all(sum(p[off] for p in planets(k)) == v for k, v in want.items())
    ok_first = all(planets(k)[0][off] == v for k, v in want.items())
    if ok_avg or ok_sum or ok_first:
        print('位移 0x%02X 命中：avg=%s sum=%s first=%s' % (off, ok_avg, ok_sum, ok_first))
# 沒有直接命中就把幾個小數值欄位印出來給人看
print()
for k, v in sorted(want.items()):
    ps = planets(k)
    print('星域%2d 治安%d%%  行星%d  +0x10..0x13=%s  +0x18=%s +0x1E=%s +0x14w=%s' % (
        k, v, per[k],
        [[p[o] for p in ps] for o in (0x10, 0x11, 0x12, 0x13)],
        [p[0x18] for p in ps], [p[0x1E] for p in ps],
        [p[0x14] | p[0x15] << 8 for p in ps]))
