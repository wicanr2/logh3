"""把設定畫面的熱區清單從執行期傾印讀出來，寫進 `screen-settings.json`。

    tools/py.sh tools/scripts/settings_hotspots.py [傾印] [清單位移hex]

清單在資料段 `ds:03D8`，一筆 12 bytes（版面見 `docs/internals/23`）：

    +0 格x  +1 格y  +2 寬（格）  +3 高（格）
    +4 可選旗標  +5 動作碼  +6 元件種類  +7 ?
    +8 far 指標 → Big5 標籤，或三個位元組的動作碼（`FF 17 00` ＝ 加）

以 `FF FF` 結束，共 22 筆（`docs/re/285`）。

**這份清單 `hotspot_scan.py` 掃不到**（它是 `docs/re/156` 從命中測試的
呼叫點補回來的七份之一）：那支要求 `+5` 從 0 開始逐筆遞增，而這裡的
`+5` 是動作碼不是索引——最後一筆是 22，跳過 21，整段因此落選。

`overlay_spec.py` 重量設定畫面時會把 `hotspots` 清成空陣列，所以這一步要
在它之後再跑一次（`tools/specs.sh` 的 `settings` 那一段已經串好）。

**位址會自己驗**：換算出來的像素矩形要與規格裡量到的小框對得起來
（框左右各比熱區多 2 像素，`docs/re/68`）。對不上就不寫檔。
"""
import json
import sys

SRC = sys.argv[1] if len(sys.argv) > 1 else '/work/workplace/s3datamem.bin'
LIST = int(sys.argv[2], 16) if len(sys.argv) > 2 else 0x03D8
SPEC = '/work/docs/internals/data/screen-settings.json'

# GIN3PS 的 DGROUP 基底（`tools/scripts/read_ds_string.py` 用的同一個，
# `docs/re/274` §1 用立即值投票交叉驗過）。
DS = 0x54B30
CELL_W, CELL_H = 16, 20

# 動作碼 → remake 的 action 名。0–11 的語意由反組譯釘住（`docs/re/274` §5）；
# 12 起只有標籤，remake 還沒有對應的狀態，先留著不接。
ACTIONS = {
    0: ('device-0', '音源 → ds:8F33 ＝ 0'),
    1: ('device-1', '音源 → ds:8F33 ＝ 1'),
    2: ('device-2', '音源 → ds:8F33 ＝ 2，不播'),
    12: ('msg-speed-inc', '訊息速度 ＋（動作碼 FF 17 00，元件種類 6 ＝ 上下箭頭）'),
    13: ('msg-speed-dec', '訊息速度 −（動作碼 FF 18 00）'),
    14: ('anim-on', '動畫效果'),
    15: ('anim-off', '動畫效果'),
    16: ('anim-speed-0', '動畫速度'),
    17: ('anim-speed-1', '動畫速度'),
    18: ('anim-speed-2', '動畫速度'),
    19: ('close', '關掉設定畫面；回到哪一張還沒量'),
    20: ('bug', '畫面上沒有框、也沒有字的隱藏熱區'),
    22: ('midi-if', '七選一，選完寫回 MUSIC.INI +4（`docs/re/274` §10）'),
}
for n in range(3, 9):
    ACTIONS[n] = ('bgm-%d' % (n - 3), '平時配樂 %d → ds:8F34 ＝ %d' % (n - 3, n - 3))
for n in range(9, 12):
    ACTIONS[n] = ('battle-%d' % (n - 9), '戰鬥配樂 %d → ds:8F35 ＝ %d' % (n - 9, n - 9))

mem = open(SRC, 'rb').read()
# 基底自檢：ds:02DA 是音源那一列的第一個標籤。錯了就整份都錯，不要往下算。
if mem[DS + 0x02DA:DS + 0x02DE] != b'MIDI':
    sys.exit('%s 的 ds:02DA 不是 "MIDI"，DGROUP 基底 %05X 不對' % (SRC, DS))


def text_at(off):
    p = DS + off
    raw = mem[p:mem.find(b'\0', p)]
    if raw[:1] == b'\xff':          # 動作碼，不是字串
        return None
    return raw.decode('big5', 'replace').rstrip()


spec = json.load(open(SPEC))
win = spec['windows']['overlay']
# 原點格 ＝ (視窗x / 16 ＋ 1, 視窗y / 20 ＋ 1)（`docs/internals/23`）。
ox, oy = win['x'] // CELL_W + 1, win['y'] // CELL_H + 1
boxes = win['boxes']

out, used, bad = [], set(), []
p = LIST
while True:
    r = mem[DS + p:DS + p + 12]
    if r[:2] == b'\xff\xff':
        break
    gx, gy, gw, gh, enabled, act, kind = r[0], r[1], r[2], r[3], r[4], r[5], r[6]
    off = r[8] | r[9] << 8
    name, why = ACTIONS.get(act, ('action-%d' % act, ''))
    label = text_at(off)
    h = {'x': (gx + ox) * CELL_W, 'y': (gy + oy) * CELL_H,
         'w': gw * CELL_W, 'h': gh * CELL_H,
         'grid': [gx, gy], 'action': name}
    if kind:
        h['kind'] = kind
    # 找對應的小框：框左右各比熱區多 2 像素，上下對齊（`docs/re/68`）。
    # **上下箭頭（元件種類 6）是例外**：那兩個框與熱區一樣大、對齊。
    pad = 0 if kind == 6 else 2
    box = next((i for i, b in enumerate(boxes)
                if b['x'] == h['x'] - pad and b['y'] == h['y']
                and b['w'] == h['w'] + 2 * pad and b['h'] == h['h']), -1)
    if box >= 0:
        used.add(box)
        got = boxes[box].get('text', '')
        if label is not None and got and got != label:
            bad.append('ds:%04X 的標籤是 %r，框 %d 上印的是 %r' % (p, label, box, got))
    elif enabled:
        # 可選、卻沒有框：`BUG!` 那一格畫面上本來就什麼都沒有。
        if name != 'bug':
            bad.append('ds:%04X（%s）算出 (%d,%d) %dx%d，規格裡沒有這個框'
                       % (p, name, h['x'], h['y'], h['w'], h['h']))
    if not enabled:
        # 可選旗標 0 ＝ 原版的命中測試直接跳過這一筆。音源那三格的旗標
        # 是開機偵測的結果（`ds:03DC`／`03E8`／`03F4`），出貨的 `MUSIC.INI`
        # 沒有 MIDI 介面，所以「ＭＩＤＩ」點不下去、框也不畫。
        h['disabled'] = True
    note = label if label else ''
    if why:
        note = ('%s：%s' % (note, why)) if note else why
    h['note'] = '%s（ds:%04X）' % (note, p) if note else 'ds:%04X' % p
    out.append(h)
    p += 12

missing = [i for i in range(len(boxes)) if i not in used]
if missing:
    bad.append('規格裡有 %d 個框沒有熱區認領：%s' % (len(missing), missing))
if bad:
    sys.exit('對不上，沒有寫檔：\n  ' + '\n  '.join(bad))

spec['hotspots'] = out
json.dump(spec, open(SPEC, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
print('ds:%04X 起 %d 筆熱區 → screen-settings.json（視窗原點格 (%d,%d)，'
      '%d 個框全數對上）' % (LIST, len(out), ox, oy, len(used)))
for h in out:
    print('  %-14s (%3d,%3d) %3dx%2d%s  %s'
          % (h['action'], h['x'], h['y'], h['w'], h['h'],
             ' 不可選' if h.get('disabled') else '     ', h['note']))
