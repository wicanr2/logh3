#!/usr/bin/env python3
"""`.MDS` 樂曲檔 → JSON ＋ 標準 MIDI。

版面（實測，見 `docs/internals/17-mds.md`）：

    0x00  "M1" ＋ 曲名（ASCII，空白補齊）
    0x40  用到的 MIDI 通道，1 起算，0xFF 結束
    0x100 起 事件流，**每個事件固定 4 bytes**

事件：

    00–7F  音符   [音高, 步進, 音長, 力度]
    E0     音色   [步進, 音色編號, 0]
    E2     控制器 [步進, 控制器編號, 值]
    E4     未知   [步進, 0, 64]
    FA     速度   [步進, 速度, 0]
    FB     未解   [n, 0, 0]，與 FC 成對出現，可能是反覆記號
    FC     未解   [n, m, 0]，一軌的第一個事件時 n ＝ MIDI 通道
    FD     小節線 [幾小節, 0, 0]
    FE FE FE FE 分軌（唯一可靠的軌界；FC／FB 也會出現在軌中間）

一拍 48 tick。小節長度沒有寫在檔頭裡，用「兩條小節線之間的步進總和」
取眾數得到——Boléro 量出 144（3/4）、其餘多為 96（2/4）或 192（4/4）。
"""
import collections, glob, json, os, struct, sys

TPQ = 48
SRC = sys.argv[1] if len(sys.argv) > 1 else '/orig/*.MDS'
OUT_JSON = '/work/docs/internals/data/mds'
OUT_MID = '/work/workplace/assets/music'


def parse(d):
    title = d[2:0x20].decode('ascii', 'replace').rstrip()
    chans = []
    for b in d[0x40:0x50]:
        if b == 0xFF:
            break
        chans.append(b)
    # FE FE FE FE 分軌。FC／FB 不能當軌界——它們也出現在軌中間。
    segs, cur, p = [], [], 0x100
    while p + 4 <= len(d):
        b, a = d[p], list(d[p + 1:p + 4])
        if b == 0xFE:
            segs.append(cur)
            cur = []
        else:
            cur.append((b, a))
        p += 4
    segs.append(cur)

    tracks = []
    unknown = collections.Counter()
    for i, seg in enumerate(segs):
        if not seg:
            continue
        # 一軌的第一個事件是 FC 時，它的第一個引數就是 MIDI 通道。
        ch = seg[0][1][0] if seg[0][0] == 0xFC else (chans[i] if i < len(chans) else 1)
        cur = {'channel': ch, 'events': []}
        for b, a in seg:
            if b < 0x80:
                cur['events'].append(('note', b, a[0], a[1], a[2]))
            elif b == 0xFD:
                cur['events'].append(('bar', a[0]))
            elif b == 0xFA:
                cur['events'].append(('tempo', a[0], a[1]))
            elif b == 0xE0:
                cur['events'].append(('program', a[0], a[1]))
            elif b == 0xE2:
                cur['events'].append(('cc', a[0], a[1], a[2]))
            else:
                unknown[b] += 1
                cur['events'].append(('raw', b, a[0], a[1], a[2]))
        if any(e[0] == 'note' for e in cur['events']):
            tracks.append(cur)
    return title, chans, tracks, unknown


def bar_length(tracks):
    sums = collections.Counter()
    for t in tracks:
        acc = 0
        for e in t['events']:
            if e[0] == 'note':
                acc += e[2]
            elif e[0] in ('tempo', 'program', 'raw'):
                acc += e[1]
            elif e[0] == 'cc':
                acc += e[1]
            elif e[0] == 'bar':
                if acc:
                    sums[acc] += 1
                acc = 0
    return sums.most_common(1)[0][0] if sums else 192


def timeline(t, barlen, sane_bars=True):
    """把一軌展成（絕對 tick, 事件）。小節線是時間錨點：
    絕對時間 ＝ 小節序號 × 小節長度 ＋ 小節內的 tick。"""
    out, bar, tick = [], 0, 0
    for e in t['events']:
        at = bar * barlen + tick
        if e[0] == 'note':
            if e[4] and e[3]:
                out.append((at, 'note', e[1], e[3], e[4]))
            tick += e[2]
        elif e[0] == 'bar':
            # 引數幾乎都是 1。END_2F／END_2M 的引數是垃圾（249、115…），
            # 但每一軌的步進總和都是 135 × 192，證明小節線本身沒問題，
            # 一律前進一小節就對了。
            bar += e[1] if sane_bars else 1
            tick = 0
        elif e[0] == 'raw' and e[1] in (0xFB, 0xFC):
            # FC／FB 不推進時間：把它們的第一個引數當步進的話，
            # 「小節內步進總和 ＝ 小節長度」的比率從 90% 掉到 86%。
            out.append((at, 'raw', e[1], e[2], e[3], e[4]))
        else:
            out.append((at, ) + tuple(e[:1]) + tuple(e[2:]))
            tick += e[1]
    return out


