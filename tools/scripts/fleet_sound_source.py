"""凍結第78輪SVF來源、原始指令與兩陣營音效表。"""
import hashlib
import json
import struct
from pathlib import Path


def digest(p):
    b = p.read_bytes()
    return {"sha256": hashlib.sha256(b).hexdigest(), "bytes": len(b)}


root = Path('/src') if Path('/src/docs').is_dir() else Path('/work')
original = Path('/orig/GIN3SP.VOC')
b = original.read_bytes()
assert digest(original)['sha256'] == 'b36a1b43f3199e776f70302671a6bc37698dd5255113995e7a7f55f78f699a9f'
assert b[:16] == b'MicroWaveSVFv1.0' and struct.unpack_from('<H', b, 16)[0] == 28
ram_path = root / 'workplace/oracle/normal79-r61-a/end-mem.bin'
ram = ram_path.read_bytes()
assert digest(ram_path)['sha256'] == '6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4'
samples = []
at = 18
for i in range(28):
    n, tc, command = struct.unpack_from('<HBB', b, at)
    assert n >= 2 and at + 4 + n <= len(b) and tc == 166 and command == 20
    payload = b[at + 4:at + 4 + n]
    samples.append({'id': i, 'header_offset': at, 'payload_offset': at + 4,
                    'length': n, 'time_constant': tc, 'command': command,
                    'payload_sha256': hashlib.sha256(payload).hexdigest(),
                    'played_samples': n - 1,
                    'played_pcm_sha256': hashlib.sha256(payload[:-1]).hexdigest(),
                    'duration_ns': (n - 1) * 90000})
    at += 4 + n
assert at == len(b)
paths = ['tools/scripts/fleet_sound_ida.py', 'tools/scripts/fleet_sound_source.py',
         'workplace/oracle/normal79-r61-a/end-mem.bin',
         'workplace/ida/scene48-r69-1DD6.com', 'workplace/ida/scene48-r69-1DD6.json',
         'docs/internals/data/fleet-scene.json']
instructions = []
for suffix in ('3B74', '2376', '2376-init'):
    prefix = 'workplace/ida/combat79-r78-sound-' + suffix
    paths += [prefix + '.com', prefix + '.json']
    j = json.loads((root / (prefix + '.json')).read_text())
    assert j['uid'] == 1000 and j['kernel_version'] == '9.4'
    assert j['input_sha256'] == digest(root / (prefix + '.com'))['sha256']
    for r in j['instructions']:
        seg, off = [int(v, 16) for v in r['original_location'].split(':')]
        raw = bytes.fromhex(r['original_bytes_hex'])
        assert ram[seg * 16 + off:seg * 16 + off + len(raw)] == raw
    instructions.extend(j['instructions'])
tables = {key: ram[0x54b30 + off:0x54b30 + off + 66].hex()
          for key, off in (('attack_hex', 0x43e0), ('defence_hex', 0x439e))}
locs = {
    '3B74:031A': '讀busy至0，輸入低byte取餘28',
    '3B74:030E': '配置64KiB樣本緩衝', '3B74:0314': '釋放樣本緩衝',
    '2376:1585': '回DS:9208 busy byte',
    '2376:1677': '檔案seek 18並逐段跳過payload，四byte標頭',
    '2376:16B3': 'word payload長度，時間常數及DSP命令分欄',
    '2376:1592': 'DSP時間常數，樣本長度參數N-2及命令',
    '2376:1717': '讀16-byte標頭，前12bytes簽章',
    '2376:1749': '讀兩byte樣本數'}
semantic = []
for loc, meaning in locs.items():
    rows = [r for r in instructions if r['original_location'] == loc]
    assert rows, loc
    semantic.append({**rows[0], 'semantic': meaning, 'grade': 'proven',
                     'warning': '', 'source': '研究415；原始bytes及本版SVF完整索引核對'})
data = {'schema': 'logh3.fleet-sound-source.v1', 'date': '2026-10-04',
        'dosgolem_head': 'f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c',
        'source': {'name': 'GIN3SP.VOC', **digest(original)},
        'format': 'MicroWaveSVFv1.0', 'samples': samples, 'tables': tables,
        'instructions': instructions, 'semantic_index': semantic,
        'files': {p: digest(root / p) for p in paths},
        'platform_reference': 'https://www.phatcode.net/res/243/files/sbhwpg.pdf',
        'timing_grade': 'hardware-spec approximation',
        'limits': ['未重證driver／ISR／DMA', '正常原版PCM與人耳尚未驗', '兩表只核本版bytes']}
output = root / 'docs/internals/data/fleet-sound-source.json'
assert not output.exists() and output.parent.stat().st_uid == 1000
output.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
spec = root / 'docs/internals/data/fleet-sound-spec.json'
assert not spec.exists()
spec.write_text(json.dumps({'schema': 'logh3.fleet-sound-spec.v1', **tables}, indent=2) + '\n')
print('28段樣本／', len(instructions), '原指令／', len(semantic), '已證實定位')
print('source', digest(output)['sha256'], 'spec', digest(spec)['sha256'])
