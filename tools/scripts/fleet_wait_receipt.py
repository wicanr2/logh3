"""凍結原版等待輸入雙跑及owner原指令，不讀remake。"""
import hashlib
import json
from pathlib import Path

root = Path('/src')
a = root / 'workplace/oracle/wait48-r72-v2-a'
b = root / 'workplace/oracle/wait48-r72-v2-b'
sources = []


def source(p):
    data = p.read_bytes()
    sources.append({'path': str(p.relative_to(root)), 'bytes': len(data),
                    'sha256': hashlib.sha256(data).hexdigest()})
    return data


x = json.loads(source(a / 'report.json'))
y = json.loads(source(b / 'report.json'))
x.pop('command')
y.pop('command')
assert x == y and len(x['cases']) == 11
for p in sorted(a.glob('*.bin')):
    assert source(p) == source(b / p.name), p
ram = (a / 'case-00-entry-mem.bin').read_bytes()
ida = json.loads(source(root / 'workplace/ida/projection48-r72-3765.json'))
for row in ida['instructions']:
    addr = int(row['runtime_linear_hex'], 16)
    raw = bytes.fromhex(row['original_bytes_hex'])
    assert ram[addr:addr+len(raw)] == raw, row['original_location']
for p in [root / 'tools/fleet_wait_oracle.sh', root / 'tools/scripts/fleet_wait_native.go',
          root / 'tools/scripts/fleet_wait_receipt.py',
          root / 'workplace/oracle/wait48-r72-a/abort.json']:
    source(p)
out = {'schema': 'logh3.fleet-wait.v1', 'date': '2026-10-04', 'native': x,
       'sources': sources, 'ida': ida,
       'semantic_index': [
           {'original_location': '3765:0076', 'semantic': '呼叫端duration寫到DS:8EFE計數器。',
            'grade': 'proven', 'source': '原指令及duration0／1／90固定案例'},
           {'original_location': '3765:0093', 'semantic': '計數器為0、事件碼1或2才結束，並清計數器，AX回1。',
            'grade': 'proven', 'source': '原指令及11案双跑'},
           {'original_location': '3765:00BF', 'semantic': '先取DS:F248佇列的一筆12-byte事件；空佇列則event word清0。',
            'grade': 'proven', 'source': '原指令及演出中左右鍵按下／放開後仍可縮短等待'},
       ],
       'limits': ['步數只作dosgolem重播定位，不換算成實機wall-clock。',
                  '空白與ESC硬體掃描碼不縮短本批等待；不宣稱所有鍵盤輸入組合已驗。',
                  '延後release在返回後未送出，report完整記錄實際送出的輸入。',
                  '返回戰場後的放開事件生命週期仍未知。',
                  '完整Game場景、起始tick及正常GUI與保存仍依056／092。']}
data = (json.dumps(out, ensure_ascii=False, indent=2)+'\n').encode()
dest = root / 'docs/internals/data/fleet-wait.json'
if dest.exists(): assert dest.read_bytes() == data, '拒絕改寫既有收據'
else: dest.write_bytes(data)
print(len(sources), 'sources', len(ida['instructions']), 'instructions',
      len(data), 'bytes', hashlib.sha256(data).hexdigest())
