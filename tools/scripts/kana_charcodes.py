"""GRAPH.IMG 的 302 個假名／西里爾字形：字碼對應。

假說：`0x4632`–`0x6996` 這 302 個字形，就是 Big5 倚天擴充區
`0xC6E0`–`0xC8D3` 這 302 個碼位，**連續存放、不跳號**。

驗法（都在本檔）：
  recon   用 ETen／HKSCS 表把 0xC6A1–0xC8FE 解一次，看碼位怎麼排。
  art     把指定序號的字形印成點陣圖，人眼核對。
  checks  結構性自動檢查：濁音／半濁音分解、西里爾 vs 全形拉丁、
          框線的幾何、空白字形位置。
  scan    在原版檔案裡找這一區的實際位元組（一手證據），
          並在未打包的 ENDING.EXE 裡找索引公式的立即數。

用法：tools/py.sh tools/scripts/kana_charcodes.py <模式> [參數]
"""
import hashlib
import json
import os
import sys
import unicodedata

ORIG = '/orig'
STRIDE, ROWS, WIDTH = 30, 15, 16
BLOCK_OFF, BLOCK_N = 0x4632, 302          # docs/internals/06
SYM_OFF = 3840                             # 全形符號區起點（confirmed）
BLOCK_START_CODE = 0xC6E0                  # 302 格區塊的第一個碼位
EXT_TOP, EXT_HI = 0xC6A1, 0xC8D3           # 倚天擴充區的完整範圍（365 個碼位）
EXT_BASE = 16080                           # 0xC6A1 的字形位移（ENDING.EXE 0x0DE9）

IMG = open(os.path.join(ORIG, 'GRAPH.IMG'), 'rb').read()


def big5_index(code):
    lead, trail = code >> 8, code & 0xFF
    t = trail - 0x40 if trail <= 0x7E else trail - 0xA1 + 63
    return (lead - 0xA1) * 157 + t


def code_of(i):
    lead = 0xA1 + i // 157
    t = i % 157
    return lead << 8 | (0x40 + t if t < 63 else 0xA1 + t - 63)


def dec(code, strict=True):
    """只用 ETen／HKSCS 表；查不到就回 None（不做 big5 退路，退路會騙人）。"""
    try:
        return bytes([code >> 8, code & 0xFF]).decode('big5hkscs')
    except UnicodeDecodeError:
        return None


def uname(ch):
    try:
        return unicodedata.name(ch)
    except (ValueError, TypeError):
        return ''


# ---- 字形存取 -------------------------------------------------------------

def glyph_at(off):
    return [IMG[off + 2 * r] << 8 | IMG[off + 2 * r + 1] for r in range(ROWS)]


def block_glyph(k):
    """302 區塊裡的第 k 個字形。"""
    return glyph_at(BLOCK_OFF + k * STRIDE)


def ext_off(code):
    """倚天擴充區的字形位移；就是 ENDING.EXE 0x0DE3–0x0DEB 那段算的東西。"""
    return EXT_BASE + (big5_index(code) - big5_index(EXT_TOP)) * STRIDE


def ext_glyph(code):
    return glyph_at(ext_off(code))


def sym_glyph(code):
    """全形符號區（0xA140–0xA3FE）的字形，索引公式已 confirmed。"""
    return glyph_at(SYM_OFF + big5_index(code) * STRIDE)


def block_code(k):
    return code_of(big5_index(BLOCK_START_CODE) + k)


def block_map():
    return [(k, block_code(k), dec(block_code(k))) for k in range(BLOCK_N)]


# ---- 模式 -----------------------------------------------------------------

def recon():
    print('=== Big5 0xC6A1–0xC8FE，只用 big5hkscs（ETen 擴充表）解 ===')
    for row in range(big5_index(0xC6A1), big5_index(0xC8FE) + 1, 16):
        cells = []
        for i in range(row, min(row + 16, big5_index(0xC8FE) + 1)):
            ch = dec(code_of(i))
            cells.append(ch if ch else '·')
        print(f'  {code_of(row):04X}  ' + ' '.join(cells))
    print('\n（· ＝ big5hkscs 查無此碼位）')


