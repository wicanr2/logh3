"""把 `introfull` 收據解成開頭動畫的逐格時程表。

    tools/py.sh tools/scripts/intro_schedule.py [收據名] [每刻幾道指令]

收據要用 `-regs-at 041F:024B,041A:0035` 跑（解碼器進入點 ＋ 讀檔常式），
兩邊合起來才回答得了「這一次解的是檔案裡第幾格」：

- `041A:0035` 的堆疊（`[bp+8]` 位移、`[bp+0A]` 段、`[bp+0C]` 長度）說出
  哪一塊資料被讀進哪一個緩衝區；`IBM16.350` 是順序讀的，所以把長度
  累加就知道每一塊對到檔案的哪一段。
- `041F:024B` 的 `DS:SI` 是這一格在緩衝區裡的起點（`BX` 是格內的段位移，
  `DX` 是 plane 的目的地偏移）。`DS` 指出是哪一個緩衝區。

兩者相減就是檔案位移，再拿 `IBMFLS16.350` 的累加長度換成格號。
輸出 JSON（給後續比對用）＋ 一張人看的表。
"""
import json
import re
import sys

REPORT = '/work/workplace/oracle/%s/report.txt' % (sys.argv[1] if len(sys.argv) > 1 else 'introfull')
TICK = int(sys.argv[2]) if len(sys.argv) > 2 else 20000
LENS = '/orig/IBMFLS16.350'

RE_HEAD = re.compile(r'^(\w{4}):(\w{4}) 執行時的暫存器')
RE_HIT = re.compile(r'^  #(\d+) .*BX=(\w{4}) CX=\w{4} DX=(\w{4}) SI=(\w{4}) .*DS=(\w{4}) ES=(\w{4})')
RE_BP = re.compile(r'^      \[bp\+\]((?: \w\w:\w{4})+)')


def parse_report(path):
    """回傳 {站點: [(步數, 暫存器dict, [bp+] 的 12 個 word)]}"""
    out, site, pending = {}, None, None
    for line in open(path, encoding='utf-8'):
        line = line.rstrip('\n')
        m = RE_HEAD.match(line)
        if m:
            site = '%s:%s' % (m.group(1), m.group(2))
            out.setdefault(site, [])
            pending = None
            continue
        if site is None:
            continue
        m = RE_HIT.match(line)
        if m:
            pending = [int(m.group(1)), {
                'BX': int(m.group(2), 16), 'DX': int(m.group(3), 16),
                'SI': int(m.group(4), 16), 'DS': int(m.group(5), 16),
                'ES': int(m.group(6), 16)}, []]
            out[site].append(pending)
            continue
        m = RE_BP.match(line)
        if m and pending is not None:
            pending[2] = [int(v.split(':')[1], 16) for v in m.group(1).split()]
            pending = None
            continue
        if line and not line.startswith('  '):
            site = None
    return out


