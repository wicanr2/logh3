"""掃出所有命中測試呼叫點，解出每一處用的是哪一份熱區清單。

命中測試是 `call far 37A9:01F9`（位元組 `9A F9 01 A9 37`）；呼叫端把清單
的 far 指標 push 進去，位移寫死在 `1E BB <位移> 53`（push ds; mov bx,list;
push bx）。所以「哪個畫面能點哪些東西」是**編譯期就定死的**，
不是執行期從某個「目前視窗」變數讀出來。

用法：tools/py.sh tools/scripts/hittest_scan.py <dump> [輸出.json]
"""
import json, sys, pathlib

SIG = bytes.fromhex('9AF901A937')
DS = 0x54B3

mem = pathlib.Path(sys.argv[1]).read_bytes()
lists = json.loads(pathlib.Path('docs/internals/data/hotspot-lists.json').read_text('utf-8'))
label = {}
for e in lists['lists']:
    label[e['ds_offset']] = '、'.join(str(x) for x in (e.get('labels') or [])) or '（無標籤）'

def find_list(pos):
    """往前 40 bytes 找 `1E BB xx xx 53`（push ds; mov bx,imm; push bx）。

    參數順序是「清單先 push、視窗後 push」（登錄常式的 `[bp+0A]` 是清單、
    `[bp+06]` 是視窗）。視窗本身也可能是靜態結構、用同一個寫法 push，
    所以抓到兩個時清單是**位址較小**的那一個，只抓到一個時它就是清單。"""
    lo = max(0, pos - 40)
    win = mem[lo:pos]
    hits = [i for i in range(len(win) - 4)
            if win[i] == 0x1E and win[i+1] == 0xBB and win[i+4] == 0x53]
    if not hits:
        return None
    i = hits[0] if len(hits) > 1 else hits[-1]
    return win[i+2] | win[i+3] << 8

rows, pos = [], 0
while True:
    pos = mem.find(SIG, pos)
    if pos < 0:
        break
    off = find_list(pos)
    rows.append((pos, off))
    pos += 1

out = []
for pos, off in rows:
    if off is None:
        out.append(f'| `{pos:05X}` | （前 40 bytes 內沒有 push ds/bx）| |')
    else:
        out.append(f'| `{pos:05X}` | `{off:04X}` | {label.get(off, "（清單表裡沒有這個位移）")} |')
print('| 命中測試呼叫點 | 清單 | 標籤 |')
print('|---|---|---|')
print('\n'.join(out))
print(f'\n共 {len(rows)} 處，解出清單 {sum(1 for _, o in rows if o is not None)} 處')

if len(sys.argv) > 2:
    pathlib.Path(sys.argv[2]).write_text(json.dumps({
        'note': '每一處命中測試呼叫（call far 37A9:01F9）與它寫死的熱區清單位移；'
                '位址是那一次執行的線性位址，換一次執行會變，穩定的是「幾處」與'
                '「哪一份清單」。清單位移是資料段內的位移，穩定。',
        'signature': '9A F9 01 A9 37',
        'count': len(rows),
        'sites': [{'linear': '%05X' % pos,
                   'list': None if off is None else '%04X' % off,
                   'labels': label.get(off)} for pos, off in rows],
    }, ensure_ascii=False, indent=1) + '\n', 'utf-8')
    print('寫出', sys.argv[2])