def vlq(n):
    b = [n & 0x7F]
    n >>= 7
    while n:
        b.append((n & 0x7F) | 0x80)
        n >>= 7
    return bytes(reversed(b))


def midi(title, tracks, barlen, sane_bars=True):
    chunks = []
    for t in tracks:
        ch = max(0, min(15, t['channel'] - 1))
        ev = []
        for it in timeline(t, barlen, sane_bars):
            at, kind = it[0], it[1]
            if kind == 'note':
                _, _, note, gate, vel = it
                ev.append((at, 0, bytes([0x90 | ch, note & 0x7F, vel & 0x7F])))
                ev.append((at + gate, -1, bytes([0x80 | ch, note & 0x7F, 0])))
            elif kind == 'program':
                ev.append((at, 0, bytes([0xC0 | ch, it[2] & 0x7F])))
            elif kind == 'cc':
                ev.append((at, 0, bytes([0xB0 | ch, it[2] & 0x7F, it[3] & 0x7F])))
            elif kind == 'tempo':
                us = int(60000000 / max(1, it[2]))
                ev.append((at, -2, b'\xff\x51\x03' + struct.pack('>I', us)[1:]))
        ev.sort(key=lambda x: (x[0], x[1]))
        data, last = bytearray(), 0
        for at, _, msg in ev:
            data += vlq(max(0, at - last)) + msg
            last = max(last, at)
        data += b'\x00\xff\x2f\x00'
        chunks.append(b'MTrk' + struct.pack('>I', len(data)) + bytes(data))
    head = b'MThd' + struct.pack('>IHHH', 6, 1, len(chunks), TPQ)
    return head + b''.join(chunks)


os.makedirs(OUT_JSON, exist_ok=True)
os.makedirs(OUT_MID, exist_ok=True)
index = []
for f in sorted(glob.glob(SRC)):
    d = open(f, 'rb').read()
    name = os.path.splitext(os.path.basename(f))[0]
    title, chans, tracks, unknown = parse(d)
    barlen = bar_length(tracks)
    notes = sum(1 for t in tracks for e in t['events'] if e[0] == 'note')
    # 小節線的引數在正常檔案裡幾乎都是 1；END_2F／END_2M 的是亂數般的值，
    # 但那兩份每一軌的步進總和都是 135 × 192，小節線本身沒問題——
    # 引數不可信的時候一律前進一小節。
    barargs = [e[1] for t in tracks for e in t['events'] if e[0] == 'bar']
    sane = bool(barargs) and sum(1 for v in barargs if v <= 2) / len(barargs) > 0.9
    open(f'{OUT_MID}/{name}.mid', 'wb').write(midi(title, tracks, barlen, sane))
    json.dump({'file': os.path.basename(f), 'title': title, 'channels': chans,
               'ticks_per_quarter': TPQ, 'bar_ticks': barlen,
               'tracks': [{'channel': t['channel'],
                           'events': [list(e) for e in t['events']]} for t in tracks]},
              open(f'{OUT_JSON}/{name}.json', 'w', encoding='utf-8'),
              ensure_ascii=False)
    index.append({'file': os.path.basename(f), 'title': title, 'tracks': len(tracks),
                  'channels': chans, 'bar_ticks': barlen, 'notes': notes,
                  'bar_args_sane': sane,
                  'unknown_cmds': {hex(k): v for k, v in unknown.items()}})
    print('%-12s %-34s 軌 %2d  小節 %3d tick  音符 %5d  %s'
          % (name, title, len(tracks), barlen, notes,
             'MIDI 已輸出' if sane else 'MIDI 已輸出（小節線引數是垃圾，一律前進一小節）'))

json.dump({'note': '.MDS 樂曲索引；MIDI 由 tools/scripts/mds_export.py 從自備原版重建，保存版在 assets/original/music/mid/',
           'ticks_per_quarter': TPQ, 'songs': index},
          open(f'{OUT_JSON}/index.json', 'w', encoding='utf-8'),
          ensure_ascii=False, indent=1)