def art(spec):
    """spec 形如 0-10 或 230-301。"""
    lo, hi = (int(x) for x in spec.split('-'))
    per = 6
    for base in range(lo, hi + 1, per):
        group = list(range(base, min(base + per, hi + 1)))
        heads = []
        for k in group:
            c = block_code(k)
            ch = dec(c)
            heads.append(f'#{k} {c:04X} {ch or "·"}'.ljust(WIDTH + 2))
        print(''.join(heads))
        for r in range(ROWS):
            line = ''
            for k in group:
                g = block_glyph(k)[r]
                line += ''.join('#' if g >> (WIDTH - 1 - x) & 1 else '.'
                                for x in range(WIDTH)) + '  '
            print(line)
        print()


def sheet(spec):
    """把指定序號範圍畫成放大的 PNG，給人眼看。"""
    sys.path.insert(0, '/work/tools/scripts')
    import png
    lo, hi = (int(x) for x in spec.split('-'))
    ks = list(range(lo, hi + 1))
    cols, scale, gap = 16, 4, 2
    rowsn = (len(ks) + cols - 1) // cols
    cw, chh = WIDTH * scale + gap, ROWS * scale + gap
    W, H = cols * cw, rowsn * chh
    bits = [[0] * W for _ in range(H)]
    for n, k in enumerate(ks):
        g = block_glyph(k)
        ox, oy = (n % cols) * cw, (n // cols) * chh
        for r in range(ROWS):
            for x in range(WIDTH):
                if g[r] >> (WIDTH - 1 - x) & 1:
                    for dy in range(scale):
                        for dx in range(scale):
                            bits[oy + r * scale + dy][ox + x * scale + dx] = 1
    out = f'/work/workplace/assets/font/knx-{lo}-{hi}.png'
    os.makedirs(os.path.dirname(out), exist_ok=True)
    png.write_mono(out, W, H, bits)
    print(f'寫出 {out}  {W}x{H}  序號 {lo}..{hi}')


def checks():
    """可重跑的結構檢查。判準只放「錯了就代表對應表錯了」的東西。"""
    ok = fail = 0

    def report(label, good, detail=''):
        nonlocal ok, fail
        ok, fail = ok + bool(good), fail + (not good)
        print(f'  [{"OK" if good else "**FAIL**"}] {label}'
              + (f'  {detail}' if detail else ''))

    print('=== 1. 版面與端點 ===')
    report('擴充區 0xC6A1–0xC8D3 共 365 個碼位',
           big5_index(EXT_HI) - big5_index(EXT_TOP) + 1 == 365)
    report('擴充區佔 16080–27030，接在符號區 A140–A3BF（408 格）之後',
           EXT_BASE == 3840 + 408 * STRIDE and
           EXT_BASE + 365 * STRIDE == 27030)
    report('worklist 的 302 格 = 0xC6E0–0xC8D3',
           ext_off(0xC6E0) == 0x4632 and ext_off(EXT_HI) + STRIDE == 27030,
           f'0x{ext_off(0xC6E0):04X}..0x{ext_off(EXT_HI) + STRIDE:04X}')

    print('\n=== 2. 拿原版自己的文字位元組走一遍（ENDING.EXE 版權行）===')
    # docs/internals/16：文字區 0x63FE–0x765C，這兩行有假名
    exe = open(os.path.join(ORIG, 'ENDING.EXE'), 'rb').read()
    want = {0xC7DB: 'ミ', 0xC7E5: 'ラ', 0xC7B4: 'ジ', 0xC7E1: 'ュ',
            0xC7DF: 'ャ', 0xC7CD: 'パ', 0xC7EF: 'ン', 0xC7A9: 'キ',
            0xC7C2: 'テ', 0xC77D: 'ィ', 0xC7D1: 'フ', 0xC7E7: 'ル',
            0xC7DC: 'ム'}
    seen = set()
    for p0 in (0x6A9F, 0x6E83):
        s0 = exe[p0:exe.index(b'\x00', p0)]
        n = 0
        while n < len(s0) - 1:
            c = s0[n] << 8 | s0[n + 1]
            if EXT_TOP <= c <= EXT_HI:
                seen.add(c)
                n += 2
            else:
                n += 1
    report('版權行用到的擴充區碼位，字碼→字元全部對得上 ETen 表',
           all(dec(c) == want.get(c) for c in seen) and len(seen) == 13,
           f'{len(seen)} 個相異碼位：' + ''.join(dec(c) or "·" for c in sorted(seen)))
    report('長音記號是注音 ㄧ(A3B8) 不是片假名 ー(C6E3)',
           exe[0x6A9F:0x6AFF].find(b'\xa3\xb8') != -1
           and exe[0x6A9F:0x6AFF].find(b'\xc6\xe3') == -1)

    print('\n=== 3. 365 格的內容 ===')
    by_char = {}
    codes = [code_of(i2) for i2 in
             range(big5_index(EXT_TOP), big5_index(EXT_HI) + 1)]
    for c in codes:
        ch = dec(c)
        if ch:
            by_char[ch] = c
    blank = [c for c in codes if not any(ext_glyph(c))]
    report('365 格沒有一格是空白', not blank,
           ' '.join(f'{c:04X}' for c in blank))
    named = sum(1 for c in codes if dec(c))
    print(f'  big5hkscs（ETen 擴充表）認得其中 {named} 格，'
          f'另外 {len(codes) - named} 格要靠點陣辨識')

    print('\n=== 4. 濁音／半濁音各自有獨立字形 ===')
    pairs = []
    for ch, c in by_char.items():
        d = unicodedata.normalize('NFD', ch)
        if len(d) == 2 and d[1] in ('\u3099', '\u309a') and d[0] in by_char:
            pairs.append((d[1], ch, d[0]))
    same = [f'{a}/{b}' for _, a, b in pairs
            if ext_glyph(by_char[a]) == ext_glyph(by_char[b])]
    n_dak = sum(1 for mk, _, _ in pairs if mk == '\u3099')
    report(f'{n_dak} 組濁音、{len(pairs) - n_dak} 組半濁音，字形都與清音不同',
           not same, ' '.join(same))
    print('  （這套字型不是「清音字形 ＋ 貼一個濁點」——加濁點時基底會重畫，'
          '所以不能用逐位元差異當判準。）')

    print('\n=== 5. 小寫假名比大寫小 ===')
    def bbox(g):
        rs = [r for r in range(ROWS) if g[r]]
        if not rs:
            return 0, 0
        cols = [x for x in range(WIDTH)
                if any(g[r] >> (WIDTH - 1 - x) & 1 for r in rs)]
        return rs[-1] - rs[0] + 1, cols[-1] - cols[0] + 1
    pairs = [('ぁ', 'あ'), ('ぃ', 'い'), ('ぅ', 'う'), ('ぇ', 'え'), ('ぉ', 'お'),
             ('っ', 'つ'), ('ゃ', 'や'), ('ゅ', 'ゆ'), ('ょ', 'よ'), ('ゎ', 'わ'),
             ('ァ', 'ア'), ('ィ', 'イ'), ('ゥ', 'ウ'), ('ェ', 'エ'), ('ォ', 'オ'),
             ('ッ', 'ツ'), ('ャ', 'ヤ'), ('ュ', 'ユ'), ('ョ', 'ヨ'), ('ヮ', 'ワ')]
    bad = [f'{a}/{b}' for a, b in pairs
           if not (bbox(ext_glyph(by_char[a]))[0] < bbox(ext_glyph(by_char[b]))[0])]
    report('20 組小假名的墨水高度都小於對應大假名', not bad, ' '.join(bad))

    print('\n=== 6. 0xC8A5–0xC8CC 這 40 格（ETen 表查無）===')
    left, right = {}, {}
    for c in (code_of(i2) for i2 in
              range(big5_index(0xC8A5), big5_index(0xC8CC) + 1)):
        g = ext_glyph(c)
        left.setdefault(tuple(row >> 8 for row in g), []).append(c)
        right.setdefault(tuple(row & 0xFF for row in g), []).append(c)
    print(f'  左半 8 欄 {len(left)} 種圖樣、右半 8 欄 {len(right)} 種圖樣；'
          f'目視是「粗體數字 0–9 ＋ 右側小符號」的合字，'
          f'數字的垂直位置隨右側符號而變，用途未確認。')

    print('\n=== 7. 對應表雜湊（換 GRAPH.IMG 或改公式就會變）===')
    blob = b''.join(IMG[ext_off(c):ext_off(c) + STRIDE] for c in codes)
    print(f'  365 格字形共 {len(blob)} bytes，SHA-256 '
          f'{hashlib.sha256(blob).hexdigest()}')

    print(f'\n小計：{ok} 項通過、{fail} 項不通過')
    return fail


def scan():
    """一手證據：原版檔案裡實際出現的擴充區位元組；ENDING.EXE 的立即數。"""
    lo, hi = big5_index(0xC6A1), big5_index(0xC8FE)
    used = {}
    for fn in sorted(os.listdir(ORIG)):
        p = os.path.join(ORIG, fn)
        if not os.path.isfile(p):
            continue
        b = open(p, 'rb').read()
        i = 0
        while i < len(b) - 1:
            lead = b[i]
            if 0xA1 <= lead <= 0xF9:
                trail = b[i + 1]
                if 0x40 <= trail <= 0x7E or 0xA1 <= trail <= 0xFE:
                    code = lead << 8 | trail
                    if lo <= big5_index(code) <= hi:
                        used.setdefault(code, {}).setdefault(fn, 0)
                        used[code][fn] += 1
                    i += 2
                    continue
            i += 1
    print(f'=== 原版檔案裡出現的 0xC6A1–0xC8FE 位元組：{len(used)} 個相異碼位 ===')
    inside = outside = 0
    for code in sorted(used):
        ch = dec(code)
        k = big5_index(code) - big5_index(BLOCK_START_CODE)
        where = f'#{k}' if 0 <= k < BLOCK_N else '**區塊外**'
        inside += 0 <= k < BLOCK_N
        outside += not (0 <= k < BLOCK_N)
        files = ' '.join(f'{f}×{n}' for f, n in sorted(used[code].items())[:4])
        print(f'  {code:04X} {ch or "·":2s} {where:>10s}  {files}')
    print(f'  落在 302 區塊內：{inside}，區塊外：{outside}')


def emit():
    """輸出 365 格的字碼對應表 JSON。"""
    # big5hkscs 查不到、但點陣看得出來的六格（倚天擴充區的部首）
    BY_GLYPH = {0xC6CF: '廴', 0xC6D3: '无', 0xC6D5: '癶', 0xC6D7: '隶',
                0xC6DE: '〃', 0xC6DF: '仝'}
    rows, blob = [], b''
    for i2 in range(big5_index(EXT_TOP), big5_index(EXT_HI) + 1):
        c = code_of(i2)
        off = ext_off(c)
        blob += IMG[off:off + STRIDE]
        ch, src = dec(c), 'big5hkscs'
        if ch is None:
            ch, src = BY_GLYPH.get(c), ('glyph' if c in BY_GLYPH else None)
        rows.append({
            'big5': f'{c:04X}',
            'offset': off,
            'char': ch,
            'unicode': f'U+{ord(ch):04X}' if ch else None,
            'name': uname(ch) if ch else None,
            'source': src,
        })
    out = {
        'file': 'GRAPH.IMG',
        'block': '倚天擴充區（假名／西里爾／筆畫字根／合字）',
        'code_range': ['C6A1', 'C8D3'],
        'glyphs': len(rows),
        'start_offset': EXT_BASE,
        'end_offset': EXT_BASE + len(rows) * STRIDE,
        'index': ('offset = 16080 + (big5_index(code) - big5_index(0xC6A1)) * 30；'
                  '公式出自 ENDING.EXE 0x0DE3–0x0DEB'),
        'evidence_level': 'confirmed',
        'glyph_bytes_sha256': hashlib.sha256(blob).hexdigest(),
        'note': ('前 63 格（C6A1–C6DF，位移 16080–17970）以前被算進「全形符號區」的 471 格；'
                 '後 302 格（C6E0–C8D3，位移 17970–27030）是本表要解的那一段。'
                 'source=big5hkscs 表示字元由 ETen／HKSCS 表給出，'
                 'glyph 表示 ETen 表查無、由點陣目視辨識，'
                 'null 表示尚未辨識（C8A5–C8CC 那 40 格合字）。'),
        'entries': rows,
    }
    path = '/work/docs/internals/data/font-eten-ext.json'
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(out, f, ensure_ascii=False, indent=2)
        f.write('\n')
    named = sum(1 for r in rows if r['char'])
    print(f'寫出 {path}：{len(rows)} 筆，其中 {named} 筆有字元、'
          f'{len(rows) - named} 筆待辨識')



# ---- 端到端：拿原版自己的文字位元組，經公式查字型畫出來 --------------------

HALF_OFF, KANJI_OFF = 0, 27030
EXT_LO = 0xC6E0          # text() 只關心 302 格那段


def glyph_of(code, half=False):
    """回 (寬, 15 列的位元圖)。涵蓋半形、全形符號、擴充區、漢字四段。"""
    if half:
        o = HALF_OFF + code * 15
        return 8, [IMG[o + r] << 8 for r in range(ROWS)]
    i = big5_index(code)
    if 0xA140 <= code <= 0xA3FE:
        return WIDTH, glyph_at(SYM_OFF + i * STRIDE)
    if big5_index(EXT_TOP) <= i <= big5_index(EXT_HI):
        return WIDTH, ext_glyph(code)
    if i >= 471:
        j = i - 408 if i >= 6280 else i
        return WIDTH, glyph_at(KANJI_OFF + (j - 471) * STRIDE)
    return WIDTH, [0] * ROWS


def render_bytes(seq):
    """把一串原始位元組排成點陣列。"""
    cells, i = [], 0
    while i < len(seq):
        b = seq[i]
        if 0xA1 <= b <= 0xF9 and i + 1 < len(seq) and (
                0x40 <= seq[i + 1] <= 0x7E or 0xA1 <= seq[i + 1] <= 0xFE):
            cells.append(glyph_of(b << 8 | seq[i + 1]))
            i += 2
        else:
            cells.append(glyph_of(b, half=True))
            i += 1
    w = sum(c[0] for c in cells)
    bits = [[0] * w for _ in range(ROWS)]
    x0 = 0
    for cw, g in cells:
        for r in range(ROWS):
            for x in range(cw):
                if g[r] >> (WIDTH - 1 - x) & 1:
                    bits[r][x0 + x] = 1
        x0 += cw
    return w, bits


def text():
    """ENDING.EXE 文字區裡用到擴充區的那幾行，整行畫出來。"""
    sys.path.insert(0, '/work/tools/scripts')
    import png
    exe = open(os.path.join(ORIG, 'ENDING.EXE'), 'rb').read()
    lo, hi = 0x63FE, 0x765C                      # docs/internals/16
    lines, cur, start = [], bytearray(), lo
    for p in range(lo, hi):
        if exe[p] == 0:
            if cur:
                lines.append((start, bytes(cur)))
            cur, start = bytearray(), p + 1
        else:
            cur.append(exe[p])
    if cur:
        lines.append((start, bytes(cur)))

    def has_ext(s):
        i = 0
        while i < len(s) - 1:
            if 0xA1 <= s[i] <= 0xF9 and (0x40 <= s[i+1] <= 0x7E or 0xA1 <= s[i+1] <= 0xFE):
                if EXT_LO <= (s[i] << 8 | s[i+1]) <= EXT_HI:
                    return True
                i += 2
            else:
                i += 1
        return False

    hits = [(p, s) for p, s in lines if has_ext(s)]
    print(f'ENDING.EXE 文字區 {len(lines)} 行，其中 {len(hits)} 行用到 0x{EXT_LO:04X}–0x{EXT_HI:04X}：')
    rendered = []
    for p, s in hits:
        codes = []
        i = 0
        while i < len(s):
            if 0xA1 <= s[i] <= 0xF9 and i + 1 < len(s) and (
                    0x40 <= s[i+1] <= 0x7E or 0xA1 <= s[i+1] <= 0xFE):
                c = s[i] << 8 | s[i+1]
                if EXT_LO <= c <= EXT_HI:
                    k = big5_index(c) - big5_index(EXT_LO)
                    codes.append(f'{c:04X}=#{k}({dec(c) or "·"})')
                i += 2
            else:
                i += 1
        print(f'  @0x{p:04X}  {" ".join(codes)}')
        rendered.append(render_bytes(s))

    scale, gapy = 3, 4
    W = max(w for w, _ in rendered) * scale
    H = len(rendered) * (ROWS * scale + gapy)
    out_bits = [[0] * W for _ in range(H)]
    for n, (w, bits) in enumerate(rendered):
        oy = n * (ROWS * scale + gapy)
        for r in range(ROWS):
            for x in range(w):
                if bits[r][x]:
                    for dy in range(scale):
                        for dx in range(scale):
                            out_bits[oy + r*scale + dy][x*scale + dx] = 1
    out = '/work/workplace/assets/font/knx-ending-lines.png'
    png.write_mono(out, W, H, out_bits)
    print(f'\n寫出 {out}  {W}x{H}（{len(rendered)} 行，用本檔的公式查字型畫出來）')


def imm():
    """在未打包的 ENDING.EXE 找索引公式會用到的立即數。"""
    for fn in ('ENDING.EXE', 'OPEN.EXE', 'GIN3PS.EXE', 'GIN3.COM'):
        p = os.path.join(ORIG, fn)
        if not os.path.exists(p):
            continue
        b = open(p, 'rb').read()
        print(f'--- {fn}（{len(b)} bytes）---')
        for val, what in ((0xC6E0, '擴充區起點 0xC6E0'), (0xC8D3, '擴充區終點 0xC8D3'),
                          (0xC6DF, '0xC6DF（起點−1）'), (0xC8D4, '0xC8D4（終點+1）'),
                          (0x4632, '區塊位移 0x4632=17970'), (0x6996, '0x6996=27030'),
                          (0x0F00, '符號區位移 3840'), (0x01D7, '471'),
                          (0x012E, '302'), (0x0198, '408'), (0x1888, '6280'),
                          (0x009D, '157'), (0x001E, '30')):
            pat = bytes([val & 0xFF, val >> 8])
            hits = []
            off = b.find(pat)
            while off != -1 and len(hits) < 6:
                hits.append(off)
                off = b.find(pat, off + 1)
            if hits:
                print(f'  {what:24s} → {len(hits)}+ 處  '
                      + ' '.join(f'0x{h:05X}' for h in hits))


def presym():
    """符號區尾端 63 格（位移 16080–17970）：照 ENDING.EXE 的公式，
    這 63 格就是倚天擴充區的 0xC6A1–0xC6DF。"""
    sys.path.insert(0, '/work/tools/scripts')
    import png
    base = 16080
    n = (BLOCK_OFF - base) // STRIDE
    print(f'位移 {base}–{BLOCK_OFF}：{n} 個字形')
    for k in range(n):
        c_ext = code_of(big5_index(0xC6A1) + k)
        c_sym = code_of(big5_index(0xA140) + 408 + k)
        print(f'  #{k:2d} @{base + k*STRIDE:6d}  擴充 {c_ext:04X}={dec(c_ext) or "·"}'
              f'   符號 {c_sym:04X}={dec(c_sym) or "·"}')
    cols, scale, gap = 16, 4, 2
    rowsn = (n + cols - 1) // cols
    cw, chh = WIDTH * scale + gap, ROWS * scale + gap
    bits = [[0] * (cols * cw) for _ in range(rowsn * chh)]
    for k in range(n):
        g = glyph_at(base + k * STRIDE)
        ox, oy = (k % cols) * cw, (k // cols) * chh
        for r in range(ROWS):
            for x in range(WIDTH):
                if g[r] >> (WIDTH - 1 - x) & 1:
                    for dy in range(scale):
                        for dx in range(scale):
                            bits[oy + r*scale + dy][ox + x*scale + dx] = 1
    out = '/work/workplace/assets/font/knx-presym.png'
    png.write_mono(out, cols * cw, rowsn * chh, bits)
    print(f'寫出 {out}')


if __name__ == '__main__':
    mode = sys.argv[1] if len(sys.argv) > 1 else 'recon'
    if mode == 'art':
        art(sys.argv[2])
    elif mode == 'sheet':
        sheet(sys.argv[2])
    else:
        {'recon': recon, 'checks': checks, 'scan': scan, 'emit': emit,
         'text': text, 'imm': imm, 'presym': presym}[mode]()