def frame_bounds():
    fl = open(LENS, 'rb').read()
    lens = [fl[2 * i] | fl[2 * i + 1] << 8 for i in range(len(fl) // 2)]
    acc, out = 0, [0]
    for n in lens:
        acc += n
        out.append(acc)
    return out, lens


def main():
    hits = parse_report(REPORT)
    reads, decodes = hits.get('041A:0035', []), hits.get('041F:024B', [])

    # 讀檔：handle 5、目的地是兩個 64 KB 緩衝區之一、長度不為 0。
    #
    # **不要用「累加長度」推檔案位移。** 動畫跑完之後載入器會回頭重讀
    # 最後幾塊（結尾在迴圈播），累加會一路偏掉。分塊表的 30 個長度彼此
    # 互異，所以用長度反查塊號是一對一的，而且可以自己驗。
    tab = json.load(open('/work/docs/internals/data/intro-chunks.json'))['分塊']
    by_len = {}
    for c in tab:
        by_len.setdefault(c['length'], []).append(c['index'])
    dup = {k: v for k, v in by_len.items() if len(v) > 1}
    if dup:
        raise SystemExit('分塊長度不唯一，這支腳本的反查不成立：%r' % dup)
    chunks = []
    for step, r, bp in reads:
        if len(bp) < 7:
            continue
        handle, off, seg, n = bp[3], bp[4], bp[5], bp[6]
        if handle != 5 or seg not in (0x175C, 0x275D) or n == 0:
            continue
        idx = by_len.get(n, [None])[0]
        chunks.append({'step': step, 'seg': seg, 'off': off, 'n': n,
                       'chunk': idx, 'file': tab[idx]['offset'] if idx is not None else None})
    pos = sum(c['n'] for c in chunks)

    bounds, lens = frame_bounds()
    b2f = {v: i for i, v in enumerate(bounds)}

    # 解碼：連續 4 次同一個 (DS, SI) ＝ 一格的四個 plane。
    groups, cur = [], None
    for step, r, _ in decodes:
        key = (r['DS'], r['SI'])
        if cur is None or cur['key'] != key:
            cur = {'key': key, 'step': step, 'calls': []}
            groups.append(cur)
        cur['calls'].append((step, r['BX'], r['DX'], r['ES']))

    rows = []
    for g in groups:
        ds, si = g['key']
        ch = None
        for c in chunks:
            if c['seg'] == ds and c['step'] < g['step']:
                ch = c
        if ch is None:
            rows.append({'step': g['step'], 'ds': ds, 'si': si, 'calls': len(g['calls']),
                         'read': None, 'chunk': None, 'file': None, 'frame': None})
            continue
        fo = ch['file'] + (si - ch['off'])
        rows.append({'step': g['step'], 'ds': ds, 'si': si, 'calls': len(g['calls']),
                     'read': chunks.index(ch), 'chunk': ch['chunk'],
                     'file': fo, 'frame': b2f.get(fo)})

    for i, row in enumerate(rows):
        row['dstep'] = row['step'] - rows[i - 1]['step'] if i else 0
        row['tick'] = round(row['dstep'] / TICK, 1)
        row['len'] = lens[row['frame']] if row['frame'] is not None and row['frame'] < len(lens) else None

    out = {'report': REPORT, 'tick': TICK, 'chunks': chunks, 'frames': rows}
    json.dump(out, open('/work/workplace/intro-schedule.json', 'w'), ensure_ascii=False, indent=1)

    print('讀檔 %d 次，累計 %d bytes（檔案本身 1,764,911）' % (len(chunks), pos))
    for i, c in enumerate(chunks):
        print('  第%-3d次 #%-11d %04X:%04X  %6d bytes  ＝ 塊%-2s 檔案[%d:%d]'
              % (i, c['step'], c['seg'], c['off'], c['n'], c['chunk'],
                 c['file'], c['file'] + c['n']))
    print()
    ok = sum(1 for r in rows if r['frame'] is not None)
    print('解碼 %d 組（%d 次呼叫），對得上格邊界的 %d 組'
          % (len(rows), sum(r['calls'] for r in rows), ok))
    seq = [r['frame'] for r in rows if r['frame'] is not None]
    print('格號序列前 20：', seq[:20])
    print('格號序列後 20：', seq[-20:])
    print('是否嚴格遞增 1：', all(b - a == 1 for a, b in zip(seq, seq[1:])))
    bad = [r for r in rows if r['frame'] is None]
    print('對不上的組：%d' % len(bad))
    for r in bad[:10]:
        print('   #%d DS=%04X SI=%04X 呼叫%d 次 chunk=%s file=%s'
              % (r['step'], r['ds'], r['si'], r['calls'], r['chunk'], r['file']))
    print()
    print('每組呼叫次數不是 4 的：',
          [(r['step'], r['calls']) for r in rows if r['calls'] != 4][:10])
    print()
    print(' 格  步數          Δ步       刻    壓縮長度  塊')
    for r in rows:
        if r['frame'] is None:
            continue
        print('%4s %-13d %-9d %7.1f %7s %5s'
              % (r['frame'], r['step'], r['dstep'], r['tick'],
                 r['len'], r['chunk']))


main()
